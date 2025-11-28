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

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func generateMockEntries(num int) []mockEntry {
	entries := make([]mockEntry, 0, num)
	for i := range num {
		entries = append(entries, mockEntry(i))
	}
	return entries
}

// Helper to check if values were found
func assertFound(t *testing.T, results []mockEntry, found []bool, expected []mockEntry) {
	assert.Equal(t, len(expected), len(results))
	for i, exp := range expected {
		if exp == -1 { // -1 indicates "not found" in our tests
			assert.False(t, found[i], "expected not found at index %d", i)
		} else {
			assert.True(t, found[i], "expected found at index %d", i)
			assert.Equal(t, exp, results[i])
		}
	}
}

// Helper to check overwritten values
func assertOverwritten(t *testing.T, results []mockEntry, wasOverwritten []bool, expected []mockEntry) {
	assert.Equal(t, len(expected), len(results))
	for i, exp := range expected {
		if exp == -1 { // -1 indicates "not overwritten"
			assert.False(t, wasOverwritten[i], "expected not overwritten at index %d", i)
		} else {
			assert.True(t, wasOverwritten[i], "expected overwritten at index %d", i)
			assert.Equal(t, exp, results[i])
		}
	}
}

func TestAVLSimpleInsert(t *testing.T) {
	i1 := New[mockEntry]()
	m1 := mockEntry(5)
	m2 := mockEntry(10)

	i2, overwritten, wasOverwritten := i1.Insert(m1, m2)
	assertOverwritten(t, overwritten, wasOverwritten, []mockEntry{-1, -1})
	assert.Equal(t, uint64(2), i2.Len())
	assert.Equal(t, uint64(0), i1.Len())

	results, found := i1.Get(m1, m2)
	assertFound(t, results, found, []mockEntry{-1, -1})

	results, found = i2.Get(m1, m2)
	assertFound(t, results, found, []mockEntry{m1, m2})

	m3 := mockEntry(1)

	i3, overwritten, wasOverwritten := i2.Insert(m3)
	assertOverwritten(t, overwritten, wasOverwritten, []mockEntry{-1})
	assert.Equal(t, uint64(3), i3.Len())
	assert.Equal(t, uint64(2), i2.Len())
	assert.Equal(t, uint64(0), i1.Len())

	results, found = i3.Get(m1, m2, m3)
	assertFound(t, results, found, []mockEntry{m1, m2, m3})
}

func TestAVLInsertRightLeaning(t *testing.T) {
	i1 := New[mockEntry]()
	m1 := mockEntry(1)
	m2 := mockEntry(5)
	m3 := mockEntry(10)

	i2, overwritten, wasOverwritten := i1.Insert(m1, m2, m3)
	assertOverwritten(t, overwritten, wasOverwritten, []mockEntry{-1, -1, -1})
	assert.Equal(t, uint64(0), i1.Len())
	assert.Equal(t, uint64(3), i2.Len())

	results, found := i2.Get(m1, m2, m3)
	assertFound(t, results, found, []mockEntry{m1, m2, m3})

	results, found = i1.Get(m1, m2, m3)
	assertFound(t, results, found, []mockEntry{-1, -1, -1})

	m4 := mockEntry(15)
	m5 := mockEntry(20)

	i3, overwritten, wasOverwritten := i2.Insert(m4, m5)
	assertOverwritten(t, overwritten, wasOverwritten, []mockEntry{-1, -1})
	assert.Equal(t, uint64(5), i3.Len())
	assert.Equal(t, uint64(3), i2.Len())

	results, found = i2.Get(m4, m5)
	assertFound(t, results, found, []mockEntry{-1, -1})

	results, found = i3.Get(m4, m5)
	assertFound(t, results, found, []mockEntry{m4, m5})
}

