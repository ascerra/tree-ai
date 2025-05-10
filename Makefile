BINARY_NAME=tree-ai
GRANITE_RUNNER=model/granite-runner

all: build

build:
	go build -o bin/$(BINARY_NAME) ./main.go

run:
	go run main.go

test:
	go test ./...

cover:
	go test -coverprofile=coverage.out ./... && \
	go tool cover -html=coverage.out -o coverage.html

fmt:
	go fmt ./...

clean:
	rm -rf bin/ coverage.out coverage.html

setup-python:
	python3 -m venv .venv && \
	source .venv/bin/activate && \
	pip install --upgrade pip && \
	pip install torch transformers

build-granite:
	go build -o $(GRANITE_RUNNER) model/granite-runner.go

install: setup-python build build-granite
	@echo "✅ tree-ai and Granite model runner installed."
	@echo "👉 To activate your environment, run: source .venv/bin/activate"

.PHONY: all build run test fmt clean cover setup-python build-granite install