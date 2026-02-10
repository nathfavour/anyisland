.PHONY: build test lint clean

build:
	go build -o bin/anyisland ./cmd/anyisland

test:
	go test ./...

test-verbose:
	go test -v ./...

lint:
	go fmt ./...
	go vet ./...

clean:
	rm -rf bin/
