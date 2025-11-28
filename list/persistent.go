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
Package list provides generic list implementations. Currently, this includes a
persistent, immutable linked list.

The PersistentList is an immutable, persistent linked list. All write operations
yield a new structure that preserves and reuses previous versions.

Example usage:

	list := list.Empty[int]()
	list = list.Add(1).Add(2).Add(3)

	head, _ := list.Head() // 3
	tail, _ := list.Tail() // list containing [2, 1]

	// Map over elements
	doubled := list.Map(func(x int) int { return x * 2 })
*/
package list

import "errors"

var (
	// ErrEmptyList is returned when an invalid operation is performed on an
	// empty list.
	ErrEmptyList = errors.New("empty list")
)

// PersistentList is a generic immutable, persistent linked list.
type PersistentList[T any] interface {
	// Head returns the head of the list. The bool will be false if the list is empty.
	Head() (T, bool)

	// Tail returns the tail of the list. The bool will be false if the list is empty.
	Tail() (PersistentList[T], bool)

	// IsEmpty indicates if the list is empty.
	IsEmpty() bool

	// Length returns the number of items in the list.
	Length() uint

	// Add will add the item to the list, returning the new list.
	Add(head T) PersistentList[T]

	// Insert will insert the item at the given position, returning the new
	// list or an error if the position is invalid.
	Insert(val T, pos uint) (PersistentList[T], error)

	// Get returns the item at the given position or false if the position is invalid.
	Get(pos uint) (T, bool)

	// Remove will remove the item at the given position, returning the new
	// list or an error if the position is invalid.
	Remove(pos uint) (PersistentList[T], error)

	// Find applies the predicate function to the list and returns the first
	// item which matches.
	Find(predicate func(T) bool) (T, bool)

	// FindIndex applies the predicate function to the list and returns the
	// index of the first item which matches or -1 if there is no match.
	FindIndex(predicate func(T) bool) int

	// Map applies the function to each entry in the list and returns the
	// resulting slice.
	Map(fn func(T) T) []T

	// ForEach applies the function to each entry in the list.
	ForEach(fn func(T))

	// Filter returns a new list containing only elements that satisfy the predicate.
	Filter(predicate func(T) bool) PersistentList[T]

	// Reduce applies a reducer function to all elements.
	Reduce(fn func(acc, item T) T, initial T) T

	// ToSlice returns the list as a slice.
	ToSlice() []T

	// Reverse returns a new list with elements in reverse order.
	Reverse() PersistentList[T]
}

// Empty returns an empty PersistentList for the given type.
func Empty[T any]() PersistentList[T] {
	return &emptyList[T]{}
}

// FromSlice creates a PersistentList from a slice.
// Items are added in order, so the last item in the slice becomes the head.
func FromSlice[T any](items []T) PersistentList[T] {
	var list PersistentList[T] = Empty[T]()
	for _, item := range items {
		list = list.Add(item)
	}
	return list
}

// FromSliceReversed creates a PersistentList from a slice.
// Items are added in reverse order, so the first item in the slice becomes the head.
func FromSliceReversed[T any](items []T) PersistentList[T] {
	var list PersistentList[T] = Empty[T]()
	for i := len(items) - 1; i >= 0; i-- {
		list = list.Add(items[i])
	}
	return list
}

type emptyList[T any] struct{}

func (e *emptyList[T]) Head() (T, bool) {
	var zero T
	return zero, false
}

func (e *emptyList[T]) Tail() (PersistentList[T], bool) {
	return nil, false
}

func (e *emptyList[T]) IsEmpty() bool {
	return true
}

func (e *emptyList[T]) Length() uint {
	return 0
}

func (e *emptyList[T]) Add(head T) PersistentList[T] {
	return &list[T]{head: head, tail: e}
}

func (e *emptyList[T]) Insert(val T, pos uint) (PersistentList[T], error) {
	if pos == 0 {
		return e.Add(val), nil
	}
	return nil, ErrEmptyList
}

