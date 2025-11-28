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
Package avl includes an immutable AVL tree.

AVL trees can be used as the foundation for many functional data types.
Combined with a B+ tree, you can make an immutable index which serves as the
backbone for many different kinds of key/value stores.

Time complexities:
Space: O(n)
Insert: O(log n)
Delete: O(log n)
Get: O(log n)

The immutable version of the AVL tree is obviously going to be slower than
the mutable version but should offer higher read availability.

Example usage with generics:

	type MyInt int

	func (m MyInt) Compare(other MyInt) int {
		return int(m - other)
	}

	tree := avl.New[MyInt]()
	tree, _ = tree.Insert(MyInt(5), MyInt(3), MyInt(7))
	results := tree.Get(MyInt(5)) // returns [5]
*/
package avl

import "math"

// Immutable represents an immutable AVL tree. This is achieved
// by branch copying.
type Immutable[T Comparable[T]] struct {
	root   *node[T]
	number uint64
	dummy  node[T] // helper for inserts.
}

// copy returns a copy of this immutable tree with a copy
// of the root and a new dummy helper for the insertion operation.
func (immutable *Immutable[T]) copy() *Immutable[T] {
	var root *node[T]
	if immutable.root != nil {
		root = immutable.root.copy()
	}
	var zero T
	cp := &Immutable[T]{
		root:   root,
		number: immutable.number,
		dummy:  *newNode(zero, false),
	}
	return cp
}

func (immutable *Immutable[T]) resetDummy() {
	immutable.dummy.children[0], immutable.dummy.children[1] = nil, nil
	immutable.dummy.balance = 0
}

func (immutable *Immutable[T]) init() {
	immutable.dummy = node[T]{
		children: [2]*node[T]{},
	}
}

func (immutable *Immutable[T]) get(entry T) (T, bool) {
	n := immutable.root
	var result int
	for n != nil {
		switch result = n.entry.Compare(entry); {
		case result == 0:
			return n.entry, true
		case result > 0:
			n = n.children[0]
		case result < 0:
			n = n.children[1]
		}
	}

	var zero T
	return zero, false
}

// Get will get the provided entries from the tree. Returns the found entries
// and a parallel slice of bools indicating if each entry was found.
func (immutable *Immutable[T]) Get(entries ...T) ([]T, []bool) {
	results := make([]T, len(entries))
	found := make([]bool, len(entries))
	for i, e := range entries {
		results[i], found[i] = immutable.get(e)
	}
	return results, found
}

// Len returns the number of items in this immutable.
func (immutable *Immutable[T]) Len() uint64 {
	return immutable.number
}

func (immutable *Immutable[T]) insert(entry T) (T, bool) {
	var zero T
	if immutable.root == nil {
		immutable.root = newNode(entry, true)
		immutable.number++
		return zero, false
	}

	immutable.resetDummy()
	var (
		dummy           = immutable.dummy
		p, s, q         *node[T]
		dir, normalized int
		helper          = &dummy
	)

	// set this AFTER clearing dummy
	helper.children[1] = immutable.root
	// we'll go ahead and copy on the way down as we'll need to branch
	// copy no matter what.
	for s, p = helper.children[1], helper.children[1]; ; {
		dir = p.entry.Compare(entry)

		normalized = normalizeComparison(dir)
		if dir > 0 { // go left
			if p.children[0] != nil {
				q = p.children[0].copy()
				p.children[0] = q
			} else {
				q = nil
			}
		} else if dir < 0 { // go right
			if p.children[1] != nil {
				q = p.children[1].copy()
				p.children[1] = q
			} else {
				q = nil
			}
		} else { // equality
			oldEntry := p.entry
			p.entry = entry
			return oldEntry, true
		}
		if q == nil {
			break
		}

		if q.balance != 0 {
			helper = p
			s = q
		}
		p = q
	}

	immutable.number++
	q = newNode(entry, true)
	p.children[normalized] = q

	immutable.root = dummy.children[1]
	for p = s; p != q; p = p.children[normalized] {
		normalized = normalizeComparison(p.entry.Compare(entry))
		if normalized == 0 {
			p.balance--
		} else {
			p.balance++
		}
	}

	q = s

	if math.Abs(float64(s.balance)) > 1 {
		normalized = normalizeComparison(s.entry.Compare(entry))
		s = insertBalance(s, normalized)
	}

	if q == dummy.children[1] {
		immutable.root = s
	} else {
		helper.children[intFromBool(helper.children[1] == q)] = s
	}
	return zero, false
}

// Insert will add the provided entries into the tree and return the new
// state. Also returned is a list of entries that were overwritten and
// bools indicating if each was overwritten.
func (immutable *Immutable[T]) Insert(entries ...T) (*Immutable[T], []T, []bool) {
	if len(entries) == 0 {
		return immutable, nil, nil
	}

	overwritten := make([]T, len(entries))
	wasOverwritten := make([]bool, len(entries))
	cp := immutable.copy()
	for i, e := range entries {
		overwritten[i], wasOverwritten[i] = cp.insert(e)
	}

	return cp, overwritten, wasOverwritten
}

