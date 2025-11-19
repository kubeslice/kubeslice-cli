GO_VERSION := 1.17

.PHONY: all build clean test test-unit test-integration lint tidy

all: build

# test-unit: Runs all unit tests.
# Uses -short to skip long-running tests.
test-unit:
	@echo "Running unit tests..."
	@go test -v -race -short -cover ./...

# test-integration: Runs only integration tests.
# Requires the 'integration' build tag.
test-integration:
	@echo "Running integration tests..."
	@go test -v -race -cover -tags=integration -run Integration ./...

# test: Runs All tests
test:
	@echo "Running all tests..."
	@make test-unit
	@make test-integration

build: tidy
	@echo "Building binaries for all platforms..."
	@GOOS=windows GOARCH=amd64 go build -o bin/kubeslice-cli-windows-amd64.exe main.go
	@GOOS=linux GOARCH=amd64 go build -o bin/kubeslice-cli-linux-amd64 main.go
	@GOOS=linux GOARCH=arm go build -o bin/kubeslice-cli-linux-arm main.go
	@GOOS=linux GOARCH=arm64 go build -o bin/kubeslice-cli-linux-arm64 main.go
	@GOOS=darwin GOARCH=amd64 go build -o bin/kubeslice-cli-darwin-amd64 main.go
	@GOOS=darwin GOARCH=arm64 go build -o bin/kubeslice-cli-darwin-arm64 main.go
	@echo "Build complete."

lint:
	@echo "Running linter..."
	@if [ "$$(gofmt -l . | wc -l)" -gt 0 ]; then \
		echo "Go files are not formatted. Please run 'gofmt -w .'"; \
		gofmt -l .; \
		exit 1; \
	fi
	@go vet ./...
	@echo "Linting passed!"

tidy:
	@echo "Running go mod tidy..."
	@go mod tidy

clean:
	@echo "Cleaning up..."
	@rm -f bin/kubeslice-cli-*
	@rm -rf bin
