
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

COVEROUT=cover.out
cover: $(COVERPATH)
	@go test -coverprofile=$(COVEROUT) -v $(PKGNAME)/lrucache
	@cat $(COVEROUT) | egrep -v '$(SKIPCOVER)' > $(COVEROUT)~
	@go test -coverprofile=$(COVEROUT) -v $(PKGNAME)/multilru
	@cat $(COVEROUT) | egrep -v 'mode: set|$(SKIPCOVER)' >> $(COVEROUT)~
	@go tool cover -func=$(COVEROUT)~|sed 's|^.*/\([^/]*/[^/]*/[^/]*\)$$|\1|g'

bench:
	@echo "[*] Operations in shared cache using one core	"
	@GOMAXPROCS=1 go test -run=- -bench='BenchmarkConcurrent.*' $(PKGNAME) | egrep -v "^PASS|^ok"
	@echo "[*] Operations in shared cache using two cores	"
	@GOMAXPROCS=2 go test -run=- -bench='BenchmarkConcurrent.*' $(PKGNAME) | egrep -v "^PASS|^ok"
	@echo "[*] Operations in shared cache using four cores	"
	@GOMAXPROCS=4 go test -run=- -bench='BenchmarkConcurrent.*' $(PKGNAME) | egrep -v "^PASS|^ok"


COVERPATH=$(GOPATH)/src/code.google.com/p/go.tools/cmd/cover
$(COVERPATH):
	go get code.google.com/p/go.tools/cmd/cover

clean:
	rm -rf $(COVEROUT) $(COVEROUT)~
