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
Package skip defines a skiplist datastructure. That is, a data structure
that probabilistically determines relationships between keys. By doing
so, it becomes easier to program than a binary search tree but maintains
similar speeds.

Performance characteristics:
Insert: O(log n)
Search: O(log n)
Delete: O(log n)
Space: O(n)

Recently added is the capability to address, insert, and replace an
entry by position. This capability is achieved by saving the width
of the "gap" between two nodes. Searching for an item by position is
very similar to searching by value in that the same basic algorithm is
used but we are searching for width instead of value. Because this avoids
the overhead associated with Golang interfaces, operations by position
are about twice as fast as operations by value. Time complexities listed
below.

SearchByPosition: O(log n)
InsertByPosition: O(log n)

More information here: http://cglab.ca/~morin/teaching/5408/refs/p90b.pdf

Example usage with generics:

	type MyInt int

	func (m MyInt) Compare(other MyInt) int {
		return int(m - other)
	}

	sl := skip.New[MyInt](uint64(0))
	sl.Insert(MyInt(5), MyInt(3), MyInt(7))
	results, found := sl.Get(MyInt(5))

Benchmarks:
BenchmarkInsert-8	 		 2000000	       930 ns/op
BenchmarkGet-8	 			 2000000	       989 ns/op
BenchmarkDelete-8	 		 3000000	       600 ns/op
BenchmarkPrepend-8	 		 1000000	      1468 ns/op
BenchmarkByPosition-8		10000000	       202 ns/op
BenchmarkInsertAtPosition-8	 3000000	       485 ns/op

CPU profiling has shown that the most expensive thing we do here
is call Compare. With generics, we can now avoid the overhead of
interface dispatch and get better performance with concrete types.
*/
package skip

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

const p = .5 // the p level defines the probability that a node
// with a value at level i also has a value at i+1. This number
// is also important in determining max level. Max level will
// be defined as L(N) where L = log base (1/p) of n where n
// is the number of items in the list and N is the number of possible
// items in the universe. If p = .5 then maxlevel = 32 is appropriate
// for uint32.

// lockedSource is an implementation of rand.Source that is safe for
// concurrent use by multiple goroutines. The code is modeled after
// https://golang.org/src/math/rand/rand.go.
type lockedSource struct {
	mu  sync.Mutex
	src rand.Source
}

// Int63 implements the rand.Source interface.
func (ls *lockedSource) Int63() (n int64) {
	ls.mu.Lock()
	n = ls.src.Int63()
	ls.mu.Unlock()
	return
}

// Seed implements the rand.Source interface.
func (ls *lockedSource) Seed(seed int64) {
	ls.mu.Lock()
	ls.src.Seed(seed)
	ls.mu.Unlock()
}

// generator will be the common generator to create random numbers. It
// is seeded with unix nanosecond when this line is executed at runtime,
// and only executed once ensuring all random numbers come from the same
// randomly seeded generator.
var generator = rand.New(&lockedSource{src: rand.NewSource(time.Now().UnixNano())})

func generateLevel(maxLevel uint8) uint8 {
	var level uint8
	for level = uint8(1); level < maxLevel-1; level++ {
		if generator.Float64() >= p {
			return level
		}
	}

	return level
}

func insertNode[T Comparable[T]](sl *SkipList[T], n *node[T], cmp T, pos uint64, cache nodes[T], posCache widths, allowDuplicate bool) (T, bool) {
	var zero T
	if !allowDuplicate && n != nil && n.hasEntry && n.Compare(cmp) == 0 { // a simple update in this case
		oldEntry := n.entry
		n.entry = cmp
		return oldEntry, true
	}
	atomic.AddUint64(&sl.num, 1)

	nodeLevel := generateLevel(sl.maxLevel)
	if nodeLevel > sl.level {
		for i := sl.level; i < nodeLevel; i++ {
			cache[i] = sl.head
		}
		sl.level = nodeLevel
	}

	nn := newNode(cmp, true, nodeLevel)
	for i := range nodeLevel {
		nn.forward[i] = cache[i].forward[i]
		cache[i].forward[i] = nn
		formerWidth := cache[i].widths[i]
		if formerWidth == 0 {
			nn.widths[i] = 0
		} else {
			nn.widths[i] = posCache[i] + formerWidth + 1 - pos
		}

		if cache[i].forward[i] != nil {
			cache[i].widths[i] = pos - posCache[i]
		}
	}

	for i := nodeLevel; i < sl.level; i++ {
		if cache[i].forward[i] == nil {
			continue
		}
		cache[i].widths[i]++
	}
	return zero, false
}

