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

package queue

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueuePut(t *testing.T) {
	q := New[string](10)

	err := q.Put("hello", "world")
	require.NoError(t, err)

	assert.Equal(t, int64(2), q.Len())
}

func TestQueueGet(t *testing.T) {
	q := New[int](10)

	q.Put(1, 2, 3)

	items, err := q.Get(2)
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2}, items)

	items, err = q.Get(2)
	require.NoError(t, err)
	assert.Equal(t, []int{3}, items)
}

func TestQueuePeek(t *testing.T) {
	q := New[string](10)

	q.Put("first", "second")

	item, err := q.Peek()
	require.NoError(t, err)
	assert.Equal(t, "first", item)

	// Peek shouldn't remove the item
	assert.Equal(t, int64(2), q.Len())
}

func TestQueuePeekEmpty(t *testing.T) {
	q := New[string](10)

	_, err := q.Peek()
	assert.Equal(t, ErrEmptyQueue, err)
}

func TestQueuePoll(t *testing.T) {
	q := New[int](10)

	// Test timeout on empty queue
	_, err := q.Poll(1, 50*time.Millisecond)
	assert.Equal(t, ErrTimeout, err)

	// Add item and get it
	q.Put(42)
	items, err := q.Poll(1, 50*time.Millisecond)
	require.NoError(t, err)
	assert.Equal(t, []int{42}, items)
}

func TestQueueDispose(t *testing.T) {
	q := New[string](10)

	q.Put("a", "b", "c")

	disposed := q.Dispose()
	assert.Equal(t, []string{"a", "b", "c"}, disposed)

	assert.True(t, q.Disposed())

	err := q.Put("d")
	assert.Equal(t, ErrDisposed, err)
}

func TestQueueTakeUntil(t *testing.T) {
	q := New[int](10)

	q.Put(1, 2, 3, 4, 5)

	items, err := q.TakeUntil(func(item int) bool {
		return item < 4
	})
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, items)

	// Remaining items
	remaining, _ := q.Get(10)
	assert.Equal(t, []int{4, 5}, remaining)
}

func TestQueueEmpty(t *testing.T) {
	q := New[int](10)

	assert.True(t, q.Empty())

	q.Put(1)
	assert.False(t, q.Empty())
}

// PriorityQueue tests

type testPriorityItem struct {
	value    string
	priority int
}

func (t testPriorityItem) Compare(other testPriorityItem) int {
	if t.priority < other.priority {
		return -1
	}
	if t.priority > other.priority {
		return 1
	}
	return 0
}

func TestPriorityQueuePut(t *testing.T) {
	pq := NewPriorityQueue[testPriorityItem](10, true)

	err := pq.Put(
		testPriorityItem{value: "low", priority: 10},
		testPriorityItem{value: "high", priority: 1},
	)
	require.NoError(t, err)

	assert.Equal(t, 2, pq.Len())
}

func TestPriorityQueueGetOrder(t *testing.T) {
	pq := NewPriorityQueue[testPriorityItem](10, true)

	pq.Put(
		testPriorityItem{value: "low", priority: 10},
		testPriorityItem{value: "high", priority: 1},
		testPriorityItem{value: "medium", priority: 5},
	)

	items, err := pq.Get(3)
	require.NoError(t, err)

	// Should come out in priority order (lowest first)
	assert.Equal(t, "high", items[0].value)
	assert.Equal(t, "medium", items[1].value)
	assert.Equal(t, "low", items[2].value)
}

func TestPriorityQueuePeek(t *testing.T) {
	pq := NewPriorityQueue[testPriorityItem](10, true)

	pq.Put(
		testPriorityItem{value: "low", priority: 10},
		testPriorityItem{value: "high", priority: 1},
	)

	item, ok := pq.Peek()
	assert.True(t, ok)
	assert.Equal(t, "high", item.value)

	// Peek shouldn't remove
	assert.Equal(t, 2, pq.Len())
}

func TestOrderedPriorityQueue(t *testing.T) {
	opq := NewOrderedPriorityQueue[string](10, true)

	opq.Enqueue("low", 10)
	opq.Enqueue("high", 1)
	opq.Enqueue("medium", 5)

	val, priority, err := opq.Dequeue()
	require.NoError(t, err)
	assert.Equal(t, "high", val)
	assert.Equal(t, 1, priority)

	val, priority, err = opq.Dequeue()
	require.NoError(t, err)
	assert.Equal(t, "medium", val)
	assert.Equal(t, 5, priority)
}

// RingBuffer tests

func TestRingBufferPutGet(t *testing.T) {
	rb := NewRingBuffer[int](4)

	err := rb.Put(1)
	require.NoError(t, err)

	val, err := rb.Get()
	require.NoError(t, err)
	assert.Equal(t, 1, val)
}

func TestRingBufferOffer(t *testing.T) {
	rb := NewRingBuffer[string](2)

	ok, err := rb.Offer("a")
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = rb.Offer("b")
	require.NoError(t, err)
	assert.True(t, ok)

	// Buffer is full, offer should return false
	ok, err = rb.Offer("c")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestRingBufferLen(t *testing.T) {
	rb := NewRingBuffer[int](4)

	assert.Equal(t, uint64(0), rb.Len())

	rb.Put(1)
	rb.Put(2)
	assert.Equal(t, uint64(2), rb.Len())
}

func TestRingBufferCap(t *testing.T) {
	rb := NewRingBuffer[int](5)

	// Should be rounded up to 8 (next power of 2)
	assert.Equal(t, uint64(8), rb.Cap())
}

func TestRingBufferDispose(t *testing.T) {
	rb := NewRingBuffer[int](4)

	rb.Dispose()
	assert.True(t, rb.IsDisposed())

	err := rb.Put(1)
	assert.Equal(t, ErrDisposed, err)

	_, err = rb.Get()
	assert.Equal(t, ErrDisposed, err)
}

func TestRingBufferPoll(t *testing.T) {
	rb := NewRingBuffer[int](4)

	// Test timeout
	_, err := rb.Poll(50 * time.Millisecond)
	assert.Equal(t, ErrTimeout, err)
}

func TestExecuteInParallel(t *testing.T) {
	q := New[int](10)

	for i := range 100 {
		q.Put(i)
	}

	sum := 0
	ch := make(chan int, 100)

	ExecuteInParallel(q, func(item int) {
		ch <- item
	})

	close(ch)
	for v := range ch {
		sum += v
	}

	// Sum of 0..99 = 4950
	assert.Equal(t, 4950, sum)
}
