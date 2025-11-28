/*
Copyright 2014 Workiva, LLC

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
Package queue provides generic queue implementations including a standard queue
and a priority queue. These queues are thread-safe and never block on send,
growing as much as necessary.

Both implementations return errors instead of panicking when operations are
attempted on disposed queues. The priority queue maintains items in priority
order using a heap.

Example usage:

	q := queue.New[string](10)
	q.Put("hello", "world")
	items, _ := q.Get(2)
	// items = []string{"hello", "world"}
*/
package queue

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type waiters []*sema

func (w *waiters) get() *sema {
	if len(*w) == 0 {
		return nil
	}

	sema := (*w)[0]
	copy((*w)[0:], (*w)[1:])
	(*w)[len(*w)-1] = nil
	*w = (*w)[:len(*w)-1]
	return sema
}

func (w *waiters) put(sema *sema) {
	*w = append(*w, sema)
}

func (w *waiters) remove(sema *sema) {
	if len(*w) == 0 {
		return
	}
	ws := *w
	newWs := make(waiters, 0, len(*w))
	for i := range ws {
		if ws[i] != sema {
			newWs = append(newWs, ws[i])
		}
	}
	*w = newWs
}

type items[T any] []T

func (items *items[T]) get(number int64) []T {
	returnItems := make([]T, 0, number)
	index := int64(0)
	for i := int64(0); i < number; i++ {
		if i >= int64(len(*items)) {
			break
		}

		returnItems = append(returnItems, (*items)[i])
		var zero T
		(*items)[i] = zero
		index++
	}

	*items = (*items)[index:]
	return returnItems
}

func (items *items[T]) peek() (T, bool) {
	length := len(*items)

	if length == 0 {
		var zero T
		return zero, false
	}

	return (*items)[0], true
}

func (items *items[T]) getUntil(checker func(item T) bool) []T {
	length := len(*items)

	if len(*items) == 0 {
		return []T{}
	}

	returnItems := make([]T, 0, length)
	index := -1
	for i, item := range *items {
		if !checker(item) {
			break
		}

		returnItems = append(returnItems, item)
		index = i
		var zero T
		(*items)[i] = zero // prevent memory leak
	}

	*items = (*items)[index+1:]
	return returnItems
}

type sema struct {
	ready    chan bool
	response *sync.WaitGroup
}

func newSema() *sema {
	return &sema{
		ready:    make(chan bool, 1),
		response: &sync.WaitGroup{},
	}
}

// Queue is a generic thread-safe queue that can hold items of any type T.
// It grows unboundedly and never blocks on Put operations.
type Queue[T any] struct {
	waiters  waiters
	items    items[T]
	lock     sync.Mutex
	disposed bool
}

// New creates a new Queue with the given initial capacity hint.
// The hint is used to pre-allocate the underlying storage for better performance.
func New[T any](hint int64) *Queue[T] {
	return &Queue[T]{
		items: make([]T, 0, hint),
	}
}

// Put adds the specified items to the queue.
// Returns ErrDisposed if the queue has been disposed.
func (q *Queue[T]) Put(items ...T) error {
	if len(items) == 0 {
		return nil
	}

	q.lock.Lock()

	if q.disposed {
		q.lock.Unlock()
		return ErrDisposed
	}

	q.items = append(q.items, items...)
	for {
		sema := q.waiters.get()
		if sema == nil {
			break
		}
		sema.response.Add(1)
		select {
		case sema.ready <- true:
			sema.response.Wait()
		default:
			// This semaphore timed out.
		}
		if len(q.items) == 0 {
			break
		}
	}

	q.lock.Unlock()
	return nil
}

// Get retrieves items from the queue. If there are some items in the
// queue, get will return a number UP TO the number passed in as a
// parameter. If no items are in the queue, this method will pause
// until items are added to the queue.
func (q *Queue[T]) Get(number int64) ([]T, error) {
	return q.Poll(number, 0)
}