func TestAVLInsertRightLeaningDoubleRotation(t *testing.T) {
	i1 := New[mockEntry]()
	m1 := mockEntry(1)
	m2 := mockEntry(10)
	m3 := mockEntry(5)

	i2, overwritten, wasOverwritten := i1.Insert(m1, m2, m3)
	assert.Equal(t, uint64(3), i2.Len())
	assertOverwritten(t, overwritten, wasOverwritten, []mockEntry{-1, -1, -1})

	results, found := i1.Get(m1, m2, m3)
	assertFound(t, results, found, []mockEntry{-1, -1, -1})

	results, found = i2.Get(m1, m2, m3)
	assertFound(t, results, found, []mockEntry{m1, m2, m3})

	m4 := mockEntry(20)
	m5 := mockEntry(15)

	i3, overwritten, wasOverwritten := i2.Insert(m4, m5)
	assertOverwritten(t, overwritten, wasOverwritten, []mockEntry{-1, -1})
	assert.Equal(t, uint64(5), i3.Len())
	assert.Equal(t, uint64(3), i2.Len())

	results, found = i2.Get(m4, m5)
	assertFound(t, results, found, []mockEntry{-1, -1})

	results, found = i3.Get(m4, m5)
	assertFound(t, results, found, []mockEntry{m4, m5})
}

func TestAVLInsertLeftLeaning(t *testing.T) {
	i1 := New[mockEntry]()
	m1 := mockEntry(20)
	m2 := mockEntry(15)
	m3 := mockEntry(10)

	i2, overwritten, wasOverwritten := i1.Insert(m1, m2, m3)
	assertOverwritten(t, overwritten, wasOverwritten, []mockEntry{-1, -1, -1})
	assert.Equal(t, uint64(0), i1.Len())
	assert.Equal(t, uint64(3), i2.Len())

	results, found := i1.Get(m1, m2, m3)
	assertFound(t, results, found, []mockEntry{-1, -1, -1})

	results, found = i2.Get(m1, m2, m3)
	assertFound(t, results, found, []mockEntry{m1, m2, m3})

	m4 := mockEntry(5)
	m5 := mockEntry(1)

	i3, overwritten, wasOverwritten := i2.Insert(m4, m5)
	assertOverwritten(t, overwritten, wasOverwritten, []mockEntry{-1, -1})
	assert.Equal(t, uint64(3), i2.Len())
	assert.Equal(t, uint64(5), i3.Len())

	results, found = i2.Get(m4, m5)
	assertFound(t, results, found, []mockEntry{-1, -1})

	results, found = i3.Get(m4, m5)
	assertFound(t, results, found, []mockEntry{m4, m5})
}

func TestAVLInsertLeftLeaningDoubleRotation(t *testing.T) {
	i1 := New[mockEntry]()
	m1 := mockEntry(20)
	m2 := mockEntry(10)
	m3 := mockEntry(15)

	i2, overwritten, wasOverwritten := i1.Insert(m1, m2, m3)
	assertOverwritten(t, overwritten, wasOverwritten, []mockEntry{-1, -1, -1})
	assert.Equal(t, uint64(0), i1.Len())
	assert.Equal(t, uint64(3), i2.Len())

	results, found := i1.Get(m1, m2, m3)
	assertFound(t, results, found, []mockEntry{-1, -1, -1})

	results, found = i2.Get(m1, m2, m3)
	assertFound(t, results, found, []mockEntry{m1, m2, m3})

	m4 := mockEntry(1)
	m5 := mockEntry(5)

	i3, overwritten, wasOverwritten := i2.Insert(m4, m5)
	assertOverwritten(t, overwritten, wasOverwritten, []mockEntry{-1, -1})
	assert.Equal(t, uint64(3), i2.Len())
	assert.Equal(t, uint64(5), i3.Len())

	results, found = i2.Get(m4, m5)
	assertFound(t, results, found, []mockEntry{-1, -1})

	results, found = i3.Get(m4, m5)
	assertFound(t, results, found, []mockEntry{m4, m5})

	results, found = i3.Get(m1, m2, m3)
	assertFound(t, results, found, []mockEntry{m1, m2, m3})
}

func TestAVLInsertOverwrite(t *testing.T) {
	i1 := New[mockEntry]()
	m1 := mockEntry(20)
	m2 := mockEntry(10)
	m3 := mockEntry(15)

	i2, _, _ := i1.Insert(m1, m2, m3)
	m4 := mockEntry(15)

	i3, overwritten, wasOverwritten := i2.Insert(m4)
	assertOverwritten(t, overwritten, wasOverwritten, []mockEntry{m3})
	assert.Equal(t, uint64(3), i2.Len())
	assert.Equal(t, uint64(3), i3.Len())

	results, found := i3.Get(m4)
	assertFound(t, results, found, []mockEntry{m4})

	results, found = i2.Get(m3)
	assertFound(t, results, found, []mockEntry{m3})
}

