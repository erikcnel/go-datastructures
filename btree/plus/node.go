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

package plus

func split[K Comparable[K]](tree *BTree[K], parent, child node[K]) node[K] {
	if !child.needsSplit(tree.nodeSize) {
		return parent
	}

	key, left, right := child.split()
	if parent == nil {
		in := newInternalNode[K](tree.nodeSize)
		in.keys = append(in.keys, key)
		in.nodes = append(in.nodes, left)
		in.nodes = append(in.nodes, right)
		return in
	}

	p := parent.(*inode[K])
	i := p.search(key)
	// we want to ensure if the children are leaves we set
	// the left node's left sibling to point to left
	if cr, ok := left.(*lnode[K]); ok {
		if i > 0 {
			p.nodes[i-1].(*lnode[K]).pointer = cr
		}
	}
	p.keys.insertAt(i, key)
	p.nodes[i] = left
	p.nodes.insertAt(i+1, right)

	return parent
}

type node[K Comparable[K]] interface {
	insert(tree *BTree[K], key K) bool
	needsSplit(nodeSize uint64) bool
	// key is the median key while left and right nodes
	// represent the left and right nodes respectively
	split() (K, node[K], node[K])
	search(key K) int
	find(key K) *iterator[K]
}

type nodes[K Comparable[K]] []node[K]

func (nodes *nodes[K]) insertAt(i int, n node[K]) {
	if i == len(*nodes) {
		*nodes = append(*nodes, n)
		return
	}

	*nodes = append(*nodes, nil)
	copy((*nodes)[i+1:], (*nodes)[i:])
	(*nodes)[i] = n
}

func (ns nodes[K]) splitAt(i int) (nodes[K], nodes[K]) {
	left := make(nodes[K], i, cap(ns))
	right := make(nodes[K], len(ns)-i, cap(ns))
	copy(left, ns[:i])
	copy(right, ns[i:])
	return left, right
}

type inode[K Comparable[K]] struct {
	keys  keySlice[K]
	nodes nodes[K]
}

func (n *inode[K]) search(key K) int {
	return n.keys.search(key)
}

func (n *inode[K]) find(key K) *iterator[K] {
	i := n.search(key)
	if i == len(n.keys) {
		return n.nodes[len(n.nodes)-1].find(key)
	}

	found := n.keys[i]
	switch found.Compare(key) {
	case 0, 1:
		return n.nodes[i+1].find(key)
	default:
		return n.nodes[i].find(key)
	}
}

func (n *inode[K]) insert(tree *BTree[K], key K) bool {
	i := n.search(key)
	var child node[K]
	if i == len(n.keys) { // we want the last child node in this case
		child = n.nodes[len(n.nodes)-1]
	} else {
		match := n.keys[i]
		switch match.Compare(key) {
		case 1, 0:
			child = n.nodes[i+1]
		default:
			child = n.nodes[i]
		}
	}

	result := child.insert(tree, key)
	if !result { // no change of state occurred
		return result
	}

	if child.needsSplit(tree.nodeSize) {
		split(tree, n, child)
	}

	return result
}

func (n *inode[K]) needsSplit(nodeSize uint64) bool {
	return uint64(len(n.keys)) >= nodeSize
}

func (n *inode[K]) split() (K, node[K], node[K]) {
	if len(n.keys) < 3 {
		var zero K
		return zero, nil, nil
	}

	i := len(n.keys) / 2
	key := n.keys[i]

	ourKeys := make(keySlice[K], len(n.keys)-i-1, cap(n.keys))
	otherKeys := make(keySlice[K], i, cap(n.keys))
	copy(ourKeys, n.keys[i+1:])
	copy(otherKeys, n.keys[:i])
	left, right := n.nodes.splitAt(i + 1)
	otherNode := &inode[K]{
		keys:  otherKeys,
		nodes: left,
	}
	n.keys = ourKeys
	n.nodes = right
	return key, otherNode, n
}

func newInternalNode[K Comparable[K]](size uint64) *inode[K] {
	return &inode[K]{
		keys:  make(keySlice[K], 0, size),
		nodes: make(nodes[K], 0, size+1),
	}
}

type lnode[K Comparable[K]] struct {
	// points to the left leaf node is there is one
	pointer *lnode[K]
	keys    keySlice[K]
}

func (n *lnode[K]) search(key K) int {
	return n.keys.search(key)
}

func (n *lnode[K]) insert(tree *BTree[K], key K) bool {
	i := keySearch(n.keys, key)
	var inserted bool
	if i == len(n.keys) { // simple append will do
		n.keys = append(n.keys, key)
		inserted = true
	} else {
		if n.keys[i].Compare(key) == 0 {
			n.keys[i] = key
		} else {
			n.keys.insertAt(i, key)
			inserted = true
		}
	}

	if !inserted {
		return false
	}

	return true
}

func (n *lnode[K]) find(key K) *iterator[K] {
	i := n.search(key)
	if i == len(n.keys) {
		if n.pointer == nil {
			return nilIterator[K]()
		}

		return &iterator[K]{
			node:  n.pointer,
			index: -1,
		}
	}

	iter := &iterator[K]{
		node:  n,
		index: i - 1,
	}
	return iter
}

func (n *lnode[K]) split() (K, node[K], node[K]) {
	if len(n.keys) < 2 {
		var zero K
		return zero, nil, nil
	}
	i := len(n.keys) / 2
	key := n.keys[i]
	otherKeys := make(keySlice[K], i, cap(n.keys))
	ourKeys := make(keySlice[K], len(n.keys)-i, cap(n.keys))
	// we perform these copies so these slices don't all end up
	// pointing to the same underlying array which may make
	// for some very difficult to debug situations later.
	copy(otherKeys, n.keys[:i])
	copy(ourKeys, n.keys[i:])

	// this should release the original array for GC
	n.keys = ourKeys
	otherNode := &lnode[K]{
		keys:    otherKeys,
		pointer: n,
	}
	return key, otherNode, n
}

func (n *lnode[K]) needsSplit(nodeSize uint64) bool {
	return uint64(len(n.keys)) >= nodeSize
}

func newLeafNode[K Comparable[K]](size uint64) *lnode[K] {
	return &lnode[K]{
		keys: make(keySlice[K], 0, size),
	}
}

type keySlice[K Comparable[K]] []K

func (ks keySlice[K]) search(key K) int {
	return keySearch(ks, key)
}

func (ks *keySlice[K]) insertAt(i int, key K) {
	if i == len(*ks) {
		*ks = append(*ks, key)
		return
	}

	var zero K
	*ks = append(*ks, zero)
	copy((*ks)[i+1:], (*ks)[i:])
	(*ks)[i] = key
}

func (ks keySlice[K]) reverse() {
	for i := 0; i < len(ks)/2; i++ {
		ks[i], ks[len(ks)-i-1] = ks[len(ks)-i-1], ks[i]
	}
}
