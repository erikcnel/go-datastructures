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

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func generateMockEntries(num int) []mockEntry {
	entries := make([]mockEntry, 0, num)
	for i := uint64(0); i < uint64(num); i++ {
		entries = append(entries, newMockEntry(i))
	}

	return entries
}

func generateRandomMockEntries(num int) []mockEntry {
	entries := make([]mockEntry, 0, num)
	for range num {
		entries = append(entries, newMockEntry(uint64(rand.Int())))
	}

	return entries
}

func TestInsertByPosition(t *testing.T) {
	m1 := newMockEntry(5)
	m2 := newMockEntry(6)
	m3 := newMockEntry(2)
	sl := New[mockEntry](uint8(0))
	sl.InsertAtPosition(2, m1)
	sl.InsertAtPosition(0, m2)
	sl.InsertAtPosition(0, m3)

	v, ok := sl.ByPosition(0)
	assert.True(t, ok)
	assert.Equal(t, m3, v)

	v, ok = sl.ByPosition(1)
	assert.True(t, ok)
	assert.Equal(t, m2, v)

	v, ok = sl.ByPosition(2)
	assert.True(t, ok)
	assert.Equal(t, m1, v)

	_, ok = sl.ByPosition(3)
	assert.False(t, ok)
}

func TestGetByPosition(t *testing.T) {
	m1 := newMockEntry(5)
	m2 := newMockEntry(6)
	sl := New[mockEntry](uint8(0))
	sl.Insert(m1, m2)

	v, ok := sl.ByPosition(0)
	assert.True(t, ok)
	assert.Equal(t, m1, v)

	v, ok = sl.ByPosition(1)
	assert.True(t, ok)
	assert.Equal(t, m2, v)

	_, ok = sl.ByPosition(2)
	assert.False(t, ok)
}

func TestSplitAt(t *testing.T) {
	m1 := newMockEntry(3)
	m2 := newMockEntry(5)
	m3 := newMockEntry(7)
	sl := New[mockEntry](uint8(0))
	sl.Insert(m1, m2, m3)

	left, right := sl.SplitAt(1)
	assert.Equal(t, uint64(2), left.Len())
	assert.Equal(t, uint64(1), right.Len())

	results, found := left.Get(m1, m2, m3)
	assert.True(t, found[0])
	assert.True(t, found[1])
	assert.False(t, found[2])
	assert.Equal(t, m1, results[0])
	assert.Equal(t, m2, results[1])

	results, found = right.Get(m1, m2, m3)
	assert.False(t, found[0])
	assert.False(t, found[1])
	assert.True(t, found[2])
	assert.Equal(t, m3, results[2])

	v, ok := left.ByPosition(0)
	assert.True(t, ok)
	assert.Equal(t, m1, v)

	v, ok = left.ByPosition(1)
	assert.True(t, ok)
	assert.Equal(t, m2, v)

	v, ok = right.ByPosition(0)
	assert.True(t, ok)
	assert.Equal(t, m3, v)

	_, ok = left.ByPosition(2)
	assert.False(t, ok)

	_, ok = right.ByPosition(1)
	assert.False(t, ok)
}

func TestSplitLargeSkipList(t *testing.T) {
	entries := generateMockEntries(100)
	leftEntries := entries[:50]
	rightEntries := entries[50:]
	sl := New[mockEntry](uint64(0))
	sl.Insert(entries...)

	left, right := sl.SplitAt(49)
	assert.Equal(t, uint64(50), left.Len())
	for _, le := range leftEntries {
		result, pos, ok := left.GetWithPosition(le)
		assert.True(t, ok)
		assert.Equal(t, le, result)
		v, ok := left.ByPosition(pos)
		assert.True(t, ok)
		assert.Equal(t, le, v)
	}

	assert.Equal(t, uint64(50), right.Len())
	for _, re := range rightEntries {
		result, pos, ok := right.GetWithPosition(re)
		assert.True(t, ok)
		assert.Equal(t, re, result)
		v, ok := right.ByPosition(pos)
		assert.True(t, ok)
		assert.Equal(t, re, v)
	}
}

func TestSplitLargeSkipListOddNumber(t *testing.T) {
	entries := generateMockEntries(99)
	leftEntries := entries[:50]
	rightEntries := entries[50:]
	sl := New[mockEntry](uint64(0))
	sl.Insert(entries...)

	left, right := sl.SplitAt(49)
	assert.Equal(t, uint64(50), left.Len())
	for _, le := range leftEntries {
		result, pos, ok := left.GetWithPosition(le)
		assert.True(t, ok)
		assert.Equal(t, le, result)
		v, ok := left.ByPosition(pos)
		assert.True(t, ok)
		assert.Equal(t, le, v)
	}

	assert.Equal(t, uint64(49), right.Len())
	for _, re := range rightEntries {
		result, pos, ok := right.GetWithPosition(re)
		assert.True(t, ok)
		assert.Equal(t, re, result)
		v, ok := right.ByPosition(pos)
		assert.True(t, ok)
		assert.Equal(t, re, v)
	}
}

