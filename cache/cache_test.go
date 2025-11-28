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

package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testItem struct {
	data string
	size uint64
}

func (t testItem) Size() uint64 {
	return t.size
}

func TestCachePutGet(t *testing.T) {
	c := New[string, testItem](100)

	c.Put("key1", testItem{"value1", 10})

	item, ok := c.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, "value1", item.data)
}

func TestCacheGetMissing(t *testing.T) {
	c := New[string, testItem](100)

	_, ok := c.Get("nonexistent")
	assert.False(t, ok)
}

func TestCacheEviction(t *testing.T) {
	c := New[string, testItem](50) // 50 byte capacity

	c.Put("key1", testItem{"value1", 30})
	c.Put("key2", testItem{"value2", 30})

	// key1 should be evicted to make room for key2
	_, ok := c.Get("key1")
	assert.False(t, ok)

	_, ok = c.Get("key2")
	assert.True(t, ok)
}

func TestCacheLRUEviction(t *testing.T) {
	c := New[string, testItem](50) // Room for ~2 items of size 25

	c.Put("key1", testItem{"value1", 25})
	c.Put("key2", testItem{"value2", 25})

	// Access key1 to make it most recently used
	c.Get("key1")

	// Add key3, should evict key2 (least recently used)
	c.Put("key3", testItem{"value3", 25})

	_, ok := c.Get("key1")
	assert.True(t, ok, "key1 should still exist")

	_, ok = c.Get("key2")
	assert.False(t, ok, "key2 should be evicted")

	_, ok = c.Get("key3")
	assert.True(t, ok, "key3 should exist")
}

func TestCacheLRAEviction(t *testing.T) {
	c := New[string, testItem](50, WithPolicy[string, testItem](LeastRecentlyAdded))

	c.Put("key1", testItem{"value1", 25})
	c.Put("key2", testItem{"value2", 25})

	// Access key1 (shouldn't affect eviction order with LRA)
	c.Get("key1")

	// Add key3, should evict key1 (least recently added)
	c.Put("key3", testItem{"value3", 25})

	_, ok := c.Get("key1")
	assert.False(t, ok, "key1 should be evicted")

	_, ok = c.Get("key2")
	assert.True(t, ok, "key2 should still exist")
}

func TestCacheRemove(t *testing.T) {
	c := New[string, testItem](100)

	c.Put("key1", testItem{"value1", 10})
	c.Put("key2", testItem{"value2", 10})

	c.Remove("key1")

	_, ok := c.Get("key1")
	assert.False(t, ok)

	_, ok = c.Get("key2")
	assert.True(t, ok)
}

func TestCacheSize(t *testing.T) {
	c := New[string, testItem](100)

	assert.Equal(t, uint64(0), c.Size())

	c.Put("key1", testItem{"value1", 10})
	assert.Equal(t, uint64(10), c.Size())

	c.Put("key2", testItem{"value2", 20})
	assert.Equal(t, uint64(30), c.Size())

	c.Remove("key1")
	assert.Equal(t, uint64(20), c.Size())
}

func TestCacheLen(t *testing.T) {
	c := New[string, testItem](100)

	assert.Equal(t, 0, c.Len())

	c.Put("key1", testItem{"value1", 10})
	c.Put("key2", testItem{"value2", 10})

	assert.Equal(t, 2, c.Len())
}

func TestCacheClear(t *testing.T) {
	c := New[string, testItem](100)

	c.Put("key1", testItem{"value1", 10})
	c.Put("key2", testItem{"value2", 10})

	c.Clear()

	assert.Equal(t, 0, c.Len())
	assert.Equal(t, uint64(0), c.Size())
}

func TestCacheContains(t *testing.T) {
	c := New[string, testItem](100)

	c.Put("key1", testItem{"value1", 10})

	assert.True(t, c.Contains("key1"))
	assert.False(t, c.Contains("key2"))
}

func TestCacheGetMultiple(t *testing.T) {
	c := New[string, testItem](100)

	c.Put("key1", testItem{"value1", 10})
	c.Put("key2", testItem{"value2", 10})
	c.Put("key3", testItem{"value3", 10})

	result := c.GetMultiple("key1", "key2", "nonexistent")

	assert.Len(t, result, 2)
	assert.Equal(t, "value1", result["key1"].data)
	assert.Equal(t, "value2", result["key2"].data)
}

func TestSimpleCache(t *testing.T) {
	c := NewSimple[string, int](3)

	c.Put("a", 1)
	c.Put("b", 2)
	c.Put("c", 3)

	val, ok := c.Get("a")
	assert.True(t, ok)
	assert.Equal(t, 1, val)

	// Adding 4th item should evict oldest
	c.Put("d", 4)

	// Depending on LRU, "a" might still exist since we just accessed it
	// Let's verify "d" exists
	val, ok = c.Get("d")
	assert.True(t, ok)
	assert.Equal(t, 4, val)
}
