# dspc [![GoDoc](https://pkg.go.dev/badge/github.com/destel/dspc)](https://pkg.go.dev/github.com/destel/dspc)

Show progress of concurrent tasks in your Go CLI apps.

![dspc demo: progress report along with the log output](demo.svg)

Think of `dspc` as a set of named atomic counters which are:
- **Fast** - lock and allocation free, faster than `map[string]int` in both single-threaded and concurrent scenarios
- **Nice to look at** - clean, readable terminal output that updates in-place
- **Log-friendly** - don't interfere with your application's log output
- **Minimalistic** - no dependencies, no configuration, tiny API


## Installation

```bash
go get -u github.com/destel/dspc
```


## Quick Start

```go
// Create an instance. Zero value is ready to use
var progress dspc.Progress

// Start printing progress to stdout every second
defer progress.PrettyPrintEvery(os.Stdout, 1*time.Second, "Progress:")()


// Then, in worker goroutines just increment/decrement/set counters as needed 
progress.Inc("ok", 1)
progress.Inc("errors", 1)
progress.Inc("skipped", 1)
```

Check out a complete example [here](/example/main.go).


## When to Use
This library is a good fit for CLI applications that do concurrent work.
When running tasks across multiple goroutines, you usually need to track their progress -
the number of tasks that are completed, failed or currently in progress. You may also want to track dynamic categories -
different kinds of tasks, or types of errors (e.g., "validation_error", "network_error", etc).

When running the app in terminal, you want to see a clean progress report that updates in-place,
while keeping your normal application logs readable and separate.

Another example is running such apps in Kubernetes. For simple one-off pods, instead of configuring metrics and dashboards, you
may just want to watch the logs and progress reports in real-time with `kubectl logs -f`.

dspc can also help to debug concurrent applications. 
Add a few counters across your goroutines to see which ones are making progress and which ones are stuck.

If such use cases sound familiar, dspc is what you need.





## Not a Good Fit For
- Long-running services/daemons
- Large number of counters that don't fit on a single screen
- Apps with high-frequency logging (e.g., logs every 10ms) - progress updates may get lost in the log stream 
- Complex metrics - your monitoring needs are not covered by simple counters/gauges


## Performance
```
cpu: Apple M2 Max
BenchmarkSingleThreaded/Map-12         135504824    8.911 ns/op    0 B/op    0 allocs/op
BenchmarkSingleThreaded/Progress-12    145009293    8.554 ns/op    0 B/op    0 allocs/op

BenchmarkMultiThreaded/Map-12                           10755136     110.7 ns/op      0 B/op    0 allocs/op
BenchmarkMultiThreaded/Progress-12                      40967048      47.00 ns/op     0 B/op    0 allocs/op
BenchmarkMultiThreaded/Progress_w_disjoint_keys-12    1000000000       1.035 ns/op    0 B/op    0 allocs/op

BenchmarkPrinting-12    2464957    467.8 ns/op    0 B/op    0 allocs/op
```