func TestSplitAtSkipListLength(t *testing.T) {
	entries := generateMockEntries(5)
	sl := New[mockEntry](uint64(0))
	sl.Insert(entries...)

	left, right := sl.SplitAt(4)
	assert.Equal(t, sl, left)
	assert.Nil(t, right)
}

func TestGetWithPosition(t *testing.T) {
	m1 := newMockEntry(5)
	m2 := newMockEntry(6)
	sl := New[mockEntry](uint8(0))
	sl.Insert(m1, m2)

	e, pos, ok := sl.GetWithPosition(m1)
	assert.True(t, ok)
	assert.Equal(t, m1, e)
	assert.Equal(t, uint64(0), pos)

	e, pos, ok = sl.GetWithPosition(m2)
	assert.True(t, ok)
	assert.Equal(t, m2, e)
	assert.Equal(t, uint64(1), pos)
}

func TestReplaceAtPosition(t *testing.T) {
	m1 := newMockEntry(5)
	m2 := newMockEntry(6)
	sl := New[mockEntry](uint8(0))

	sl.Insert(m1, m2)
	m3 := newMockEntry(9)
	sl.ReplaceAtPosition(0, m3)

	v, ok := sl.ByPosition(0)
	assert.True(t, ok)
	assert.Equal(t, m3, v)

	v, ok = sl.ByPosition(1)
	assert.True(t, ok)
	assert.Equal(t, m2, v)
}

func TestInsertRandomGetByPosition(t *testing.T) {
	entries := generateRandomMockEntries(100)
	sl := New[mockEntry](uint64(0))
	sl.Insert(entries...)

	for _, e := range entries {
		_, pos, ok := sl.GetWithPosition(e)
		if ok {
			v, ok := sl.ByPosition(pos)
			assert.True(t, ok)
			assert.Equal(t, e, v)
		}
	}
}

func TestGetManyByPosition(t *testing.T) {
	entries := generateMockEntries(10)
	sl := New[mockEntry](uint64(0))
	sl.Insert(entries...)

	for i, e := range entries {
		v, ok := sl.ByPosition(uint64(i))
		assert.True(t, ok)
		assert.Equal(t, e, v)
	}
}

func TestGetPositionAfterDelete(t *testing.T) {
	m1 := newMockEntry(5)
	m2 := newMockEntry(6)
	sl := New[mockEntry](uint8(0))
	sl.Insert(m1, m2)

	sl.Delete(m1)

	v, ok := sl.ByPosition(0)
	assert.True(t, ok)
	assert.Equal(t, m2, v)

	_, ok = sl.ByPosition(1)
	assert.False(t, ok)

	sl.Delete(m2)

	_, ok = sl.ByPosition(0)
	assert.False(t, ok)

	_, ok = sl.ByPosition(1)
	assert.False(t, ok)
}

func TestGetPositionBulkDelete(t *testing.T) {
	es := generateMockEntries(20)
	e1 := es[:10]
	e2 := es[10:]
	sl := New[mockEntry](uint64(0))
	sl.Insert(e1...)
	sl.Insert(e2...)

	for _, e := range e1 {
		sl.Delete(e)
	}
	for i, e := range e2 {
		v, ok := sl.ByPosition(uint64(i))
		assert.True(t, ok)
		assert.Equal(t, e, v)
	}
}

func TestSimpleInsert(t *testing.T) {
	m1 := newMockEntry(5)
	m2 := newMockEntry(6)

	sl := New[mockEntry](uint8(0))

	overwritten, wasOverwritten := sl.Insert(m1)
	results, found := sl.Get(m1)
	assert.True(t, found[0])
	assert.Equal(t, m1, results[0])
	assert.Equal(t, uint64(1), sl.Len())
	assert.False(t, wasOverwritten[0])

	results, found = sl.Get(mockEntry(1))
	assert.False(t, found[0])

	overwritten, wasOverwritten = sl.Insert(m2)
	results, found = sl.Get(m2)
	assert.True(t, found[0])
	assert.Equal(t, m2, results[0])

	results, found = sl.Get(mockEntry(7))
	assert.False(t, found[0])
	assert.Equal(t, uint64(2), sl.Len())
	assert.False(t, wasOverwritten[0])
	_ = overwritten
}

func TestSimpleOverwrite(t *testing.T) {
	m1 := newMockEntry(5)
	m2 := newMockEntry(5)

	sl := New[mockEntry](uint8(0))

	_, wasOverwritten := sl.Insert(m1)
	assert.False(t, wasOverwritten[0])
	assert.Equal(t, uint64(1), sl.Len())

	overwritten, wasOverwritten := sl.Insert(m2)
	assert.True(t, wasOverwritten[0])
	assert.Equal(t, m1, overwritten[0])
	assert.Equal(t, uint64(1), sl.Len())
}

func TestInsertOutOfOrder(t *testing.T) {
	m1 := newMockEntry(6)
	m2 := newMockEntry(5)

	sl := New[mockEntry](uint8(0))

	_, wasOverwritten := sl.Insert(m1, m2)
	assert.False(t, wasOverwritten[0])
	assert.False(t, wasOverwritten[1])

	results, found := sl.Get(m1, m2)
	assert.True(t, found[0])
	assert.True(t, found[1])
	assert.Equal(t, m1, results[0])
	assert.Equal(t, m2, results[1])
}

