all: bin/webmentiond

bin:
	mkdir -p bin

clean:
	rm -rf bin

bin/webmentiond: $(shell find . -name '*.go') go.mod bin
	cd cmd/webmentiond && go build -o ../../$@

test:
	go test ./... -v

.PHONY: clean all test
