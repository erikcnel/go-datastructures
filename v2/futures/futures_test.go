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

package futures

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFutureGetResult(t *testing.T) {
	completer := make(chan string, 1)
	future := New[string](completer, time.Second)

	completer <- "hello"

	result, err := future.GetResult()
	require.NoError(t, err)
	assert.Equal(t, "hello", result)
}

func TestFutureTimeout(t *testing.T) {
	completer := make(chan string, 1)
	future := New[string](completer, 50*time.Millisecond)

	// Don't complete the future

	result, err := future.GetResult()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
	assert.Equal(t, "", result)
}

func TestFutureHasResult(t *testing.T) {
	completer := make(chan string, 1)
	future := New[string](completer, time.Second)

	assert.False(t, future.HasResult())

	completer <- "hello"
	time.Sleep(10 * time.Millisecond)

	assert.True(t, future.HasResult())
}

func TestFutureMultipleListeners(t *testing.T) {
	completer := make(chan int, 1)
	future := New[int](completer, time.Second)

	results := make(chan int, 3)

	for i := 0; i < 3; i++ {
		go func() {
			r, _ := future.GetResult()
			results <- r
		}()
	}

	completer <- 42

	for i := 0; i < 3; i++ {
		assert.Equal(t, 42, <-results)
	}
}

func TestPromise(t *testing.T) {
	promise := NewPromise[string](time.Second)

	go func() {
		time.Sleep(10 * time.Millisecond)
		promise.Complete("done")
	}()

	result, err := promise.Future().GetResult()
	require.NoError(t, err)
	assert.Equal(t, "done", result)
}

func TestPromiseMultipleComplete(t *testing.T) {
	promise := NewPromise[int](time.Second)

	promise.Complete(1)
	promise.Complete(2) // Should be ignored

	result, _ := promise.Future().GetResult()
	assert.Equal(t, 1, result)
}

func TestAll(t *testing.T) {
	c1 := make(chan int, 1)
	c2 := make(chan int, 1)
	c3 := make(chan int, 1)

	f1 := New[int](c1, time.Second)
	f2 := New[int](c2, time.Second)
	f3 := New[int](c3, time.Second)

	go func() {
		c1 <- 1
		c2 <- 2
		c3 <- 3
	}()

	results, err := All(f1, f2, f3)
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, results)
}

func TestRace(t *testing.T) {
	c1 := make(chan string, 1)
	c2 := make(chan string, 1)

	f1 := New[string](c1, time.Second)
	f2 := New[string](c2, time.Second)

	c1 <- "first"
	// c2 never completes

	result, err := Race(f1, f2)
	require.NoError(t, err)
	assert.Equal(t, "first", result)
}
