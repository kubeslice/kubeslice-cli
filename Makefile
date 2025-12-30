.PHONY: help build build-all clean

## Show help info
help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  build        Build CLI for current OS/Arch"
	@echo "  build-all    Cross-compile CLI for all major platforms"
	@echo "  clean        Remove all compiled binaries from ./bin"
	@echo "  help         Show this help message"

## Build CLI for current OS/Arch
build:
	@mkdir -p bin
	@goos=$$(go env GOOS); \
	goarch=$$(go env GOARCH); \
	ext=""; \
	if [ "$$goos" = "windows" ]; then ext=".exe"; fi; \
	out="bin/kubeslice-cli-$$goos-$$goarch$$ext"; \
	echo "Building for $$goos/$$goarch -> $$out"; \
	go build -o $$out main.go

## Build CLI for all target platforms
build-all:
	@mkdir -p bin
	GOOS=windows GOARCH=amd64 go build -o bin/kubeslice-cli-windows-amd64.exe main.go
	GOOS=linux GOARCH=amd64 go build -o bin/kubeslice-cli-linux-amd64 main.go
	GOOS=linux GOARCH=arm go build -o bin/kubeslice-cli-linux-arm main.go
	GOOS=linux GOARCH=arm64 go build -o bin/kubeslice-cli-linux-arm64 main.go
	GOOS=darwin GOARCH=amd64 go build -o bin/kubeslice-cli-darwin-amd64 main.go
	GOOS=darwin GOARCH=arm64 go build -o bin/kubeslice-cli-darwin-arm64 main.go

## Clean up build artifacts
clean:
	@echo "Removing bin directory and compiled binaries..."
	@rm -rf bin