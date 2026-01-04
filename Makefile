# lah Makefile

.PHONY: build build-linux build-mac build-all install install-linux install-mac clean test help

# Default target
all: build

# Build for current platform
build:
	go build -o lah lah.go

# Build for Linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o lah-linux-amd64 lah.go

# Build for macOS
build-mac:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o lah-darwin-amd64 lah.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o lah-darwin-arm64 lah.go

# Build for all platforms
build-all: build-linux build-mac

# Install to ~/bin (current platform)
install: build
	@echo "Installing lah to ~/bin..."
	@mkdir -p ~/bin
	@cp lah ~/bin/
	@echo "Installation complete! Make sure ~/bin is in your PATH."

# Install Linux binary
install-linux: build-linux
	@echo "Installing lah for Linux..."
	@mkdir -p ~/bin
	@cp lah-linux-amd64 ~/bin/lah
	@echo "Installation complete!"

# Install macOS binary
install-mac: build-mac
	@echo "Installing lah for macOS..."
	@if [ "$(shell uname -m)" = "arm64" ]; then \
		cp lah-darwin-arm64 ~/bin/lah; \
	else \
		cp lah-darwin-amd64 ~/bin/lah; \
	fi
	@mkdir -p ~/bin
	@echo "Installation complete!"

# Install to /usr/local/bin (requires sudo)
install-system: build
	@echo "Installing lah to /usr/local/bin..."
	@sudo cp lah /usr/local/bin/
	@echo "Installation complete! lah is now available system-wide."

# Run lah from source (pass args with ARGS, e.g. make run ARGS="-tg")
ARGS ?=
run:
	@echo "Running lah from source with args: $(ARGS)"
	go run . $(ARGS)

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f lah lah-*
	@go clean -cache
	@echo "Clean complete!"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	@golangci-lint run

# Show help
help:
	@echo "Available targets:"
	@echo "  build          - Build the lah binary"
	@echo "  install        - Install lah to ~/bin"
	@echo "  install-system - Install lah to /usr/local/bin (requires sudo)"
	@echo "  clean          - Remove build artifacts"
	@echo "  test           - Run tests"
	@echo "  fmt            - Format code"
	@echo "  lint           - Lint code"
	@echo "  help           - Show this help message"