func (immutable *Immutable[T]) delete(entry T) (T, bool) {
	var zero T
	if immutable.root == nil {
		return zero, false
	}

	var (
		cache                      = make(nodes[T], 64)
		it, p, q                   *node[T]
		top, done, dir, normalized int
		dirs                       = make([]int, 64)
		oldEntry                   T
	)

	it = immutable.root

	for {
		if it == nil {
			return zero, false
		}

		dir = it.entry.Compare(entry)
		if dir == 0 {
			break
		}
		normalized = normalizeComparison(dir)
		dirs[top] = normalized
		cache[top] = it
		top++
		it = it.children[normalized]
	}
	immutable.number--
	oldEntry = it.entry

	// we need to make a branch copy now
	for i := 0; i < top; i++ {
		p = cache[i]
		if p.children[dirs[i]] != nil {
			q = p.children[dirs[i]].copy()
			p.children[dirs[i]] = q
			if i != top-1 {
				cache[i+1] = q
			}
		}
	}
	it = it.copy()

	oldTop := top
	if it.children[0] == nil || it.children[1] == nil {
		dir = intFromBool(it.children[0] == nil)
		if top != 0 {
			cache[top-1].children[dirs[top-1]] = it.children[dir]
		} else {
			immutable.root = it.children[dir]
		}
	} else {
		heir := it.children[1]
		dirs[top] = 1
		cache[top] = it
		top++

		for heir.children[0] != nil {
			dirs[top] = 0
			cache[top] = heir
			top++
			heir = heir.children[0]
		}

		it.entry = heir.entry
		if oldTop != 0 {
			cache[oldTop-1].children[dirs[oldTop-1]] = it
		} else {
			immutable.root = it
		}
		cache[top-1].children[intFromBool(cache[top-1] == it)] = heir.children[1]
	}

	for top-1 >= 0 && done == 0 {
		top--
		if dirs[top] != 0 {
			cache[top].balance--
		} else {
			cache[top].balance++
		}

		if math.Abs(float64(cache[top].balance)) == 1 {
			break
		} else if math.Abs(float64(cache[top].balance)) > 1 {
			cache[top] = removeBalance(cache[top], dirs[top], &done)

			if top != 0 {
				cache[top-1].children[dirs[top-1]] = cache[top]
			} else {
				immutable.root = cache[0]
			}
		}
	}

	return oldEntry, true
}

// Delete will remove the provided entries from this AVL tree and
// return a new tree and any entries removed. The bool slice indicates
// if each entry was found and deleted.
func (immutable *Immutable[T]) Delete(entries ...T) (*Immutable[T], []T, []bool) {
	if len(entries) == 0 {
		return immutable, nil, nil
	}

	deleted := make([]T, len(entries))
	wasDeleted := make([]bool, len(entries))
	cp := immutable.copy()
	for i, e := range entries {
		deleted[i], wasDeleted[i] = cp.delete(e)
	}

	return cp, deleted, wasDeleted
}

func insertBalance[T Comparable[T]](root *node[T], dir int) *node[T] {
	n := root.children[dir]
	var bal int8
	if dir == 0 {
		bal = -1
	} else {
		bal = 1
	}

	if n.balance == bal {
		root.balance, n.balance = 0, 0
		root = rotate(root, takeOpposite(dir))
	} else {
		adjustBalance(root, dir, int(bal))
		root = doubleRotate(root, takeOpposite(dir))
	}

	return root
}

func removeBalance[T Comparable[T]](root *node[T], dir int, done *int) *node[T] {
	n := root.children[takeOpposite(dir)].copy()
	root.children[takeOpposite(dir)] = n
	var bal int8
	if dir == 0 {
		bal = -1
	} else {
		bal = 1
	}

	if n.balance == -bal {
		root.balance, n.balance = 0, 0
		root = rotate(root, dir)
	} else if n.balance == bal {
		adjustBalance(root, takeOpposite(dir), int(-bal))
		root = doubleRotate(root, dir)
	} else {
		root.balance = -bal
		n.balance = bal
		root = rotate(root, dir)
		*done = 1
	}

	return root
}

func intFromBool(value bool) int {
	if value {
		return 1
	}
	return 0
}

func takeOpposite(value int) int {
	return 1 - value
}

func adjustBalance[T Comparable[T]](root *node[T], dir, bal int) {
	n := root.children[dir]
	nn := n.children[takeOpposite(dir)]

	if nn.balance == 0 {
		root.balance, n.balance = 0, 0
	} else if int(nn.balance) == bal {
		root.balance = int8(-bal)
		n.balance = 0
	} else {
		root.balance = 0
		n.balance = int8(bal)
	}
	nn.balance = 0
}

func rotate[T Comparable[T]](parent *node[T], dir int) *node[T] {
	otherDir := takeOpposite(dir)

	child := parent.children[otherDir]
	parent.children[otherDir] = child.children[dir]
	child.children[dir] = parent

	return child
}

func doubleRotate[T Comparable[T]](parent *node[T], dir int) *node[T] {
	otherDir := takeOpposite(dir)

	parent.children[otherDir] = rotate(parent.children[otherDir], otherDir)
	return rotate(parent, dir)
}

// normalizeComparison converts the value returned from Compare
// into a direction, ie, left or right, 0 or 1.
func normalizeComparison(i int) int {
	if i < 0 {
		return 1
	}
	if i > 0 {
		return 0
	}
	return -1
}

// New allocates, initializes, and returns a new immutable AVL tree.
func New[T Comparable[T]]() *Immutable[T] {
	immutable := &Immutable[T]{}
	immutable.init()
	return immutable
}

// NewImmutable is kept for backward compatibility.
// Deprecated: Use New[T]() instead.
func NewImmutable() *Immutable[Entry] {
	return New[Entry]()
}
