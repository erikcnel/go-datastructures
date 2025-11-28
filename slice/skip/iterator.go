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

package skip

const iteratorExhausted = -2

// iterator represents an object that can be iterated. It will
// return false on Next and zero value on Value if there are no further
// values to be iterated.
type iterator[T Comparable[T]] struct {
	first bool
	n     *node[T]
}

// Next returns a bool indicating if there are any further values
// in this iterator.
func (iter *iterator[T]) Next() bool {
	if iter.first {
		iter.first = false
		return iter.n != nil && iter.n.hasEntry
	}

	if iter.n == nil {
		return false
	}

	iter.n = iter.n.forward[0]
	return iter.n != nil && iter.n.hasEntry
}

// Value returns a value representing the iterator's present
// position in the query. Returns zero value if no values remain to iterate.
func (iter *iterator[T]) Value() T {
	if iter.n == nil || !iter.n.hasEntry {
		var zero T
		return zero
	}

	return iter.n.entry
}

// exhaust is a helper method to exhaust this iterator and return
// all remaining entries.
func (iter *iterator[T]) exhaust() []T {
	entries := make([]T, 0, 10)
	for iter.Next() {
		entries = append(entries, iter.Value())
	}

	return entries
}

// nilIterator returns an iterator that will always return false
// for Next and zero value for Value.
func nilIterator[T Comparable[T]]() *iterator[T] {
	return &iterator[T]{}
}
