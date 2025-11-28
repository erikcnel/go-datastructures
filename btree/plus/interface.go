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

package plus

// Comparable is a constraint for types that can be compared for ordering.
// The Compare method should return:
//   - negative value if receiver < other
//   - zero if receiver == other
//   - positive value if receiver > other
type Comparable[T any] interface {
	Compare(other T) int
}

// Key is kept for backward compatibility.
// Deprecated: Use Comparable[T] with the generic BTree[T] instead.
type Key interface {
	Compare(Key) int
}

// Keys is a typed list of Key interfaces.
// Deprecated: Use []T with the generic BTree[T] instead.
type Keys []Key

// Iterator defines an interface for traversing tree results in order.
type Iterator[T any] interface {
	// Next moves the iterator to the next position and returns
	// true if there is a value.
	Next() bool
	// Value returns the key at the current iterator position.
	// Returns zero value if exhausted or not yet started.
	Value() T
}
