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
Package set provides a generic, thread-safe Set implementation.

The Set uses Go's comparable constraint, allowing it to work with any type
that supports equality comparison (==). This includes all basic types,
structs with comparable fields, and pointers.

Example usage:

	s := set.New[string]("a", "b", "c")
	s.Add("d")
	exists := s.Exists("a") // true

	// Set operations
	s2 := set.New[string]("c", "d", "e")
	union := s.Union(s2)
	intersection := s.Intersection(s2)
*/
package set

import (
	"sync"
)

// Set is a generic thread-safe set implementation.
// T must be a comparable type (supports == and can be used as map key).
type Set[T comparable] struct {
	items     map[T]struct{}
	lock      sync.RWMutex
	flattened []T
}

// New creates a new Set with the given initial items.
func New[T comparable](items ...T) *Set[T] {
	s := &Set[T]{
		items: make(map[T]struct{}, len(items)),
	}
	for _, item := range items {
		s.items[item] = struct{}{}
	}
	return s
}

// NewWithCapacity creates a new Set with the given capacity hint.
func NewWithCapacity[T comparable](capacity int) *Set[T] {
	return &Set[T]{
		items: make(map[T]struct{}, capacity),
	}
}

// Add will add the provided items to the set.
func (s *Set[T]) Add(items ...T) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.flattened = nil
	for _, item := range items {
		s.items[item] = struct{}{}
	}
}

// Remove will remove the given items from the set.
func (s *Set[T]) Remove(items ...T) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.flattened = nil
	for _, item := range items {
		delete(s.items, item)
	}
}

// Exists returns true if the given item exists in the set.
func (s *Set[T]) Exists(item T) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	_, ok := s.items[item]
	return ok
}

// Contains is an alias for Exists.
func (s *Set[T]) Contains(item T) bool {
	return s.Exists(item)
}

// Flatten returns a slice of all items in the set.
// The returned slice is cached and reused until the set is modified.
func (s *Set[T]) Flatten() []T {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.flattened != nil {
		return s.flattened
	}

	s.flattened = make([]T, 0, len(s.items))
	for item := range s.items {
		s.flattened = append(s.flattened, item)
	}
	return s.flattened
}

// ToSlice returns a new slice containing all items in the set.
// Unlike Flatten, this always creates a new slice.
func (s *Set[T]) ToSlice() []T {
	s.lock.RLock()
	defer s.lock.RUnlock()

	result := make([]T, 0, len(s.items))
	for item := range s.items {
		result = append(result, item)
	}
	return result
}

// Len returns the number of items in the set.
func (s *Set[T]) Len() int {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return len(s.items)
}

// Clear will remove all items from the set.
func (s *Set[T]) Clear() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.items = make(map[T]struct{})
	s.flattened = nil
}

// All returns true if all of the supplied items exist in the set.
func (s *Set[T]) All(items ...T) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	for _, item := range items {
		if _, ok := s.items[item]; !ok {
			return false
		}
	}
	return true
}

// Any returns true if any of the supplied items exist in the set.
func (s *Set[T]) Any(items ...T) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	for _, item := range items {
		if _, ok := s.items[item]; ok {
			return true
		}
	}
	return false
}

// IsEmpty returns true if the set has no items.
func (s *Set[T]) IsEmpty() bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return len(s.items) == 0
}

// Union returns a new set containing all items from both sets.
func (s *Set[T]) Union(other *Set[T]) *Set[T] {
	s.lock.RLock()
	other.lock.RLock()
	defer s.lock.RUnlock()
	defer other.lock.RUnlock()

	result := NewWithCapacity[T](len(s.items) + len(other.items))
	for item := range s.items {
		result.items[item] = struct{}{}
	}
	for item := range other.items {
		result.items[item] = struct{}{}
	}
	return result
}

// Intersection returns a new set containing only items present in both sets.
func (s *Set[T]) Intersection(other *Set[T]) *Set[T] {
	s.lock.RLock()
	other.lock.RLock()
	defer s.lock.RUnlock()
	defer other.lock.RUnlock()

	// Iterate over the smaller set for efficiency
	small, large := s.items, other.items
	if len(s.items) > len(other.items) {
		small, large = other.items, s.items
	}

	result := NewWithCapacity[T](len(small))
	for item := range small {
		if _, ok := large[item]; ok {
			result.items[item] = struct{}{}
		}
	}
	return result
}

// Difference returns a new set containing items in s but not in other.
func (s *Set[T]) Difference(other *Set[T]) *Set[T] {
	s.lock.RLock()
	other.lock.RLock()
	defer s.lock.RUnlock()
	defer other.lock.RUnlock()

	result := NewWithCapacity[T](len(s.items))
	for item := range s.items {
		if _, ok := other.items[item]; !ok {
			result.items[item] = struct{}{}
		}
	}
	return result
}

// SymmetricDifference returns a new set containing items in either set but not both.
func (s *Set[T]) SymmetricDifference(other *Set[T]) *Set[T] {
	s.lock.RLock()
	other.lock.RLock()
	defer s.lock.RUnlock()
	defer other.lock.RUnlock()

	result := NewWithCapacity[T](len(s.items) + len(other.items))
	for item := range s.items {
		if _, ok := other.items[item]; !ok {
			result.items[item] = struct{}{}
		}
	}
	for item := range other.items {
		if _, ok := s.items[item]; !ok {
			result.items[item] = struct{}{}
		}
	}
	return result
}

// IsSubset returns true if all items in s are also in other.
func (s *Set[T]) IsSubset(other *Set[T]) bool {
	s.lock.RLock()
	other.lock.RLock()
	defer s.lock.RUnlock()
	defer other.lock.RUnlock()

	for item := range s.items {
		if _, ok := other.items[item]; !ok {
			return false
		}
	}
	return true
}

// IsSuperset returns true if all items in other are also in s.
func (s *Set[T]) IsSuperset(other *Set[T]) bool {
	return other.IsSubset(s)
}

// Equal returns true if both sets contain the same items.
func (s *Set[T]) Equal(other *Set[T]) bool {
	s.lock.RLock()
	other.lock.RLock()
	defer s.lock.RUnlock()
	defer other.lock.RUnlock()

	if len(s.items) != len(other.items) {
		return false
	}
	for item := range s.items {
		if _, ok := other.items[item]; !ok {
			return false
		}
	}
	return true
}

// Clone creates a shallow copy of the set.
func (s *Set[T]) Clone() *Set[T] {
	s.lock.RLock()
	defer s.lock.RUnlock()

	result := NewWithCapacity[T](len(s.items))
	for item := range s.items {
		result.items[item] = struct{}{}
	}
	return result
}

// ForEach calls the given function for each item in the set.
// The iteration order is not guaranteed.
func (s *Set[T]) ForEach(fn func(T)) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	for item := range s.items {
		fn(item)
	}
}

// Filter returns a new set containing only items for which the predicate returns true.
func (s *Set[T]) Filter(predicate func(T) bool) *Set[T] {
	s.lock.RLock()
	defer s.lock.RUnlock()

	result := NewWithCapacity[T](len(s.items))
	for item := range s.items {
		if predicate(item) {
			result.items[item] = struct{}{}
		}
	}
	return result
}
