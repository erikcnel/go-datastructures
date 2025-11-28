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

package set

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetNew(t *testing.T) {
	s := New[string]("a", "b", "c")

	assert.Equal(t, 3, s.Len())
	assert.True(t, s.Exists("a"))
	assert.True(t, s.Exists("b"))
	assert.True(t, s.Exists("c"))
}

func TestSetAdd(t *testing.T) {
	s := New[int]()

	s.Add(1, 2, 3)
	assert.Equal(t, 3, s.Len())

	// Adding duplicates shouldn't increase length
	s.Add(2, 3, 4)
	assert.Equal(t, 4, s.Len())
}

func TestSetRemove(t *testing.T) {
	s := New[string]("a", "b", "c")

	s.Remove("b")
	assert.Equal(t, 2, s.Len())
	assert.False(t, s.Exists("b"))

	// Removing non-existent item is a no-op
	s.Remove("x")
	assert.Equal(t, 2, s.Len())
}

func TestSetExists(t *testing.T) {
	s := New[int](1, 2, 3)

	assert.True(t, s.Exists(1))
	assert.True(t, s.Contains(2)) // alias
	assert.False(t, s.Exists(4))
}

func TestSetToSlice(t *testing.T) {
	s := New[int](3, 1, 2)

	slice := s.ToSlice()
	assert.Len(t, slice, 3)

	// Sort for consistent comparison
	sort.Ints(slice)
	assert.Equal(t, []int{1, 2, 3}, slice)
}

func TestSetClear(t *testing.T) {
	s := New[string]("a", "b", "c")

	s.Clear()
	assert.Equal(t, 0, s.Len())
	assert.True(t, s.IsEmpty())
}

func TestSetAll(t *testing.T) {
	s := New[int](1, 2, 3, 4, 5)

	assert.True(t, s.All(1, 2, 3))
	assert.False(t, s.All(1, 2, 6))
}

func TestSetAny(t *testing.T) {
	s := New[int](1, 2, 3)

	assert.True(t, s.Any(3, 4, 5))
	assert.False(t, s.Any(4, 5, 6))
}

func TestSetUnion(t *testing.T) {
	s1 := New[int](1, 2, 3)
	s2 := New[int](3, 4, 5)

	union := s1.Union(s2)

	assert.Equal(t, 5, union.Len())
	assert.True(t, union.All(1, 2, 3, 4, 5))
}

func TestSetIntersection(t *testing.T) {
	s1 := New[int](1, 2, 3, 4)
	s2 := New[int](3, 4, 5, 6)

	intersection := s1.Intersection(s2)

	assert.Equal(t, 2, intersection.Len())
	assert.True(t, intersection.All(3, 4))
}

func TestSetDifference(t *testing.T) {
	s1 := New[int](1, 2, 3, 4)
	s2 := New[int](3, 4, 5, 6)

	diff := s1.Difference(s2)

	assert.Equal(t, 2, diff.Len())
	assert.True(t, diff.All(1, 2))
}

func TestSetSymmetricDifference(t *testing.T) {
	s1 := New[int](1, 2, 3)
	s2 := New[int](2, 3, 4)

	symDiff := s1.SymmetricDifference(s2)

	assert.Equal(t, 2, symDiff.Len())
	assert.True(t, symDiff.All(1, 4))
}

func TestSetIsSubset(t *testing.T) {
	s1 := New[int](1, 2)
	s2 := New[int](1, 2, 3, 4)

	assert.True(t, s1.IsSubset(s2))
	assert.False(t, s2.IsSubset(s1))
}

func TestSetIsSuperset(t *testing.T) {
	s1 := New[int](1, 2, 3, 4)
	s2 := New[int](1, 2)

	assert.True(t, s1.IsSuperset(s2))
	assert.False(t, s2.IsSuperset(s1))
}

func TestSetEqual(t *testing.T) {
	s1 := New[string]("a", "b", "c")
	s2 := New[string]("c", "b", "a")
	s3 := New[string]("a", "b", "d")

	assert.True(t, s1.Equal(s2))
	assert.False(t, s1.Equal(s3))
}

func TestSetClone(t *testing.T) {
	s1 := New[int](1, 2, 3)
	s2 := s1.Clone()

	assert.True(t, s1.Equal(s2))

	// Modifying clone shouldn't affect original
	s2.Add(4)
	assert.False(t, s1.Exists(4))
}

func TestSetForEach(t *testing.T) {
	s := New[int](1, 2, 3)

	sum := 0
	s.ForEach(func(item int) {
		sum += item
	})

	assert.Equal(t, 6, sum)
}

func TestSetFilter(t *testing.T) {
	s := New[int](1, 2, 3, 4, 5, 6)

	evens := s.Filter(func(n int) bool {
		return n%2 == 0
	})

	assert.Equal(t, 3, evens.Len())
	assert.True(t, evens.All(2, 4, 6))
}

// Test with struct type
type point struct {
	x, y int
}

func TestSetWithStruct(t *testing.T) {
	s := New[point]()

	s.Add(point{1, 2}, point{3, 4})

	assert.True(t, s.Exists(point{1, 2}))
	assert.False(t, s.Exists(point{1, 3}))
}
