BINARY_NAME=stqry
BUILD_DIR=bin
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

.PHONY: build test test-e2e build-recorder record lint clean

build:
	go build -ldflags "-X github.com/mytours/stqry-cli/internal/buildinfo.Version=$(VERSION)" \
		-o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/stqry

build-recorder:
	go build -o $(BUILD_DIR)/recorder ./e2e/recorder

test:
	go test ./... -v

test-e2e: build build-recorder
	e2e/run.sh

record: build build-recorder
	@if [ -z "$(TEST_API_URL)" ]; then echo "Error: TEST_API_URL is required. Set it in .env or run: TEST_API_URL=https://... make record"; exit 1; fi
	bin/recorder --mode=record --target=$(TEST_API_URL) --cassettes=e2e/cassettes/happypath

lint:
	go vet ./...

clean:
	rm -rf $(BUILD_DIR)