func TestAVLSimpleDelete(t *testing.T) {
	i1 := New[mockEntry]()
	m1 := mockEntry(10)
	m2 := mockEntry(15)
	m3 := mockEntry(20)

	i2, _, _ := i1.Insert(m1, m2, m3)

	i3, deleted, wasDeleted := i2.Delete(m3)
	assertOverwritten(t, deleted, wasDeleted, []mockEntry{m3})
	assert.Equal(t, uint64(3), i2.Len())
	assert.Equal(t, uint64(2), i3.Len())

	results, found := i2.Get(m1, m2, m3)
	assertFound(t, results, found, []mockEntry{m1, m2, m3})

	results, found = i3.Get(m1, m2, m3)
	assertFound(t, results, found, []mockEntry{m1, m2, -1})

	i4, deleted, wasDeleted := i3.Delete(m2)
	assertOverwritten(t, deleted, wasDeleted, []mockEntry{m2})
	assert.Equal(t, uint64(2), i3.Len())
	assert.Equal(t, uint64(1), i4.Len())

	results, found = i3.Get(m1, m2, m3)
	assertFound(t, results, found, []mockEntry{m1, m2, -1})

	results, found = i4.Get(m1, m2, m3)
	assertFound(t, results, found, []mockEntry{m1, -1, -1})

	i5, deleted, wasDeleted := i4.Delete(m1)
	assertOverwritten(t, deleted, wasDeleted, []mockEntry{m1})
	assert.Equal(t, uint64(0), i5.Len())
	assert.Equal(t, uint64(1), i4.Len())

	results, found = i4.Get(m1, m2, m3)
	assertFound(t, results, found, []mockEntry{m1, -1, -1})

	results, found = i5.Get(m1, m2, m3)
	assertFound(t, results, found, []mockEntry{-1, -1, -1})
}

func TestAVLDeleteWithRotation(t *testing.T) {
	i1 := New[mockEntry]()
	m1 := mockEntry(1)
	m2 := mockEntry(5)
	m3 := mockEntry(10)
	m4 := mockEntry(15)
	m5 := mockEntry(20)

	i2, _, _ := i1.Insert(m1, m2, m3, m4, m5)
	assert.Equal(t, uint64(5), i2.Len())

	i3, deleted, wasDeleted := i2.Delete(m1)
	assert.Equal(t, uint64(4), i3.Len())
	assertOverwritten(t, deleted, wasDeleted, []mockEntry{m1})

	results, found := i2.Get(m1, m2, m3, m4, m5)
	assertFound(t, results, found, []mockEntry{m1, m2, m3, m4, m5})

	results, found = i3.Get(m1, m2, m3, m4, m5)
	assertFound(t, results, found, []mockEntry{-1, m2, m3, m4, m5})
}

func TestAVLDeleteWithDoubleRotation(t *testing.T) {
	i1 := New[mockEntry]()
	m1 := mockEntry(1)
	m2 := mockEntry(5)
	m3 := mockEntry(10)
	m4 := mockEntry(15)

	i2, _, _ := i1.Insert(m2, m1, m3, m4)
	assert.Equal(t, uint64(4), i2.Len())

	i3, deleted, wasDeleted := i2.Delete(m1)
	assertOverwritten(t, deleted, wasDeleted, []mockEntry{m1})
	assert.Equal(t, uint64(3), i3.Len())

	results, found := i2.Get(m1, m2, m3, m4)
	assertFound(t, results, found, []mockEntry{m1, m2, m3, m4})

	results, found = i3.Get(m1, m2, m3, m4)
	assertFound(t, results, found, []mockEntry{-1, m2, m3, m4})
}

func TestAVLDeleteAll(t *testing.T) {
	i1 := New[mockEntry]()
	m1 := mockEntry(1)
	m2 := mockEntry(5)
	m3 := mockEntry(10)
	m4 := mockEntry(15)

	i2, _, _ := i1.Insert(m2, m1, m3, m4)
	assert.Equal(t, uint64(4), i2.Len())

	i3, deleted, wasDeleted := i2.Delete(m1, m2, m3, m4)
	assertOverwritten(t, deleted, wasDeleted, []mockEntry{m1, m2, m3, m4})
	assert.Equal(t, uint64(0), i3.Len())

	results, found := i3.Get(m1, m2, m3, m4)
	assertFound(t, results, found, []mockEntry{-1, -1, -1, -1})

	results, found = i2.Get(m1, m2, m3, m4)
	assertFound(t, results, found, []mockEntry{m1, m2, m3, m4})
}

