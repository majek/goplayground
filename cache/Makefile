
RACE+=--race

PKGNAME=github.com/majek/goplayground/cache
SKIPCOVER=list.go|list_extension.go|priorityqueue.go

.PHONY: all test bench cover clean

all:
	@echo "Targets:"
	@echo "  test: run tests with race detector"
	@echo "  cover: print test coverage"
	@echo "  bench: run basic benchmarks"

test:
	@go test $(RACE) -bench=. -v $(PKGNAME)/lrucache
	@go test $(RACE) -bench=. -v $(PKGNAME)/multilru

COVEROUT=cover.out
cover: $(COVERPATH)
	@go test -coverprofile=$(COVEROUT) -v $(PKGNAME)/lrucache
	@cat $(COVEROUT) | egrep -v '$(SKIPCOVER)' > $(COVEROUT)~
	@go test -coverprofile=$(COVEROUT) -v $(PKGNAME)/multilru
	@cat $(COVEROUT) | egrep -v 'mode: set|$(SKIPCOVER)' >> $(COVEROUT)~
	@go tool cover -func=$(COVEROUT)~|sed 's|^.*/\([^/]*/[^/]*/[^/]*\)$$|\1|g'

bench:
	@echo "[*] Scalability of cache/lrucache"
	@echo "[ ] Operations in shared cache using one core	"
	@GOMAXPROCS=1 go test -run=- -bench='BenchmarkConcurrent.*' $(PKGNAME)/lrucache | egrep -v "^PASS|^ok"
	@echo "[ ] Operations in shared cache using two cores	"
	@GOMAXPROCS=2 go test -run=- -bench='BenchmarkConcurrent.*' $(PKGNAME)/lrucache | egrep -v "^PASS|^ok"
	@echo "[ ] Operations in shared cache using four cores	"
	@GOMAXPROCS=4 go test -run=- -bench='BenchmarkConcurrent.*' $(PKGNAME)/lrucache | egrep -v "^PASS|^ok"

	@echo "[*] Scalability of cache/multilru"
	@echo "[ ] Operations in four caches using one core	"
	@GOMAXPROCS=1 go test -run=- -bench='BenchmarkConcurrent.*' $(PKGNAME)/multilru | egrep -v "^PASS|^ok"
	@echo "[ ] Operations in four caches using two cores	"
	@GOMAXPROCS=2 go test -run=- -bench='BenchmarkConcurrent.*' $(PKGNAME)/multilru | egrep -v "^PASS|^ok"
	@echo "[ ] Operations in four caches using four cores	"
	@GOMAXPROCS=4 go test -run=- -bench='BenchmarkConcurrent.*' $(PKGNAME)/multilru | egrep -v "^PASS|^ok"


	@(cd benchmark; go build github.com/majek/goplayground/cache/benchmark)
	@./benchmark/benchmark

COVERPATH=$(GOPATH)/src/code.google.com/p/go.tools/cmd/cover
$(COVERPATH):
	go get code.google.com/p/go.tools/cmd/cover

clean:
	rm -rf $(COVEROUT) $(COVEROUT)~
