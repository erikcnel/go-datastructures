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

package list

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmptyList(t *testing.T) {
	l := Empty[int]()

	assert.True(t, l.IsEmpty())
	assert.Equal(t, uint(0), l.Length())

	_, ok := l.Head()
	assert.False(t, ok)

	_, ok = l.Tail()
	assert.False(t, ok)
}

func TestListAdd(t *testing.T) {
	l := Empty[string]().Add("a").Add("b").Add("c")

	assert.False(t, l.IsEmpty())
	assert.Equal(t, uint(3), l.Length())

	head, ok := l.Head()
	assert.True(t, ok)
	assert.Equal(t, "c", head)
}

func TestListTail(t *testing.T) {
	l := Empty[int]().Add(1).Add(2).Add(3)

	tail, ok := l.Tail()
	require.True(t, ok)

	head, ok := tail.Head()
	assert.True(t, ok)
	assert.Equal(t, 2, head)
}

func TestListGet(t *testing.T) {
	l := Empty[int]().Add(1).Add(2).Add(3)
	// List is now [3, 2, 1] (3 is head)

	val, ok := l.Get(0)
	assert.True(t, ok)
	assert.Equal(t, 3, val)

	val, ok = l.Get(1)
	assert.True(t, ok)
	assert.Equal(t, 2, val)

	val, ok = l.Get(2)
	assert.True(t, ok)
	assert.Equal(t, 1, val)

	_, ok = l.Get(3)
	assert.False(t, ok)
}

func TestListInsert(t *testing.T) {
	l := Empty[int]().Add(1).Add(3)
	// List is now [3, 1]

	l2, err := l.Insert(2, 1)
	require.NoError(t, err)
	// List is now [3, 2, 1]

	val, ok := l2.Get(1)
	assert.True(t, ok)
	assert.Equal(t, 2, val)
}

func TestListRemove(t *testing.T) {
	l := Empty[int]().Add(1).Add(2).Add(3)
	// List is [3, 2, 1]

	l2, err := l.Remove(1)
	require.NoError(t, err)
	// List is [3, 1]

	assert.Equal(t, uint(2), l2.Length())

	val, ok := l2.Get(1)
	assert.True(t, ok)
	assert.Equal(t, 1, val)
}

func TestListFind(t *testing.T) {
	l := Empty[int]().Add(1).Add(2).Add(3).Add(4).Add(5)

	found, ok := l.Find(func(x int) bool { return x%2 == 0 })
	assert.True(t, ok)
	assert.Equal(t, 4, found) // First even number from head

	_, ok = l.Find(func(x int) bool { return x > 10 })
	assert.False(t, ok)
}

func TestListFindIndex(t *testing.T) {
	l := Empty[string]().Add("a").Add("b").Add("c")
	// List is [c, b, a]

	idx := l.FindIndex(func(s string) bool { return s == "b" })
	assert.Equal(t, 1, idx)

	idx = l.FindIndex(func(s string) bool { return s == "x" })
	assert.Equal(t, -1, idx)
}

func TestListMap(t *testing.T) {
	l := Empty[int]().Add(1).Add(2).Add(3)
	// List is [3, 2, 1]

	doubled := l.Map(func(x int) int { return x * 2 })
	// Map traverses from head, but appends, so result is [2, 4, 6]
	assert.Equal(t, []int{2, 4, 6}, doubled)
}

func TestListForEach(t *testing.T) {
	l := Empty[int]().Add(1).Add(2).Add(3)

	sum := 0
	l.ForEach(func(x int) {
		sum += x
	})
	assert.Equal(t, 6, sum)
}

func TestListFilter(t *testing.T) {
	l := Empty[int]().Add(1).Add(2).Add(3).Add(4).Add(5)
	// List is [5, 4, 3, 2, 1]

	evens := l.Filter(func(x int) bool { return x%2 == 0 })

	assert.Equal(t, uint(2), evens.Length())
}

func TestListReduce(t *testing.T) {
	l := Empty[int]().Add(1).Add(2).Add(3).Add(4)
	// List is [4, 3, 2, 1]

	sum := l.Reduce(func(acc, x int) int { return acc + x }, 0)
	assert.Equal(t, 10, sum)
}

func TestListToSlice(t *testing.T) {
	l := Empty[int]().Add(1).Add(2).Add(3)
	// List is [3, 2, 1]

	slice := l.ToSlice()
	assert.Equal(t, []int{3, 2, 1}, slice)
}

func TestListReverse(t *testing.T) {
	l := Empty[int]().Add(1).Add(2).Add(3)
	// List is [3, 2, 1]

	reversed := l.Reverse()
	// Reversed is [1, 2, 3]

	slice := reversed.ToSlice()
	assert.Equal(t, []int{1, 2, 3}, slice)
}

func TestFromSlice(t *testing.T) {
	l := FromSlice([]int{1, 2, 3})
	// Items added in order, so list is [3, 2, 1]

	head, ok := l.Head()
	assert.True(t, ok)
	assert.Equal(t, 3, head)
}

func TestFromSliceReversed(t *testing.T) {
	l := FromSliceReversed([]int{1, 2, 3})
	// Items added in reverse, so list is [1, 2, 3]

	head, ok := l.Head()
	assert.True(t, ok)
	assert.Equal(t, 1, head)
}

func TestListImmutability(t *testing.T) {
	l1 := Empty[int]().Add(1).Add(2)
	l2 := l1.Add(3)

	// l1 should be unchanged
	assert.Equal(t, uint(2), l1.Length())
	assert.Equal(t, uint(3), l2.Length())
}

