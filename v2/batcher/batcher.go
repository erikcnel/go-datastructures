/*
Copyright 2015 Workiva, LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*
Package batcher provides a generic API for accumulating items into batches
for processing.

Batch readiness can be determined by:
  - Maximum number of bytes per batch
  - Maximum number of items per batch
  - Maximum amount of time waiting for a batch

Example usage:

	b, _ := batcher.New[string](batcher.Config[string]{
		MaxTime:  time.Second,
		MaxItems: 100,
	})

	go func() {
		for {
			batch, _ := b.Get()
			process(batch)
		}
	}()

	b.Put("item1")
	b.Put("item2")
*/
package batcher

import (
	"errors"
	"time"
)

// mutex is a simple TryLock-capable mutex using channels
type mutex struct {
	lock chan struct{}
}

func newMutex() *mutex {
	return &mutex{lock: make(chan struct{}, 1)}
}

func (m *mutex) Lock() {
	m.lock <- struct{}{}
}

func (m *mutex) Unlock() {
	<-m.lock
}

func (m *mutex) TryLock() bool {
	select {
	case m.lock <- struct{}{}:
		return true
	default:
		return false
	}
}

// ErrDisposed is returned when an operation is attempted on a disposed Batcher.
var ErrDisposed = errors.New("batcher: disposed")

// CalculateBytes evaluates the number of bytes in an item added to a Batcher.
type CalculateBytes[T any] func(T) uint

// Config holds configuration options for creating a Batcher.
type Config[T any] struct {
	// MaxTime is the maximum time to wait before returning a batch.
	// A zero value means no time limit.
	MaxTime time.Duration

	// MaxItems is the maximum number of items per batch.
	// A zero value means no item limit.
	MaxItems uint

	// MaxBytes is the maximum number of bytes per batch.
	// A zero value means no byte limit.
	// If MaxBytes > 0, CalculateBytes must be provided.
	MaxBytes uint

	// CalculateBytes returns the byte size of an item.
	// Required if MaxBytes > 0.
	CalculateBytes CalculateBytes[T]

	// QueueLen is the buffer size for completed batches.
	// Defaults to 10 if not specified.
	QueueLen uint
}

// Batcher provides an API for accumulating items into batches for processing.
type Batcher[T any] struct {
	maxTime        time.Duration
	maxItems       uint
	maxBytes       uint
	calculateBytes CalculateBytes[T]
	disposed       bool
	items          []T
	batchChan      chan []T
	availableBytes uint
	lock           *mutex
}

// New creates a new Batcher with the given configuration.
func New[T any](config Config[T]) (*Batcher[T], error) {
	if config.MaxBytes > 0 && config.CalculateBytes == nil {
		return nil, errors.New("batcher: must provide CalculateBytes function when MaxBytes is set")
	}

	queueLen := config.QueueLen
	if queueLen == 0 {
		queueLen = 10
	}

	maxItems := config.MaxItems
	if maxItems == 0 {
		maxItems = 100
	}

	return &Batcher[T]{
		maxTime:        config.MaxTime,
		maxItems:       config.MaxItems,
		maxBytes:       config.MaxBytes,
		calculateBytes: config.CalculateBytes,
		items:          make([]T, 0, maxItems),
		batchChan:      make(chan []T, queueLen),
		lock:           newMutex(),
	}, nil
}

// Put adds an item to the batcher.
// Returns ErrDisposed if the batcher has been disposed.
func (b *Batcher[T]) Put(item T) error {
	b.lock.Lock()
	if b.disposed {
		b.lock.Unlock()
		return ErrDisposed
	}

	b.items = append(b.items, item)
	if b.calculateBytes != nil {
		b.availableBytes += b.calculateBytes(item)
	}
	if b.ready() {
		b.flush()
	}

	b.lock.Unlock()
	return nil
}

// Get retrieves a batch from the batcher. This call will block until
// one of the conditions for a "complete" batch is reached.
// Returns ErrDisposed if the batcher is disposed and no more batches are available.
func (b *Batcher[T]) Get() ([]T, error) {
	var timeout <-chan time.Time
	if b.maxTime > 0 {
		timeout = time.After(b.maxTime)
	}

	select {
	case items, ok := <-b.batchChan:
		if !ok {
			return nil, ErrDisposed
		}
		return items, nil
	case <-timeout:
		for {
			if b.lock.TryLock() {
				select {
				case items, ok := <-b.batchChan:
					b.lock.Unlock()
					if !ok {
						return nil, ErrDisposed
					}
					return items, nil
				default:
				}

				items := b.items
				b.items = make([]T, 0, b.maxItems)
				b.availableBytes = 0
				b.lock.Unlock()
				return items, nil
			} else {
				select {
				case items, ok := <-b.batchChan:
					if !ok {
						return nil, ErrDisposed
					}
					return items, nil
				default:
				}
			}
		}
	}
}

// Flush forcibly completes the batch currently being built.
// Returns ErrDisposed if the batcher has been disposed.
func (b *Batcher[T]) Flush() error {
	b.lock.Lock()
	if b.disposed {
		b.lock.Unlock()
		return ErrDisposed
	}
	b.flush()
	b.lock.Unlock()
	return nil
}

// Dispose will dispose of the batcher. Any calls to Put or Flush
// will return ErrDisposed. Calls to Get will return an error if
// there are no more ready batches.
func (b *Batcher[T]) Dispose() {
	for {
		if b.lock.TryLock() {
			if b.disposed {
				b.lock.Unlock()
				return
			}

			b.disposed = true
			b.items = nil
			b.drainBatchChan()
			close(b.batchChan)
			b.lock.Unlock()
			return
		} else {
			b.drainBatchChan()
		}
	}
}

// IsDisposed returns true if the batcher has been disposed.
func (b *Batcher[T]) IsDisposed() bool {
	b.lock.Lock()
	disposed := b.disposed
	b.lock.Unlock()
	return disposed
}

func (b *Batcher[T]) flush() {
	b.batchChan <- b.items
	b.items = make([]T, 0, b.maxItems)
	b.availableBytes = 0
}

func (b *Batcher[T]) ready() bool {
	if b.maxItems != 0 && uint(len(b.items)) >= b.maxItems {
		return true
	}
	if b.maxBytes != 0 && b.availableBytes >= b.maxBytes {
		return true
	}
	return false
}

func (b *Batcher[T]) drainBatchChan() {
	for {
		select {
		case <-b.batchChan:
		default:
			return
		}
	}
}
