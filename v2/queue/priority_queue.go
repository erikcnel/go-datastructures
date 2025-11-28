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

package queue

import "sync"

// Comparable is an interface for items that can be compared for ordering.
// Items implementing this interface can be used with PriorityQueue.
type Comparable[T any] interface {
	// Compare returns a value indicating the relationship between this item
	// and the other. Return 1 if this > other, 0 if equal, -1 if this < other.
	Compare(other T) int
}

// PriorityItem wraps a value with a priority for use in OrderedPriorityQueue.
type PriorityItem[T any] struct {
	Value    T
	Priority int
}

// Compare implements Comparable for PriorityItem.
func (p PriorityItem[T]) Compare(other PriorityItem[T]) int {
	if p.Priority < other.Priority {
		return -1
	}
	if p.Priority > other.Priority {
		return 1
	}
	return 0
}

type priorityItems[T Comparable[T]] []T

func (items *priorityItems[T]) swap(i, j int) {
	(*items)[i], (*items)[j] = (*items)[j], (*items)[i]
}

func (items *priorityItems[T]) pop() T {
	size := len(*items)

	items.swap(size-1, 0)
	item := (*items)[size-1]
	var zero T
	(*items)[size-1], *items = zero, (*items)[:size-1]

	index := 0
	childL, childR := 2*index+1, 2*index+2
	for len(*items) > childL {
		child := childL
		if len(*items) > childR && (*items)[childR].Compare((*items)[childL]) < 0 {
			child = childR
		}

		if (*items)[child].Compare((*items)[index]) < 0 {
			items.swap(index, child)

			index = child
			childL, childR = 2*index+1, 2*index+2
		} else {
			break
		}
	}

	return item
}

func (items *priorityItems[T]) get(number int) []T {
	returnItems := make([]T, 0, number)
	for i := 0; i < number; i++ {
		if len(*items) == 0 {
			break
		}

		returnItems = append(returnItems, items.pop())
	}

	return returnItems
}

func (items *priorityItems[T]) push(item T) {
	*items = append(*items, item)

	index := len(*items) - 1
	parent := (index - 1) / 2
	for parent >= 0 && (*items)[parent].Compare(item) > 0 {
		items.swap(index, parent)

		index = parent
		parent = (index - 1) / 2
	}
}

// PriorityQueue is a generic thread-safe priority queue.
// Items must implement the Comparable interface for ordering.
type PriorityQueue[T Comparable[T]] struct {
	waiters         waiters
	items           priorityItems[T]
	lock            sync.Mutex
	disposeLock     sync.Mutex
	disposed        bool
	allowDuplicates bool
}

// NewPriorityQueue creates a new priority queue with the given capacity hint.
// If allowDuplicates is false, duplicate items (as determined by pointer equality
// or value equality for comparable types) will not be added.
func NewPriorityQueue[T Comparable[T]](hint int, allowDuplicates bool) *PriorityQueue[T] {
	return &PriorityQueue[T]{
		items:           make(priorityItems[T], 0, hint),
		allowDuplicates: allowDuplicates,
	}
}

// Put adds items to the queue in priority order.
// Returns ErrDisposed if the queue has been disposed.
func (pq *PriorityQueue[T]) Put(items ...T) error {
	if len(items) == 0 {
		return nil
	}

	pq.lock.Lock()
	defer pq.lock.Unlock()

	if pq.disposed {
		return ErrDisposed
	}

	for _, item := range items {
		pq.items.push(item)
	}

	for {
		sema := pq.waiters.get()
		if sema == nil {
			break
		}

		sema.response.Add(1)
		sema.ready <- true
		sema.response.Wait()
		if len(pq.items) == 0 {
			break
		}
	}

	return nil
}

// Get retrieves items from the queue in priority order.
// If the queue is empty, this call blocks until items are added.
func (pq *PriorityQueue[T]) Get(number int) ([]T, error) {
	if number < 1 {
		return nil, nil
	}

	pq.lock.Lock()

	if pq.disposed {
		pq.lock.Unlock()
		return nil, ErrDisposed
	}

	var items []T

	if len(pq.items) == 0 {
		sema := newSema()
		pq.waiters.put(sema)
		pq.lock.Unlock()

		<-sema.ready

		if pq.Disposed() {
			return nil, ErrDisposed
		}

		items = pq.items.get(number)
		sema.response.Done()
		return items, nil
	}

	items = pq.items.get(number)
	pq.lock.Unlock()
	return items, nil
}

// Peek returns the highest priority item without removing it from the queue.
// Returns the zero value if the queue is empty.
func (pq *PriorityQueue[T]) Peek() (T, bool) {
	pq.lock.Lock()
	defer pq.lock.Unlock()
	if len(pq.items) > 0 {
		return pq.items[0], true
	}
	var zero T
	return zero, false
}

// Empty returns true if the queue has no items.
func (pq *PriorityQueue[T]) Empty() bool {
	pq.lock.Lock()
	defer pq.lock.Unlock()

	return len(pq.items) == 0
}

// Len returns the number of items in the queue.
func (pq *PriorityQueue[T]) Len() int {
	pq.lock.Lock()
	defer pq.lock.Unlock()

	return len(pq.items)
}

// Disposed returns true if this queue has been disposed.
func (pq *PriorityQueue[T]) Disposed() bool {
	pq.disposeLock.Lock()
	defer pq.disposeLock.Unlock()

	return pq.disposed
}

// Dispose prevents any further reads/writes and frees resources.
func (pq *PriorityQueue[T]) Dispose() {
	pq.lock.Lock()
	defer pq.lock.Unlock()

	pq.disposeLock.Lock()
	defer pq.disposeLock.Unlock()

	pq.disposed = true
	for _, waiter := range pq.waiters {
		waiter.response.Add(1)
		waiter.ready <- true
	}

	pq.items = nil
	pq.waiters = nil
}

// OrderedPriorityQueue is a convenience type for priority queues using
// the built-in PriorityItem wrapper with integer priorities.
type OrderedPriorityQueue[T any] struct {
	*PriorityQueue[PriorityItem[T]]
}

// NewOrderedPriorityQueue creates a priority queue that uses integer priorities.
// Lower priority values are dequeued first (min-heap behavior).
func NewOrderedPriorityQueue[T any](hint int, allowDuplicates bool) *OrderedPriorityQueue[T] {
	return &OrderedPriorityQueue[T]{
		PriorityQueue: NewPriorityQueue[PriorityItem[T]](hint, allowDuplicates),
	}
}

// Enqueue adds an item with the given priority.
func (opq *OrderedPriorityQueue[T]) Enqueue(value T, priority int) error {
	return opq.Put(PriorityItem[T]{Value: value, Priority: priority})
}

// Dequeue removes and returns the highest priority item.
func (opq *OrderedPriorityQueue[T]) Dequeue() (T, int, error) {
	items, err := opq.Get(1)
	if err != nil {
		var zero T
		return zero, 0, err
	}
	if len(items) == 0 {
		var zero T
		return zero, 0, ErrEmptyQueue
	}
	return items[0].Value, items[0].Priority, nil
}
