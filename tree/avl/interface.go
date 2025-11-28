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

package avl

// Comparable is a constraint for types that can be compared for ordering.
// The Compare method should return:
//   - negative value if receiver < other
//   - zero if receiver == other
//   - positive value if receiver > other
type Comparable[T any] interface {
	Compare(other T) int
}

// Entry is kept for backward compatibility.
// Deprecated: Use Comparable[T] with the generic Immutable[T] instead.
type Entry interface {
	Compare(Entry) int
}

// Entries is a list of type Entry.
// Deprecated: Use []T with the generic Immutable[T] instead.
type Entries []Entry
