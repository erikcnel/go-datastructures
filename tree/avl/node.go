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

type nodes[T Comparable[T]] []*node[T]

func (ns nodes[T]) reset() {
	for i := range ns {
		ns[i] = nil
	}
}

type node[T Comparable[T]] struct {
	balance  int8 // bounded, |balance| should be <= 1
	children [2]*node[T]
	entry    T
	hasEntry bool // needed since T might not be nillable
}

// copy returns a copy of this node with pointers to the original children.
func (n *node[T]) copy() *node[T] {
	return &node[T]{
		balance:  n.balance,
		children: [2]*node[T]{n.children[0], n.children[1]},
		entry:    n.entry,
		hasEntry: n.hasEntry,
	}
}

// newNode returns a new node for the provided entry.
func newNode[T Comparable[T]](entry T, hasEntry bool) *node[T] {
	return &node[T]{
		entry:    entry,
		hasEntry: hasEntry,
		children: [2]*node[T]{},
	}
}
