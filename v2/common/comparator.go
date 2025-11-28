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

// Package common provides shared interfaces and constraints for the v2 data structures.
package common

import "cmp"

// Ordered is a constraint that matches any ordered type.
// This includes all integer, float, and string types.
type Ordered = cmp.Ordered

// Comparator is a generic interface for items that can be compared.
// Returns a positive number if this item is greater, 0 if equal,
// negative number if less than other.
type Comparator[T any] interface {
	Compare(other T) int
}

// CompareFunc is a function type for comparing two values.
// Returns negative if a < b, 0 if a == b, positive if a > b.
type CompareFunc[T any] func(a, b T) int

// OrderedCompare returns a CompareFunc for any ordered type.
func OrderedCompare[T Ordered]() CompareFunc[T] {
	return func(a, b T) int {
		return cmp.Compare(a, b)
	}
}

// Less returns true if a < b for ordered types.
func Less[T Ordered](a, b T) bool {
	return cmp.Less(a, b)
}

// Equal returns true if a == b for comparable types.
func Equal[T comparable](a, b T) bool {
	return a == b
}
