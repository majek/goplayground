LRU Cache
----

To install:

    go get github.com/majek/goplayground/cache

To test:

    cd $GOPATH/src/github.com/majek/goplayground/cache
    make test

For coverage:

    make cover

For benchmarks (mostly about scalability on muticore):

```
$ make bench  # As tested on my two core i5
[*] Operations in shared cache using one core
BenchmarkConcurrentGet           5000000               446 ns/op
BenchmarkConcurrentSet           2000000               965 ns/op
BenchmarkConcurrentSetNX         5000000               752 ns/op
[*] Operations in shared cache using two cores
BenchmarkConcurrentGet-2         2000000               797 ns/op
BenchmarkConcurrentSet-2         1000000              1633 ns/op
BenchmarkConcurrentSetNX-2       1000000              1361 ns/op
[*] Operations in shared cache using four cores
BenchmarkConcurrentGet-4         5000000               745 ns/op
BenchmarkConcurrentSet-4         1000000              1649 ns/op
BenchmarkConcurrentSetNX-4       1000000              1278 ns/op
```

