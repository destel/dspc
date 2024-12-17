# DSPC

DSPC - a dead simple progress counter for Go terminal apps. 
Perfect for when you need to track progress of concurrent operations but a full monitoring stack would be overkill.

Think of it as a set of named atomic counters that:
- **Lock-free** - faster than `map[string]int` in both single-threaded and multi-threaded scenarios
- **Zero allocations** - after initialization
- **Nice to look at** - real-time in-place updates in the terminal
- **Minimalistic** - no dependencies, small API



<p align="center"><img src="/img/demo.gif?raw=true"/></p>

## Installation

```bash
go get -u github.com/desel/dspc
```




## Usage

```go
var progress dspc.Progress

// Start progress display that updates every second
defer progress.PrettyPrintEvery(os.Stdout, time.Second, "Progress:")()


// Then, in worker goroutines increment/decrement/set counters as needed 
progress.Inc("processed", 1)
progress.Inc("errors", 1)
progress.Inc("skipped", 1)
```

## When you may (or may not) need it
- You write CLI tools that process data in parallel and launch them:
  - Locally
  - Remotely via SSH
  - In kubernetes via `kubectl exec`
  - In kubernetes as one-off job that's monitored via `kubectl logs -f`
- Your tool is too small to justify adding a monitoring stack or configuring dashboards
- You need multiple counters and tired of declaring atomic variables for each of them
- You want clean, in-place progress updates in the terminal 
- Your list of counters is dynamic (e.g., counting errors by type - "validation_error", "network_error", etc.)
- You find yourself copy-pasting the same counter tracking code between projects


## Customizing the output
It's possible to iterate over all available counters and print them in a custom way. For example:
```go
for key, value := range progress.All() {
    fmt.Printf("[%s]  %d\n", key, value)
}
```






## Benchmarks
```
cpu: Apple M2 Max
BenchmarkSingleThreaded/Map-12         135504824    8.911 ns/op    0 B/op    0 allocs/op
BenchmarkSingleThreaded/Progress-12    145009293    8.554 ns/op    0 B/op    0 allocs/op

BenchmarkMultiThreaded/Map-12                           10755136     110.7 ns/op      0 B/op    0 allocs/op
BenchmarkMultiThreaded/Progress-12                      40967048      47.00 ns/op     0 B/op    0 allocs/op
BenchmarkMultiThreaded/Progress_w_disjoint_keys-12    1000000000       1.035 ns/op    0 B/op    0 allocs/op

BenchmarkPrinting-12    2464957    467.8 ns/op    0 B/op    0 allocs/op
```