func splitAt[T Comparable[T]](sl *SkipList[T], index uint64) (*SkipList[T], *SkipList[T]) {
	var zero T
	right := &SkipList[T]{}
	right.maxLevel = sl.maxLevel
	right.level = sl.level
	right.cache = make(nodes[T], sl.maxLevel)
	right.posCache = make(widths, sl.maxLevel)
	right.head = newNode(zero, false, sl.maxLevel)
	sl.searchByPosition(index, sl.cache, sl.posCache) // populate the cache that needs updating

	for i := uint8(0); i <= sl.level; i++ {
		right.head.forward[i] = sl.cache[i].forward[i]
		if sl.cache[i].forward[i] != nil {
			right.head.widths[i] = sl.cache[i].widths[i] - (index - sl.posCache[i])
		}
		sl.cache[i].widths[i] = 0
		sl.cache[i].forward[i] = nil
	}

	right.num = sl.Len() - index // right is not in user's hands yet
	atomic.AddUint64(&sl.num, -right.num)

	sl.resetMaxLevel()
	right.resetMaxLevel()

	return sl, right
}

// SkipList is a datastructure that probabalistically determines
// relationships between nodes. This results in a structure
// that performs similarly to a BST but is much easier to build
// from a programmatic perspective (no rotations).
type SkipList[T Comparable[T]] struct {
	maxLevel, level uint8
	head            *node[T]
	num             uint64
	// a list of nodes that can be reused, should reduce
	// the number of allocations in the insert/delete case.
	cache    nodes[T]
	posCache widths
}

// init will initialize this skiplist. The parameter is expected
// to be of some uint type which will set this skiplist's maximum
// level.
func (sl *SkipList[T]) init(ifc any) {
	switch ifc.(type) {
	case uint8:
		sl.maxLevel = 8
	case uint16:
		sl.maxLevel = 16
	case uint32:
		sl.maxLevel = 32
	case uint64, uint:
		sl.maxLevel = 64
	}
	var zero T
	sl.cache = make(nodes[T], sl.maxLevel)
	sl.posCache = make(widths, sl.maxLevel)
	sl.head = newNode(zero, false, sl.maxLevel)
}

func (sl *SkipList[T]) search(cmp T, update nodes[T], widthCache widths) (*node[T], uint64) {
	if sl.Len() == 0 { // nothing in the list
		return nil, 1
	}

	var pos uint64 = 0
	var offset uint8
	var alreadyChecked *node[T]
	n := sl.head
	for i := uint8(0); i <= sl.level; i++ {
		offset = sl.level - i
		for n.forward[offset] != nil && n.forward[offset] != alreadyChecked && n.forward[offset].hasEntry && n.forward[offset].Compare(cmp) < 0 {
			pos += n.widths[offset]
			n = n.forward[offset]
		}

		alreadyChecked = n
		if update != nil {
			update[offset] = n
			widthCache[offset] = pos
		}
	}

	return n.forward[0], pos + 1
}

func (sl *SkipList[T]) resetMaxLevel() {
	if sl.level < 1 {
		sl.level = 1
		return
	}
	for sl.head.forward[sl.level-1] == nil && sl.level > 1 {
		sl.level--
	}
}

func (sl *SkipList[T]) searchByPosition(position uint64, update nodes[T], widthCache widths) (*node[T], uint64) {
	if sl.Len() == 0 { // nothing in the list
		return nil, 1
	}

	if position > sl.Len() {
		return nil, 1
	}

	var pos uint64 = 0
	var offset uint8
	n := sl.head
	for i := uint8(0); i <= sl.level; i++ {
		offset = sl.level - i
		for n.forward[offset] != nil && pos+n.widths[offset] <= position {
			pos += n.widths[offset]
			n = n.forward[offset]
		}

		if update != nil {
			update[offset] = n
			widthCache[offset] = pos
		}
	}

	return n, pos + 1
}

// Get will retrieve values associated with the keys provided. Returns
// the found values and a parallel slice of bools indicating if each was found.
// This is an O(log n) operation.
func (sl *SkipList[T]) Get(comparators ...T) ([]T, []bool) {
	results := make([]T, len(comparators))
	found := make([]bool, len(comparators))

	var n *node[T]
	for i, cmp := range comparators {
		n, _ = sl.search(cmp, nil, nil)
		if n != nil && n.hasEntry && n.Compare(cmp) == 0 {
			results[i] = n.entry
			found[i] = true
		}
	}

	return results, found
}

// GetWithPosition will retrieve the value with the provided key and
// return the position of that value within the list. Returns (zero, 0, false)
// if an associated value could not be found.
func (sl *SkipList[T]) GetWithPosition(cmp T) (T, uint64, bool) {
	n, pos := sl.search(cmp, nil, nil)
	if n == nil || !n.hasEntry {
		var zero T
		return zero, 0, false
	}

	return n.entry, pos - 1, true
}

// ByPosition returns the value at the given position.
// Returns (zero, false) if position is out of bounds.
func (sl *SkipList[T]) ByPosition(position uint64) (T, bool) {
	n, _ := sl.searchByPosition(position+1, nil, nil)
	if n == nil || !n.hasEntry {
		var zero T
		return zero, false
	}

	return n.entry, true
}

func (sl *SkipList[T]) insert(cmp T) (T, bool) {
	n, pos := sl.search(cmp, sl.cache, sl.posCache)
	return insertNode(sl, n, cmp, pos, sl.cache, sl.posCache, false)
}

