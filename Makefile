.PHONY: all build run test test-race clean docker-build docker-run

BINARY_NAME=afterdark
PORT?=2222
DATA_DIR?=data

all: test build

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BINARY_NAME) ./cmd/afterdark

run: build
	PORT=$(PORT) DATA_DIR=$(DATA_DIR) ./$(BINARY_NAME)

test:
	go test -v -count=1 ./internal/...

test-race:
	go test -race -v ./internal/chat ./internal/security

clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe

docker-build:
	docker build -t afterdark:latest .

docker-run:
	docker run -d --name afterdark -p $(PORT):2222 -v $(PWD)/data:/app/data afterdark:latest