func (e *emptyList[T]) Get(pos uint) (T, bool) {
	var zero T
	return zero, false
}

func (e *emptyList[T]) Remove(pos uint) (PersistentList[T], error) {
	return nil, ErrEmptyList
}

func (e *emptyList[T]) Find(predicate func(T) bool) (T, bool) {
	var zero T
	return zero, false
}

func (e *emptyList[T]) FindIndex(predicate func(T) bool) int {
	return -1
}

func (e *emptyList[T]) Map(fn func(T) T) []T {
	return nil
}

func (e *emptyList[T]) ForEach(fn func(T)) {}

func (e *emptyList[T]) Filter(predicate func(T) bool) PersistentList[T] {
	return e
}

func (e *emptyList[T]) Reduce(fn func(acc, item T) T, initial T) T {
	return initial
}

func (e *emptyList[T]) ToSlice() []T {
	return nil
}

func (e *emptyList[T]) Reverse() PersistentList[T] {
	return e
}

type list[T any] struct {
	head T
	tail PersistentList[T]
}

func (l *list[T]) Head() (T, bool) {
	return l.head, true
}

func (l *list[T]) Tail() (PersistentList[T], bool) {
	return l.tail, true
}

func (l *list[T]) IsEmpty() bool {
	return false
}

func (l *list[T]) Length() uint {
	curr := l
	length := uint(0)
	for {
		length++
		tail, _ := curr.Tail()
		if tail.IsEmpty() {
			return length
		}
		curr = tail.(*list[T])
	}
}

func (l *list[T]) Add(head T) PersistentList[T] {
	return &list[T]{head: head, tail: l}
}

func (l *list[T]) Insert(val T, pos uint) (PersistentList[T], error) {
	if pos == 0 {
		return l.Add(val), nil
	}
	nl, err := l.tail.Insert(val, pos-1)
	if err != nil {
		return nil, err
	}
	return nl.Add(l.head), nil
}

func (l *list[T]) Get(pos uint) (T, bool) {
	if pos == 0 {
		return l.head, true
	}
	return l.tail.Get(pos - 1)
}

func (l *list[T]) Remove(pos uint) (PersistentList[T], error) {
	if pos == 0 {
		nl, _ := l.Tail()
		return nl, nil
	}

	nl, err := l.tail.Remove(pos - 1)
	if err != nil {
		return nil, err
	}
	return &list[T]{head: l.head, tail: nl}, nil
}

func (l *list[T]) Find(predicate func(T) bool) (T, bool) {
	if predicate(l.head) {
		return l.head, true
	}
	return l.tail.Find(predicate)
}

func (l *list[T]) FindIndex(predicate func(T) bool) int {
	curr := l
	idx := 0
	for {
		if predicate(curr.head) {
			return idx
		}
		tail, _ := curr.Tail()
		if tail.IsEmpty() {
			return -1
		}
		curr = tail.(*list[T])
		idx++
	}
}

func (l *list[T]) Map(fn func(T) T) []T {
	return append(l.tail.Map(fn), fn(l.head))
}

func (l *list[T]) ForEach(fn func(T)) {
	fn(l.head)
	l.tail.ForEach(fn)
}

func (l *list[T]) Filter(predicate func(T) bool) PersistentList[T] {
	filtered := l.tail.Filter(predicate)
	if predicate(l.head) {
		return filtered.Add(l.head)
	}
	return filtered
}

func (l *list[T]) Reduce(fn func(acc, item T) T, initial T) T {
	acc := fn(initial, l.head)
	return l.tail.Reduce(fn, acc)
}

func (l *list[T]) ToSlice() []T {
	result := make([]T, 0, l.Length())
	l.ForEach(func(item T) {
		result = append(result, item)
	})
	return result
}

func (l *list[T]) Reverse() PersistentList[T] {
	var result PersistentList[T] = Empty[T]()
	l.ForEach(func(item T) {
		result = result.Add(item)
	})
	return result
}
