# go-datastructures v2

Go-datastructures v2 is a modernized version of the popular go-datastructures library,
rebuilt with Go generics for type safety and improved developer experience.

## Requirements

- **Go 1.22+** (for generics support)

## Installation

```bash
go get github.com/Workiva/go-datastructures/v2
```

## What's New in v2

### Generics Throughout
All data structures now use Go generics, providing:
- **Type safety**: Compile-time type checking instead of runtime type assertions
- **No more `interface{}`**: Use concrete types directly
- **Better IDE support**: Full autocomplete and type inference

### Modern Go Idioms
- Uses `any` instead of `interface{}`
- Leverages `cmp.Ordered` constraint for ordered types
- Follows Go 1.22+ best practices

### API Improvements
- More consistent method naming
- Additional utility methods (e.g., `Filter`, `Map`, `Reduce` on lists)
- Better error handling patterns

## Available Data Structures

### Queue
Thread-safe queue implementations that never block on send.

```go
import "github.com/Workiva/go-datastructures/v2/queue"

// Standard queue
q := queue.New[string](100)
q.Put("hello", "world")
items, _ := q.Get(2) // []string{"hello", "world"}

// Priority queue with custom ordering
type Task struct {
    Name     string
    Priority int
}

func (t Task) Compare(other Task) int {
    return t.Priority - other.Priority
}

pq := queue.NewPriorityQueue[Task](10, true)
pq.Put(Task{"low", 10}, Task{"high", 1})
items, _ := pq.Get(1) // Task{"high", 1} (lowest priority value first)

// Convenience priority queue with integer priorities
opq := queue.NewOrderedPriorityQueue[string](10, true)
opq.Enqueue("task1", 5)
opq.Enqueue("task2", 1)
val, priority, _ := opq.Dequeue() // "task2", 1

// Lock-free ring buffer
rb := queue.NewRingBuffer[int](1024)
rb.Put(42)
val, _ := rb.Get() // 42
```

### Set
Thread-safe set with full set operations.

```go
import "github.com/Workiva/go-datastructures/v2/set"

s := set.New[string]("a", "b", "c")
s.Add("d")
s.Remove("a")

exists := s.Exists("b")    // true
all := s.All("b", "c")     // true
any := s.Any("x", "c")     // true

// Set operations
s1 := set.New[int](1, 2, 3)
s2 := set.New[int](2, 3, 4)

union := s1.Union(s2)           // {1, 2, 3, 4}
inter := s1.Intersection(s2)   // {2, 3}
diff := s1.Difference(s2)      // {1}
symDiff := s1.SymmetricDifference(s2) // {1, 4}

// Functional operations
evens := s1.Filter(func(n int) bool { return n%2 == 0 })
```

### List (Persistent/Immutable)
Immutable linked list with functional operations.

```go
import "github.com/Workiva/go-datastructures/v2/list"

// Create list (items are prepended)
l := list.Empty[int]().Add(1).Add(2).Add(3)
// List is [3, 2, 1] (3 is head)

head, _ := l.Head() // 3
tail, _ := l.Tail() // list containing [2, 1]

// Immutability - original unchanged
l2 := l.Add(4)
l.Length()  // 3
l2.Length() // 4

// Functional operations
doubled := l.Map(func(x int) int { return x * 2 })
sum := l.Reduce(func(acc, x int) int { return acc + x }, 0)
evens := l.Filter(func(x int) bool { return x%2 == 0 })

// Create from slice
l3 := list.FromSliceReversed([]int{1, 2, 3}) // head is 1
```

### Cache
Bounded-size LRU/LRA cache.

```go
import "github.com/Workiva/go-datastructures/v2/cache"

// Items must implement Sized interface
type MyItem struct {
    Data []byte
}

func (m MyItem) Size() uint64 {
    return uint64(len(m.Data))
}

// Create cache with 1MB capacity
c := cache.New[string, MyItem](1024 * 1024)

c.Put("key1", MyItem{Data: []byte("hello")})
item, ok := c.Get("key1")

// Simple cache (counts items, not bytes)
sc := cache.NewSimple[string, int](100) // max 100 items
sc.Put("key", 42)
val, ok := sc.Get("key")
```

### Batcher
Accumulate items into batches for processing.

```go
import "github.com/Workiva/go-datastructures/v2/batcher"

b, _ := batcher.New[string](batcher.Config[string]{
    MaxTime:  time.Second,
    MaxItems: 100,
})

// Consumer
go func() {
    for {
        batch, err := b.Get()
        if err != nil {
            return
        }
        processBatch(batch)
    }
}()

// Producer
b.Put("item1")
b.Put("item2")
```

### Futures
Broadcast results to multiple listeners.

```go
import "github.com/Workiva/go-datastructures/v2/futures"

// Basic future
completer := make(chan string, 1)
future := futures.New[string](completer, 5*time.Second)

// Multiple goroutines can wait
go func() {
    result, err := future.GetResult()
    // ...
}()

completer <- "hello"

// Promise pattern
promise := futures.NewPromise[int](5 * time.Second)
go func() {
    result, _ := promise.Future().GetResult()
    fmt.Println(result)
}()
promise.Complete(42)

// Await helper
future := futures.Await(func() (string, error) {
    return fetchData()
}, 10*time.Second)

// Combinators
results, err := futures.All(future1, future2, future3)
first, err := futures.Race(future1, future2)
```

### Common Utilities
Shared types and constraints.

```go
import "github.com/Workiva/go-datastructures/v2/common"

// Use Ordered constraint for sorted collections
func Min[T common.Ordered](a, b T) T {
    if common.Less(a, b) {
        return a
    }
    return b
}

// Compare function for ordered types
cmp := common.OrderedCompare[int]()
result := cmp(1, 2) // -1
```

## Migration from v1

### Key Changes

1. **Import path**: Change imports from `github.com/Workiva/go-datastructures` to `github.com/Workiva/go-datastructures/v2`

2. **Add type parameters**: 
   ```go
   // v1
   q := queue.New(100)
   q.Put("hello")
   items, _ := q.Get(1)
   str := items[0].(string) // type assertion required
   
   // v2
   q := queue.New[string](100)
   q.Put("hello")
   items, _ := q.Get(1)
   str := items[0] // already string type
   ```

3. **Interface changes**: Some interfaces have been updated to use generics. Check method signatures in the documentation.

4. **Removed `interface{}`**: All uses of `interface{}` replaced with type parameters or `any`.

## Testing

```bash
cd v2
go test ./...
```

## Contributing

Contributions are welcome! Please ensure:
- Code is `gofmt`'d
- Tests are included
- Documentation is updated

## License

Apache License 2.0 - See [LICENSE](../LICENSE) for details.

