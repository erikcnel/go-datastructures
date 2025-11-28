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

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func constructMockPayloads(num int) keySlice[*mockKey] {
	keys := make(keySlice[*mockKey], 0, num)
	for i := range num {
		keys = append(keys, newMockKey(i))
	}

	return keys
}

func constructMockNodes(num int) nodes[*mockKey] {
	ns := make(nodes[*mockKey], 0, num)
	for i := range num {
		keys := make(keySlice[*mockKey], 0, num)
		for j := range num {
			keys = append(keys, newMockKey(j*i+j))
		}

		n := &lnode[*mockKey]{
			keys: keys,
		}
		ns = append(ns, n)
		if i > 0 {
			ns[i-1].(*lnode[*mockKey]).pointer = n
		}
	}

	return ns
}

func constructMockInternalNode(ns nodes[*mockKey]) *inode[*mockKey] {
	if len(ns) < 2 {
		return nil
	}

	keys := make(keySlice[*mockKey], 0, len(ns)-1)
	for i := 1; i < len(ns); i++ {
		keys = append(keys, ns[i].(*lnode[*mockKey]).keys[0])
	}

	in := &inode[*mockKey]{
		keys:  keys,
		nodes: ns,
	}
	return in
}

func TestLeafNodeInsert(t *testing.T) {
	tree := New[*mockKey](3)
	n := newLeafNode[*mockKey](3)
	key := newMockKey(3)

	n.insert(tree, key)

	assert.Len(t, n.keys, 1)
	assert.Nil(t, n.pointer)
	assert.Equal(t, n.keys[0], key)
	assert.Equal(t, 0, n.keys[0].Compare(key))
}

func TestDuplicateLeafNodeInsert(t *testing.T) {
	tree := New[*mockKey](3)
	n := newLeafNode[*mockKey](3)
	k1 := newMockKey(3)
	k2 := newMockKey(3)

	assert.True(t, n.insert(tree, k1))
	assert.False(t, n.insert(tree, k2))
	assert.False(t, n.insert(tree, k1))
}

func TestMultipleLeafNodeInsert(t *testing.T) {
	tree := New[*mockKey](3)
	n := newLeafNode[*mockKey](3)

	k1 := newMockKey(3)
	k2 := newMockKey(4)

	assert.True(t, n.insert(tree, k1))
	n.insert(tree, k2)

	if !assert.Len(t, n.keys, 2) {
		return
	}
	assert.Nil(t, n.pointer)
	assert.Equal(t, k1, n.keys[0])
	assert.Equal(t, k2, n.keys[1])
}

func TestLeafNodeSplitEvenNumber(t *testing.T) {
	keys := constructMockPayloads(4)

	node := &lnode[*mockKey]{
		keys: keys,
	}

	key, left, right := node.split()
	assert.Equal(t, keys[2], key)
	assert.Equal(t, left.(*lnode[*mockKey]).keys, keys[:2])
	assert.Equal(t, right.(*lnode[*mockKey]).keys, keys[2:])
	assert.Equal(t, left.(*lnode[*mockKey]).pointer, right)
}

func TestLeafNodeSplitOddNumber(t *testing.T) {
	keys := constructMockPayloads(3)

	node := &lnode[*mockKey]{
		keys: keys,
	}

	key, left, right := node.split()
	assert.Equal(t, keys[1], key)
	assert.Equal(t, left.(*lnode[*mockKey]).keys, keys[:1])
	assert.Equal(t, right.(*lnode[*mockKey]).keys, keys[1:])
	assert.Equal(t, left.(*lnode[*mockKey]).pointer, right)
}

func TestTwoKeysLeafNodeSplit(t *testing.T) {
	keys := constructMockPayloads(2)

	node := &lnode[*mockKey]{
		keys: keys,
	}

	key, left, right := node.split()
	assert.Equal(t, keys[1], key)
	assert.Equal(t, left.(*lnode[*mockKey]).keys, keys[:1])
	assert.Equal(t, right.(*lnode[*mockKey]).keys, keys[1:])
	assert.Equal(t, left.(*lnode[*mockKey]).pointer, right)
}

func TestLessThanTwoKeysSplit(t *testing.T) {
	keys := constructMockPayloads(1)

	node := &lnode[*mockKey]{
		keys: keys,
	}

	key, left, right := node.split()
	assert.Nil(t, key)
	assert.Nil(t, left)
	assert.Nil(t, right)
}

func TestInternalNodeSplit2_3_4(t *testing.T) {
	ns := constructMockNodes(4)
	in := constructMockInternalNode(ns)

	key, left, right := in.split()
	assert.Equal(t, ns[3].(*lnode[*mockKey]).keys[0], key)
	assert.Len(t, left.(*inode[*mockKey]).keys, 1)
	assert.Len(t, right.(*inode[*mockKey]).keys, 1)
	assert.Equal(t, ns[:2], left.(*inode[*mockKey]).nodes)
	assert.Equal(t, ns[2:], right.(*inode[*mockKey]).nodes)
}

func TestInternalNodeSplit3_4_5(t *testing.T) {
	ns := constructMockNodes(5)
	in := constructMockInternalNode(ns)

	key, left, right := in.split()
	assert.Equal(t, ns[4].(*lnode[*mockKey]).keys[0], key)
	assert.Len(t, left.(*inode[*mockKey]).keys, 2)
	assert.Len(t, right.(*inode[*mockKey]).keys, 1)
	assert.Equal(t, ns[:3], left.(*inode[*mockKey]).nodes)
	assert.Equal(t, ns[3:], right.(*inode[*mockKey]).nodes)
}

func TestInternalNodeLessThan3Keys(t *testing.T) {
	ns := constructMockNodes(2)
	in := constructMockInternalNode(ns)

	key, left, right := in.split()
	assert.Nil(t, key)
	assert.Nil(t, left)
	assert.Nil(t, right)
}