func TestSimpleDelete(t *testing.T) {
	m1 := newMockEntry(5)
	sl := New[mockEntry](uint8(0))
	sl.Insert(m1)

	deleted, wasDeleted := sl.Delete(m1)
	assert.True(t, wasDeleted[0])
	assert.Equal(t, m1, deleted[0])
	assert.Equal(t, uint64(0), sl.Len())

	_, found := sl.Get(m1)
	assert.False(t, found[0])

	deleted, wasDeleted = sl.Delete(m1)
	assert.False(t, wasDeleted[0])
	_ = deleted
}

func TestDeleteAll(t *testing.T) {
	m1 := newMockEntry(5)
	m2 := newMockEntry(6)
	sl := New[mockEntry](uint8(0))
	sl.Insert(m1, m2)

	deleted, wasDeleted := sl.Delete(m1, m2)
	assert.True(t, wasDeleted[0])
	assert.True(t, wasDeleted[1])
	assert.Equal(t, m1, deleted[0])
	assert.Equal(t, m2, deleted[1])
	assert.Equal(t, uint64(0), sl.Len())

	_, found := sl.Get(m1, m2)
	assert.False(t, found[0])
	assert.False(t, found[1])
}

func TestIter(t *testing.T) {
	sl := New[mockEntry](uint8(0))
	m1 := newMockEntry(5)
	m2 := newMockEntry(10)

	sl.Insert(m1, m2)

	iter := sl.Iter(mockEntry(0)).(*iterator[mockEntry])
	assert.Equal(t, []mockEntry{m1, m2}, iter.exhaust())

	iter = sl.Iter(mockEntry(5)).(*iterator[mockEntry])
	assert.Equal(t, []mockEntry{m1, m2}, iter.exhaust())

	iter = sl.Iter(mockEntry(6)).(*iterator[mockEntry])
	assert.Equal(t, []mockEntry{m2}, iter.exhaust())

	iter = sl.Iter(mockEntry(10)).(*iterator[mockEntry])
	assert.Equal(t, []mockEntry{m2}, iter.exhaust())

	iter = sl.Iter(mockEntry(11)).(*iterator[mockEntry])
	assert.Equal(t, []mockEntry{}, iter.exhaust())
}

func TestIterAtPosition(t *testing.T) {
	sl := New[mockEntry](uint8(0))
	m1 := newMockEntry(5)
	m2 := newMockEntry(10)

	sl.Insert(m1, m2)

	iter := sl.IterAtPosition(0).(*iterator[mockEntry])
	assert.Equal(t, []mockEntry{m1, m2}, iter.exhaust())

	iter = sl.IterAtPosition(1).(*iterator[mockEntry])
	assert.Equal(t, []mockEntry{m2}, iter.exhaust())

	iter = sl.IterAtPosition(2).(*iterator[mockEntry])
	assert.Equal(t, []mockEntry{}, iter.exhaust())
}

func BenchmarkInsert(b *testing.B) {
	numItems := b.N
	sl := New[mockEntry](uint64(0))

	entries := generateMockEntries(numItems)

	for i := 0; b.Loop(); i++ {
		sl.Insert(entries[i%numItems])
	}
}

func BenchmarkGet(b *testing.B) {
	numItems := b.N
	sl := New[mockEntry](uint64(0))

	entries := generateMockEntries(numItems)
	sl.Insert(entries...)

	for i := 0; b.Loop(); i++ {
		sl.Get(entries[i%numItems])
	}
}

func BenchmarkDelete(b *testing.B) {
	numItems := b.N
	sl := New[mockEntry](uint64(0))

	entries := generateMockEntries(numItems)
	sl.Insert(entries...)

	for i := 0; b.Loop(); i++ {
		sl.Delete(entries[i])
	}
}

func BenchmarkPrepend(b *testing.B) {
	numItems := b.N
	sl := New[mockEntry](uint64(0))

	entries := make([]mockEntry, 0, numItems)
	for i := b.N; i < b.N+numItems; i++ {
		entries = append(entries, newMockEntry(uint64(i)))
	}

	sl.Insert(entries...)

	for i := 0; b.Loop(); i++ {
		sl.Insert(newMockEntry(uint64(i)))
	}
}

func BenchmarkByPosition(b *testing.B) {
	numItems := b.N
	sl := New[mockEntry](uint64(0))
	entries := generateMockEntries(numItems)
	sl.Insert(entries...)

	for i := 0; b.Loop(); i++ {
		sl.ByPosition(uint64(i % numItems))
	}
}

func BenchmarkInsertAtPosition(b *testing.B) {
	numItems := b.N
	sl := New[mockEntry](uint64(0))
	entries := generateRandomMockEntries(numItems)

	for i := 0; b.Loop(); i++ {
		sl.InsertAtPosition(0, entries[i%numItems])
	}
}