// Insert will insert the provided comparators into the list. Returns
// the overwritten values and bools indicating if each was overwritten.
// This is expected to be an O(log n) operation.
func (sl *SkipList[T]) Insert(comparators ...T) ([]T, []bool) {
	overwritten := make([]T, len(comparators))
	wasOverwritten := make([]bool, len(comparators))
	for i, cmp := range comparators {
		overwritten[i], wasOverwritten[i] = sl.insert(cmp)
	}

	return overwritten, wasOverwritten
}

func (sl *SkipList[T]) insertAtPosition(position uint64, cmp T) {
	if position > sl.Len() {
		position = sl.Len()
	}
	n, pos := sl.searchByPosition(position, sl.cache, sl.posCache)
	insertNode(sl, n, cmp, pos, sl.cache, sl.posCache, true)
}

// InsertAtPosition will insert the provided value at the provided position.
// If position is greater than the length of the skiplist, the value
// is appended. This method bypasses order checks and checks for
// duplicates so use with caution.
func (sl *SkipList[T]) InsertAtPosition(position uint64, cmp T) {
	sl.insertAtPosition(position, cmp)
}

func (sl *SkipList[T]) replaceAtPosition(position uint64, cmp T) {
	n, _ := sl.searchByPosition(position+1, nil, nil)
	if n == nil || !n.hasEntry {
		return
	}

	n.entry = cmp
}

// ReplaceAtPosition will replace the value at the provided position
// with the provided value. If the provided position does not exist,
// this operation is a no-op.
func (sl *SkipList[T]) ReplaceAtPosition(position uint64, cmp T) {
	sl.replaceAtPosition(position, cmp)
}

func (sl *SkipList[T]) delete(cmp T) (T, bool) {
	n, _ := sl.search(cmp, sl.cache, sl.posCache)

	if n == nil || !n.hasEntry || n.Compare(cmp) != 0 {
		var zero T
		return zero, false
	}

	atomic.AddUint64(&sl.num, ^uint64(0)) // decrement

	for i := uint8(0); i <= sl.level; i++ {
		if sl.cache[i].forward[i] != n {
			if sl.cache[i].forward[i] != nil {
				sl.cache[i].widths[i]--
			}
			continue
		}

		sl.cache[i].widths[i] += n.widths[i] - 1
		sl.cache[i].forward[i] = n.forward[i]
	}

	for sl.level > 1 && sl.head.forward[sl.level-1] == nil {
		sl.head.widths[sl.level] = 0
		sl.level--
	}

	return n.entry, true
}

// Delete will remove the provided keys from the skiplist and return
// the deleted values and bools indicating if each was deleted.
// This is a no-op if an associated key could not be found.
// This is an O(log n) operation.
func (sl *SkipList[T]) Delete(comparators ...T) ([]T, []bool) {
	deleted := make([]T, len(comparators))
	wasDeleted := make([]bool, len(comparators))

	for i, cmp := range comparators {
		deleted[i], wasDeleted[i] = sl.delete(cmp)
	}

	return deleted, wasDeleted
}

// Len returns the number of items in this skiplist.
func (sl *SkipList[T]) Len() uint64 {
	return atomic.LoadUint64(&sl.num)
}

func (sl *SkipList[T]) iterAtPosition(pos uint64) *iterator[T] {
	n, _ := sl.searchByPosition(pos, nil, nil)
	if n == nil || !n.hasEntry {
		return nilIterator[T]()
	}

	return &iterator[T]{
		first: true,
		n:     n,
	}
}

// IterAtPosition is the sister method to Iter only the user defines
// a position in the skiplist to begin iteration instead of a value.
func (sl *SkipList[T]) IterAtPosition(pos uint64) Iterator[T] {
	return sl.iterAtPosition(pos + 1)
}

func (sl *SkipList[T]) iter(cmp T) *iterator[T] {
	n, _ := sl.search(cmp, nil, nil)
	if n == nil || !n.hasEntry {
		return nilIterator[T]()
	}

	return &iterator[T]{
		first: true,
		n:     n,
	}
}

// Iter will return an iterator that can be used to iterate
// over all the values with a key equal to or greater than
// the key provided.
func (sl *SkipList[T]) Iter(cmp T) Iterator[T] {
	return sl.iter(cmp)
}

// SplitAt will split the current skiplist into two lists. The first
// skiplist returned is the "left" list and the second is the "right."
// The index defines the last item in the left list. If index is greater
// then the length of this list, only the left skiplist is returned
// and the right will be nil. This is a mutable operation and modifies
// the content of this list.
func (sl *SkipList[T]) SplitAt(index uint64) (*SkipList[T], *SkipList[T]) {
	index++ // 0-index offset
	if index >= sl.Len() {
		return sl, nil
	}
	return splitAt(sl, index)
}

// New will allocate, initialize, and return a new skiplist.
// The provided parameter should be of type uint and will determine
// the maximum possible level that will be created to ensure
// a random and quick distribution of levels. Parameter must
// be a uint type.
func New[T Comparable[T]](ifc any) *SkipList[T] {
	sl := &SkipList[T]{}
	sl.init(ifc)
	return sl
}
