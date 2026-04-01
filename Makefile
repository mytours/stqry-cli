BINARY_NAME=stqry
BUILD_DIR=bin

.PHONY: build test lint clean

build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/stqry

test:
	go test ./... -v

lint:
	go vet ./...

clean:
	rm -rf $(BUILD_DIR)
