.PHONY: build build-cli build-server build-web build-desktop build-all run run-cli test clean

build:
	go build -o bin/server ./cmd/server

build-cli:
	go build -o bin/cli ./cmd/cli

build-server:
	go build -o bin/server ./cmd/server

build-web:
	cd web && npm install && npm run build

build-desktop:
	cd desktop && npm install

build-all: build-server build-web build-desktop

run: build
	./bin/server

run-cli: build-cli
	./bin/cli

test:
	go test ./...

clean:
	rm -rf bin/