func TestAVLDeleteNotLeaf(t *testing.T) {
	i1 := New[mockEntry]()
	m1 := mockEntry(1)
	m2 := mockEntry(5)
	m3 := mockEntry(10)
	m4 := mockEntry(15)

	i2, _, _ := i1.Insert(m2, m1, m3, m4)
	i3, deleted, wasDeleted := i2.Delete(m3)
	assertOverwritten(t, deleted, wasDeleted, []mockEntry{m3})
	assert.Equal(t, uint64(3), i3.Len())
}

func TestAVLBulkDeleteAll(t *testing.T) {
	i1 := New[mockEntry]()
	entries := generateMockEntries(5)
	i2, _, _ := i1.Insert(entries...)

	i3, deleted, wasDeleted := i2.Delete(entries...)
	for i, e := range entries {
		assert.True(t, wasDeleted[i])
		assert.Equal(t, e, deleted[i])
	}
	assert.Equal(t, uint64(0), i3.Len())

	i3, deleted, wasDeleted = i2.Delete(entries...)
	for i, e := range entries {
		assert.True(t, wasDeleted[i])
		assert.Equal(t, e, deleted[i])
	}
	assert.Equal(t, uint64(0), i3.Len())
}

func TestAVLDeleteReplay(t *testing.T) {
	i1 := New[mockEntry]()
	m1 := mockEntry(1)
	m2 := mockEntry(5)
	m3 := mockEntry(10)
	m4 := mockEntry(15)

	i2, _, _ := i1.Insert(m2, m1, m3, m4)

	i3, deleted, wasDeleted := i2.Delete(m3)
	assert.Equal(t, uint64(3), i3.Len())
	assertOverwritten(t, deleted, wasDeleted, []mockEntry{m3})
	assert.Equal(t, uint64(4), i2.Len())

	i3, deleted, wasDeleted = i2.Delete(m3)
	assert.Equal(t, uint64(3), i3.Len())
	assertOverwritten(t, deleted, wasDeleted, []mockEntry{m3})
	assert.Equal(t, uint64(4), i2.Len())
}

func TestAVLFails(t *testing.T) {
	keys := []mockEntry{
		mockEntry(0),
		mockEntry(1),
		mockEntry(3),
		mockEntry(4),
		mockEntry(5),
		mockEntry(6),
		mockEntry(7),
		mockEntry(2),
	}
	i1 := New[mockEntry]()
	for _, k := range keys {
		i1, _, _ = i1.Insert(k)
	}

	for _, k := range keys {
		var deleted []mockEntry
		var wasDeleted []bool
		i1, deleted, wasDeleted = i1.Delete(k)
		assert.True(t, wasDeleted[0])
		assert.Equal(t, k, deleted[0])
	}
}

func BenchmarkImmutableInsert(b *testing.B) {
	numItems := b.N
	sl := New[mockEntry]()

	entries := generateMockEntries(numItems)
	sl, _, _ = sl.Insert(entries...)

	for i := 0; b.Loop(); i++ {
		sl, _, _ = sl.Insert(entries[i%numItems])
	}
}

func BenchmarkImmutableGet(b *testing.B) {
	numItems := b.N
	sl := New[mockEntry]()

	entries := generateMockEntries(numItems)
	sl, _, _ = sl.Insert(entries...)

	for i := 0; b.Loop(); i++ {
		sl.Get(entries[i%numItems])
	}
}

func BenchmarkImmutableBulkInsert(b *testing.B) {
	numItems := b.N
	sl := New[mockEntry]()

	entries := generateMockEntries(numItems)

	for b.Loop() {
		sl.Insert(entries...)
	}
}

func BenchmarkImmutableDelete(b *testing.B) {
	numItems := b.N
	sl := New[mockEntry]()

	entries := generateMockEntries(numItems)
	sl, _, _ = sl.Insert(entries...)

	for i := 0; b.Loop(); i++ {
		sl, _, _ = sl.Delete(entries[i%numItems])
	}
}

func BenchmarkImmutableBulkDelete(b *testing.B) {
	numItems := b.N
	sl := New[mockEntry]()

	entries := generateMockEntries(numItems)
	sl, _, _ = sl.Insert(entries...)

	for b.Loop() {
		sl.Delete(entries...)
	}
}
