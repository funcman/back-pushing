.PHONY: build build-cli run run-cli test clean

build:
	go build -o bin/server ./cmd/server

build-cli:
	go build -o bin/cli ./cmd/cli

run: build
	./bin/server

run-cli: build-cli
	./bin/cli

test:
	go test ./...

clean:
	rm -rf bin/
