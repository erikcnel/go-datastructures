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

import "github.com/Workiva/go-datastructures/common"

// Comparable is a constraint for types that can be compared for ordering.
// The Compare method should return:
//   - negative value if receiver < other
//   - zero if receiver == other
//   - positive value if receiver > other
type Comparable[T any] interface {
	Compare(other T) int
}

// Iterator defines a generic interface that allows a consumer to iterate
// all results of a query. All values will be visited in-order.
type Iterator[T any] interface {
	// Next returns a bool indicating if there is future value
	// in the iterator and moves the iterator to that value.
	Next() bool
	// Value returns a value representing the iterator's current
	// position. Returns zero value if exhausted.
	Value() T
}

// ComparatorWrapper wraps common.Comparator to implement Comparable[ComparatorWrapper].
// This allows using the generic SkipList with the old common.Comparator interface.
type ComparatorWrapper struct {
	C common.Comparator
}

// Compare implements Comparable[ComparatorWrapper]
func (cw ComparatorWrapper) Compare(other ComparatorWrapper) int {
	return cw.C.Compare(other.C)
}

// NewComparatorSkipList creates a SkipList that works with common.Comparator.
func NewComparatorSkipList(ifc any) *SkipList[ComparatorWrapper] {
	return New[ComparatorWrapper](ifc)
}
