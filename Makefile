APP_NAME := pingforge
BUILD_DIR := bin
CMD_DIR := ./cmd/pingforge
VERSION := 1.0.0
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +'%Y-%m-%d')
LDFLAGS := -ldflags "-X github.com/sujanv/pingforge/internal/cli.Version=$(VERSION) -X github.com/sujanv/pingforge/internal/cli.GitCommit=$(COMMIT) -X github.com/sujanv/pingforge/internal/cli.BuildDate=$(DATE)"

.PHONY: all build test benchmark clean run install

all: test build

build:
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) $(CMD_DIR)
	@echo "✔ Binary built at $(BUILD_DIR)/$(APP_NAME)"

test:
	go test -v -race ./...

benchmark:
	go test -bench=. -benchmem ./tests/...

run: build
	./$(BUILD_DIR)/$(APP_NAME) --help

clean:
	rm -rf $(BUILD_DIR) report.html report.csv history.json

install:
	go install $(LDFLAGS) $(CMD_DIR)
