/*
Copyright 2015 Workiva, LLC

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

package batcher

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBatcherPut(t *testing.T) {
	b, err := New[string](Config[string]{
		MaxItems: 10,
		MaxTime:  time.Second,
	})
	require.NoError(t, err)
	defer b.Dispose()

	err = b.Put("item1")
	assert.NoError(t, err)
}

func TestBatcherMaxItems(t *testing.T) {
	b, err := New[int](Config[int]{
		MaxItems: 3,
		MaxTime:  time.Minute,
	})
	require.NoError(t, err)
	defer b.Dispose()

	done := make(chan []int, 1)
	go func() {
		batch, _ := b.Get()
		done <- batch
	}()

	b.Put(1)
	b.Put(2)
	b.Put(3)

	select {
	case batch := <-done:
		assert.Equal(t, []int{1, 2, 3}, batch)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for batch")
	}
}

func TestBatcherMaxTime(t *testing.T) {
	b, err := New[string](Config[string]{
		MaxItems: 100,
		MaxTime:  100 * time.Millisecond,
	})
	require.NoError(t, err)
	defer b.Dispose()

	b.Put("item1")

	batch, err := b.Get()
	require.NoError(t, err)
	assert.Equal(t, []string{"item1"}, batch)
}

func TestBatcherFlush(t *testing.T) {
	b, err := New[string](Config[string]{
		MaxItems: 100,
		MaxTime:  time.Minute,
	})
	require.NoError(t, err)
	defer b.Dispose()

	done := make(chan []string, 1)
	go func() {
		batch, _ := b.Get()
		done <- batch
	}()

	b.Put("item1")
	b.Put("item2")
	b.Flush()

	select {
	case batch := <-done:
		assert.Equal(t, []string{"item1", "item2"}, batch)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for batch")
	}
}

func TestBatcherDispose(t *testing.T) {
	b, err := New[string](Config[string]{
		MaxItems: 10,
		MaxTime:  time.Minute,
	})
	require.NoError(t, err)

	b.Put("item1")
	b.Dispose()

	assert.True(t, b.IsDisposed())

	err = b.Put("item2")
	assert.Equal(t, ErrDisposed, err)

	err = b.Flush()
	assert.Equal(t, ErrDisposed, err)
}

func TestBatcherMaxBytes(t *testing.T) {
	b, err := New[string](Config[string]{
		MaxBytes: 10,
		MaxTime:  time.Minute,
		CalculateBytes: func(s string) uint {
			return uint(len(s))
		},
	})
	require.NoError(t, err)
	defer b.Dispose()

	done := make(chan []string, 1)
	go func() {
		batch, _ := b.Get()
		done <- batch
	}()

	b.Put("hello") // 5 bytes
	b.Put("world") // 5 bytes = 10 total, triggers batch

	select {
	case batch := <-done:
		assert.Equal(t, []string{"hello", "world"}, batch)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for batch")
	}
}

func TestBatcherRequiresCalculateBytesWithMaxBytes(t *testing.T) {
	_, err := New[string](Config[string]{
		MaxBytes: 100,
		// No CalculateBytes provided
	})
	assert.Error(t, err)
}