// Poll retrieves items from the queue. If there are some items in the queue,
// Poll will return a number UP TO the number passed in as a parameter. If no
// items are in the queue, this method will pause until items are added to the
// queue or the provided timeout is reached. A non-positive timeout will block
// until items are added. If a timeout occurs, ErrTimeout is returned.
func (q *Queue[T]) Poll(number int64, timeout time.Duration) ([]T, error) {
	if number < 1 {
		return []T{}, nil
	}

	q.lock.Lock()

	if q.disposed {
		q.lock.Unlock()
		return nil, ErrDisposed
	}

	var items []T

	if len(q.items) == 0 {
		sema := newSema()
		q.waiters.put(sema)
		q.lock.Unlock()

		var timeoutC <-chan time.Time
		if timeout > 0 {
			timeoutC = time.After(timeout)
		}
		select {
		case <-sema.ready:
			if q.disposed {
				return nil, ErrDisposed
			}
			items = q.items.get(number)
			sema.response.Done()
			return items, nil
		case <-timeoutC:
			select {
			case sema.ready <- true:
				q.lock.Lock()
				q.waiters.remove(sema)
				q.lock.Unlock()
			default:
				sema.response.Done()
			}
			return nil, ErrTimeout
		}
	}

	items = q.items.get(number)
	q.lock.Unlock()
	return items, nil
}

// Peek returns the first item in the queue by value without modifying the queue.
// Returns ErrEmptyQueue if the queue is empty, ErrDisposed if disposed.
func (q *Queue[T]) Peek() (T, error) {
	q.lock.Lock()
	defer q.lock.Unlock()

	if q.disposed {
		var zero T
		return zero, ErrDisposed
	}

	peekItem, ok := q.items.peek()
	if !ok {
		var zero T
		return zero, ErrEmptyQueue
	}

	return peekItem, nil
}

// TakeUntil takes a function and returns a list of items that
// match the checker until the checker returns false. This does not
// wait if there are no items in the queue.
func (q *Queue[T]) TakeUntil(checker func(item T) bool) ([]T, error) {
	if checker == nil {
		return nil, nil
	}

	q.lock.Lock()

	if q.disposed {
		q.lock.Unlock()
		return nil, ErrDisposed
	}

	result := q.items.getUntil(checker)
	q.lock.Unlock()
	return result, nil
}

// Empty returns a bool indicating if the queue is empty.
func (q *Queue[T]) Empty() bool {
	q.lock.Lock()
	defer q.lock.Unlock()

	return len(q.items) == 0
}

// Len returns the number of items in this queue.
func (q *Queue[T]) Len() int64 {
	q.lock.Lock()
	defer q.lock.Unlock()

	return int64(len(q.items))
}

// Disposed returns a bool indicating if this queue has been disposed.
func (q *Queue[T]) Disposed() bool {
	q.lock.Lock()
	defer q.lock.Unlock()

	return q.disposed
}

// Dispose will dispose of this queue and returns the items disposed.
// Any subsequent calls to Get or Put will return an error.
func (q *Queue[T]) Dispose() []T {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.disposed = true
	for _, waiter := range q.waiters {
		waiter.response.Add(1)
		select {
		case waiter.ready <- true:
		default:
		}
	}

	disposedItems := q.items

	q.items = nil
	q.waiters = nil

	return disposedItems
}

// ExecuteInParallel will (in parallel) call the provided function
// with each item in the queue until the queue is exhausted. When the queue
// is exhausted execution is complete and all goroutines will be killed.
// This means that the queue will be disposed so cannot be used again.
func ExecuteInParallel[T any](q *Queue[T], fn func(T)) {
	if q == nil {
		return
	}

	q.lock.Lock()
	todo, done := uint64(len(q.items)), int64(-1)
	if todo == 0 {
		q.lock.Unlock()
		return
	}

	numCPU := 1
	if runtime.NumCPU() > 1 {
		numCPU = runtime.NumCPU() - 1
	}

	var wg sync.WaitGroup
	wg.Add(numCPU)
	items := q.items

	for i := 0; i < numCPU; i++ {
		go func() {
			for {
				index := atomic.AddInt64(&done, 1)
				if index >= int64(todo) {
					wg.Done()
					break
				}

				fn(items[index])
				var zero T
				items[index] = zero
			}
		}()
	}
	wg.Wait()
	q.lock.Unlock()
	q.Dispose()
}
