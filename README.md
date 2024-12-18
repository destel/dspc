# DSPC

DSPC - a dead simple progress counter for concurrent CLI tools in Go. 

Think of it as a set of named atomic counters that:
- **Fast** - lock and allocation free, faster than `map[string]int` in both single-threaded and multi-threaded scenarios
- **Nice to look at** - clean, readable terminal output that updates in place
- **Log-friendly** - won't interfere with your application's logging output in most cases
- **Minimalistic** - no dependencies, tiny API



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

## Perfect when
- Building concurrent CLI tools 
- Counter-based tracking is sufficient
- Want to track progress via stdout/stderr or log tailers - `tail -f`, `kubectl logs -f` etc
- Need to track dynamic categories (e.g., counting errors by type - "validation_error", "network_error", etc.) 
- Want clean progress output that doesn't interfere with normal application logs

## Not a good fit if
- Building a long-running service/daemeon
- Need a large number of counters that don't fit on a single screen when printed
- Application produces a lot of log output (e.g., logs every 10ms) - the progress output might be lost in the log stream 
- Need something more complex than a simple counter (e.g., percentage, rate, progress bar etc.)


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