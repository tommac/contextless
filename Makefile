.PHONY: build test test-race demo fmt vet clean

BIN := bin/cx

build:
	@mkdir -p bin
	go build -o $(BIN) ./cmd/cx

test:
	go test ./... -count=1

test-race:
	go test -race ./... -count=1

demo: build
	./$(BIN) demo

fmt:
	gofmt -w cmd internal

vet:
	go vet ./...

clean:
	rm -rf bin
