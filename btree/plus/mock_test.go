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

func chunkKeys(ks keySlice[*mockKey], numParts int64) []keySlice[*mockKey] {
	parts := make([]keySlice[*mockKey], numParts)
	for i := range numParts {
		parts[i] = ks[i*int64(len(ks))/numParts : (i+1)*int64(len(ks))/numParts]
	}
	return parts
}

type mockKey struct {
	value int
}

// Compare implements Comparable[*mockKey]
func (mk *mockKey) Compare(other *mockKey) int {
	if other.value == mk.value {
		return 0
	}
	if other.value > mk.value {
		return 1
	}
	return -1
}

func newMockKey(value int) *mockKey {
	return &mockKey{value}
}

func constructMockKeys(num int) keySlice[*mockKey] {
	keys := make(keySlice[*mockKey], 0, num)
	for i := range num {
		keys = append(keys, newMockKey(i))
	}
	return keys
}
