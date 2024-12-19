# DSPC

DSPC - a dead simple progress counter for concurrent CLI apps in Go.

Think of it as a set of named atomic counters that:
- **Fast** - lock and allocation free, faster than `map[string]int` in both single-threaded and concurrent scenarios
- **Nice to look at** - clean, readable terminal output that updates in place
- **Log-friendly** - won't interfere with your application's log output
- **Minimalistic** - no dependencies, tiny API



<p align="center"><img src="/img/demo.gif?raw=true"/></p>

## Installation

```bash
go get -u github.com/destel/dspc
```


## Usage

```go
// Create an instance. Zero value is ready to use
var progress dspc.Progress

// Start printing progress to stdout every second
defer progress.PrettyPrintEvery(os.Stdout, time.Second, "Progress:")()


// Then, in worker goroutines increment/decrement/set counters as needed 
progress.Inc("ok", 1)
progress.Inc("errors", 1)
progress.Inc("skipped", 1)
```

## Good when
This library is a good fit for CLI applications that do concurrent work. 
When running tasks across multiple goroutines, you'll likely want to track their progress - 
the number of completed tasks, errors, tasks currently in progress. You even may want to track dynamic categories -
like counting errors by type - "validation_error", "network_error", etc.

When running the app in terminal, you'll want to see a clean progress report that updates in-place and in real-time, 
while keeping your normal application logs readable and separate.

Another example is running such app in Kubernetes. For simple one-off pods, instead of configuring metrics and dashboards, you'll 
likely want to just watch the logs and progress in real-time with `kubectl logs -f`.

If this matches your needs - tracking concurrent work with simple counters and clean terminal output - then DSPC is likely
what you're looking for.





## Not a good fit for
- Long-running services/daemons
- Large number of counters that don't fit on a single screen
- Apps with high-frequency logging (e.g., logs every 10ms) - progress updates may get lost in the log stream 
- Complex metrics - your monitoring needs is not covered by counters/gauges


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