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
Package futures provides a generic implementation for broadcasting a result
to multiple listeners.

Unlike channels which choose a listener at random when multiple goroutines
are listening, a Future will broadcast the result to all listeners. The future
also caches the result so any subsequent calls will immediately return the
cached value.

Example usage:

	completer := make(chan string, 1)
	future := futures.New[string](completer, 5*time.Second)

	// Multiple goroutines can wait on the same future
	go func() {
		result, err := future.GetResult()
		// ...
	}()

	// Complete the future
	completer <- "hello"
*/
package futures

import (
	"fmt"
	"sync"
	"time"
)

// Completer is a channel that the future expects to receive a result on.
type Completer[T any] <-chan T

// Future represents an object that can be used to perform asynchronous tasks.
// The constructor of the future will complete it, and listeners will block
// on GetResult until a result is received. This is different from a channel
// in that the future is only completed once, and anyone listening on the
// future will get the result, regardless of the number of listeners.
type Future[T any] struct {
	triggered bool
	item      T
	err       error
	lock      sync.Mutex
	wg        sync.WaitGroup
}

// New creates a new Future that will be completed when a value is received
// on the completer channel or when the timeout is reached.
func New[T any](completer Completer[T], timeout time.Duration) *Future[T] {
	f := &Future[T]{}
	f.wg.Add(1)
	var wg sync.WaitGroup
	wg.Add(1)
	go listenForResult(f, completer, timeout, &wg)
	wg.Wait()
	return f
}

// GetResult will immediately return the result if it exists,
// or wait until the result is ready.
func (f *Future[T]) GetResult() (T, error) {
	f.lock.Lock()
	if f.triggered {
		item, err := f.item, f.err
		f.lock.Unlock()
		return item, err
	}
	f.lock.Unlock()

	f.wg.Wait()
	return f.item, f.err
}

// HasResult returns true if the result is available.
func (f *Future[T]) HasResult() bool {
	f.lock.Lock()
	hasResult := f.triggered
	f.lock.Unlock()
	return hasResult
}

func (f *Future[T]) setItem(item T, err error) {
	f.lock.Lock()
	f.triggered = true
	f.item = item
	f.err = err
	f.lock.Unlock()
	f.wg.Done()
}

func listenForResult[T any](f *Future[T], ch Completer[T], timeout time.Duration, wg *sync.WaitGroup) {
	wg.Done()
	t := time.NewTimer(timeout)
	select {
	case item := <-ch:
		f.setItem(item, nil)
		t.Stop()
	case <-t.C:
		var zero T
		f.setItem(zero, fmt.Errorf("timeout after %f seconds", timeout.Seconds()))
	}
}

// Promise provides a way to complete a Future from the producer side.
// It wraps a Future and provides a Complete method.
type Promise[T any] struct {
	future    *Future[T]
	completer chan T
	once      sync.Once
}

// NewPromise creates a new Promise with the given timeout.
func NewPromise[T any](timeout time.Duration) *Promise[T] {
	completer := make(chan T, 1)
	return &Promise[T]{
		future:    New(completer, timeout),
		completer: completer,
	}
}

// Complete completes the promise with the given value.
// Calling Complete multiple times has no effect after the first call.
func (p *Promise[T]) Complete(value T) {
	p.once.Do(func() {
		p.completer <- value
		close(p.completer)
	})
}

// Future returns the Future associated with this Promise.
func (p *Promise[T]) Future() *Future[T] {
	return p.future
}

// Await is a convenience function that creates a Future from a function
// that returns a value and an error.
func Await[T any](fn func() (T, error), timeout time.Duration) *Future[T] {
	completer := make(chan T, 1)
	f := New(completer, timeout)

	go func() {
		result, err := fn()
		if err != nil {
			var zero T
			f.setItem(zero, err)
		} else {
			completer <- result
		}
		close(completer)
	}()

	return f
}

// All waits for all futures to complete and returns their results.
// If any future returns an error, the first error is returned.
func All[T any](futures ...*Future[T]) ([]T, error) {
	results := make([]T, len(futures))
	for i, f := range futures {
		result, err := f.GetResult()
		if err != nil {
			return nil, err
		}
		results[i] = result
	}
	return results, nil
}

// Race returns the result of the first future to complete.
func Race[T any](futures ...*Future[T]) (T, error) {
	if len(futures) == 0 {
		var zero T
		return zero, fmt.Errorf("no futures provided")
	}

	done := make(chan struct {
		result T
		err    error
	}, 1)

	var once sync.Once
	for _, f := range futures {
		go func(f *Future[T]) {
			result, err := f.GetResult()
			once.Do(func() {
				done <- struct {
					result T
					err    error
				}{result, err}
			})
		}(f)
	}

	r := <-done
	return r.result, r.err
}
