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
Package btree/plus implements the ubiquitous B+ tree. As of this writing,
the tree is not quite finished. The delete-node merge functionality needs
to be added. There are also some performance improvements that can be
made, with some possible concurrency mechanisms.

This is a mutable b-tree so it is not threadsafe.

Performance characteristics:
Space: O(n)
Insert: O(log n)
Search: O(log n)

Example usage with generics:

	type MyInt int

	func (m MyInt) Compare(other MyInt) int {
		return int(m - other)
	}

	tree := plus.New[MyInt](32)
	tree.Insert(MyInt(5), MyInt(3), MyInt(7))
	results, found := tree.Get(MyInt(5))

BenchmarkIteration-8	   	10000	   		 	109347 ns/op
BenchmarkInsert-8	 		3000000	       		608 ns/op
BenchmarkGet-8	 			3000000	       		627 ns/op
*/
package plus

func keySearch[K Comparable[K]](keys keySlice[K], key K) int {
	low, high := 0, len(keys)-1
	var mid int
	for low <= high {
		mid = (high + low) / 2
		switch keys[mid].Compare(key) {
		case 1:
			low = mid + 1
		case -1:
			high = mid - 1
		case 0:
			return mid
		}
	}
	return low
}

// BTree is a generic B+ tree implementation.
type BTree[K Comparable[K]] struct {
	root             node[K]
	nodeSize, number uint64
}

func (tree *BTree[K]) insert(key K) {
	if tree.root == nil {
		n := newLeafNode[K](tree.nodeSize)
		n.insert(tree, key)
		tree.number = 1
		return
	}

	result := tree.root.insert(tree, key)
	if result {
		tree.number++
	}

	if tree.root.needsSplit(tree.nodeSize) {
		tree.root = split(tree, nil, tree.root)
	}
}

// Insert will insert the provided keys into the btree. This is an
// O(m*log n) operation where m is the number of keys to be inserted
// and n is the number of items in the tree.
func (tree *BTree[K]) Insert(keys ...K) {
	for _, key := range keys {
		tree.insert(key)
	}
}

// Iter returns an iterator that can be used to traverse the b-tree
// starting from the specified key or its successor.
func (tree *BTree[K]) Iter(key K) Iterator[K] {
	if tree.root == nil {
		return nilIterator[K]()
	}

	return tree.root.find(key)
}

func (tree *BTree[K]) get(key K) (K, bool) {
	iter := tree.root.find(key)
	if !iter.Next() {
		var zero K
		return zero, false
	}

	if iter.Value().Compare(key) == 0 {
		return iter.Value(), true
	}

	var zero K
	return zero, false
}

// Get will retrieve any keys matching the provided keys in the tree.
// Returns the found values and a parallel slice of bools indicating
// if each key was found.
func (tree *BTree[K]) Get(keys ...K) ([]K, []bool) {
	results := make([]K, len(keys))
	found := make([]bool, len(keys))
	for i, k := range keys {
		results[i], found[i] = tree.get(k)
	}

	return results, found
}

// Len returns the number of items in this tree.
func (tree *BTree[K]) Len() uint64 {
	return tree.number
}

// New creates a new B+ tree with the specified node size.
// The node size determines how many keys each node can hold.
func New[K Comparable[K]](nodeSize uint64) *BTree[K] {
	return &BTree[K]{
		nodeSize: nodeSize,
		root:     newLeafNode[K](nodeSize),
	}
}

// Deprecated: Use New[T] instead.
// btree is the old non-generic type kept for backward compatibility.
type btree = BTree[Key]

// Deprecated: Use New[T] instead.
func newBTree(nodeSize uint64) *btree {
	return New[Key](nodeSize)
}

// Deprecated type aliases for backward compatibility
type keys = keySlice[Key]
