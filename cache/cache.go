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

/*
Package cache provides a generic bounded-size in-memory cache with
configurable eviction policies.

Example usage:

	cache := cache.New[string, MyItem](1024 * 1024) // 1MB capacity

	cache.Put("key1", myItem)
	item, ok := cache.Get("key1")
*/
package cache

import (
	"container/list"
	"sync"
)

// Sized is an interface for items that have a size.
type Sized interface {
	// Size returns the item's size in bytes.
	Size() uint64
}

// Policy is a cache eviction policy.
type Policy uint8

const (
	// LeastRecentlyAdded evicts items in the order they were added.
	LeastRecentlyAdded Policy = iota
	// LeastRecentlyUsed evicts items that haven't been accessed recently.
	LeastRecentlyUsed
)

// Option configures a cache.
type Option[K comparable, V Sized] func(*Cache[K, V])

// WithPolicy sets the eviction policy.
// Default is LeastRecentlyUsed.
func WithPolicy[K comparable, V Sized](policy Policy) Option[K, V] {
	return func(c *Cache[K, V]) {
		c.policy = policy
	}
}

// cached wraps an item with its eviction list element.
type cached[V Sized] struct {
	item    V
	element *list.Element
}

// Cache is a generic bounded-size in-memory cache.
// K must be comparable (usable as map key), V must implement Sized.
type Cache[K comparable, V Sized] struct {
	sync.RWMutex
	capacity uint64
	size     uint64
	items    map[K]*cached[V]
	keyList  *list.List
	policy   Policy
}

// New creates a new cache with the given capacity in bytes.
func New[K comparable, V Sized](capacity uint64, options ...Option[K, V]) *Cache[K, V] {
	c := &Cache[K, V]{
		capacity: capacity,
		items:    make(map[K]*cached[V]),
		keyList:  list.New(),
		policy:   LeastRecentlyUsed,
	}

	for _, opt := range options {
		opt(c)
	}

	return c
}

// Get retrieves an item from the cache.
// Returns the item and true if found, zero value and false otherwise.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.Lock()
	defer c.Unlock()

	cached, ok := c.items[key]
	if !ok {
		var zero V
		return zero, false
	}

	if c.policy == LeastRecentlyUsed {
		c.keyList.MoveToFront(cached.element)
	}

	return cached.item, true
}

// GetMultiple retrieves multiple items from the cache.
// Returns a map of found items.
func (c *Cache[K, V]) GetMultiple(keys ...K) map[K]V {
	c.Lock()
	defer c.Unlock()

	result := make(map[K]V, len(keys))
	for _, key := range keys {
		if cached, ok := c.items[key]; ok {
			if c.policy == LeastRecentlyUsed {
				c.keyList.MoveToFront(cached.element)
			}
			result[key] = cached.item
		}
	}
	return result
}

// Put adds an item to the cache, evicting items if necessary to make room.
func (c *Cache[K, V]) Put(key K, item V) {
	c.Lock()
	defer c.Unlock()

	// Remove existing item with this key
	c.removeUnlocked(key)

	// Ensure capacity
	c.ensureCapacity(item.Size())

	// Add new item
	element := c.keyList.PushFront(key)
	c.items[key] = &cached[V]{
		item:    item,
		element: element,
	}
	c.size += item.Size()
}

// Remove removes items with the given keys from the cache.
func (c *Cache[K, V]) Remove(keys ...K) {
	c.Lock()
	defer c.Unlock()

	for _, key := range keys {
		c.removeUnlocked(key)
	}
}

// Size returns the current size of all items in the cache.
func (c *Cache[K, V]) Size() uint64 {
	c.RLock()
	defer c.RUnlock()
	return c.size
}

// Len returns the number of items in the cache.
func (c *Cache[K, V]) Len() int {
	c.RLock()
	defer c.RUnlock()
	return len(c.items)
}

// Clear removes all items from the cache.
func (c *Cache[K, V]) Clear() {
	c.Lock()
	defer c.Unlock()

	c.items = make(map[K]*cached[V])
	c.keyList = list.New()
	c.size = 0
}

// Contains returns true if the key exists in the cache.
func (c *Cache[K, V]) Contains(key K) bool {
	c.RLock()
	defer c.RUnlock()
	_, ok := c.items[key]
	return ok
}

// Keys returns all keys currently in the cache.
func (c *Cache[K, V]) Keys() []K {
	c.RLock()
	defer c.RUnlock()

	keys := make([]K, 0, len(c.items))
	for key := range c.items {
		keys = append(keys, key)
	}
	return keys
}

// ensureCapacity evicts items until there's room for the given size.
// Caller must hold the lock.
func (c *Cache[K, V]) ensureCapacity(toAdd uint64) {
	mustRemove := int64(c.size+toAdd) - int64(c.capacity)
	for mustRemove > 0 && c.keyList.Len() > 0 {
		element := c.keyList.Back()
		key := element.Value.(K)
		if cached, ok := c.items[key]; ok {
			mustRemove -= int64(cached.item.Size())
			c.removeUnlocked(key)
		}
	}
}

// removeUnlocked removes an item without acquiring the lock.
// Caller must hold the lock.
func (c *Cache[K, V]) removeUnlocked(key K) {
	if cached, ok := c.items[key]; ok {
		delete(c.items, key)
		c.size -= cached.item.Size()
		if cached.element != nil {
			c.keyList.Remove(cached.element)
		}
	}
}

// SimpleCache is a cache for items where each item counts as size 1.
// This is useful when you want to limit by count rather than bytes.
type SimpleCache[K comparable, V any] struct {
	*Cache[K, sizedWrapper[V]]
}

type sizedWrapper[V any] struct {
	value V
}

func (s sizedWrapper[V]) Size() uint64 {
	return 1
}

// NewSimple creates a cache that limits by item count rather than bytes.
func NewSimple[K comparable, V any](maxItems uint64) *SimpleCache[K, V] {
	return &SimpleCache[K, V]{
		Cache: New[K, sizedWrapper[V]](maxItems),
	}
}

// Get retrieves an item from the cache.
func (c *SimpleCache[K, V]) Get(key K) (V, bool) {
	wrapper, ok := c.Cache.Get(key)
	if !ok {
		var zero V
		return zero, false
	}
	return wrapper.value, true
}

// Put adds an item to the cache.
func (c *SimpleCache[K, V]) Put(key K, value V) {
	c.Cache.Put(key, sizedWrapper[V]{value: value})
}
