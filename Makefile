.PHONY: build build-all clean run watch explain help test deps

# Directories
BIN_DIR := bin

# Detect OS
UNAME_S := $(shell uname -s 2>/dev/null || echo Windows_NT)

# Build the procscope binary for current platform
build:
	@echo "Building procscope for current platform..."
	@mkdir -p $(BIN_DIR)
ifeq ($(UNAME_S),Linux)
	@go build -o $(BIN_DIR)/procscope .
else ifeq ($(UNAME_S),Darwin)
	@go build -o $(BIN_DIR)/procscope .
else
	@go build -o $(BIN_DIR)/procscope.exe .
endif
	@echo "Build complete: $(BIN_DIR)/procscope"

# Build for all platforms
build-all:
	@echo "Building procscope for all platforms..."
	@mkdir -p $(BIN_DIR)
	@echo "Building for Windows..."
	@GOOS=windows GOARCH=amd64 go build -o $(BIN_DIR)/procscope.exe .
	@echo "Building for Linux..."
	@GOOS=linux GOARCH=amd64 go build -o $(BIN_DIR)/procscope-linux .
	@echo "Building for macOS..."
	@GOOS=darwin GOARCH=amd64 go build -o $(BIN_DIR)/procscope-mac .
	@echo "All builds complete in $(BIN_DIR)/"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BIN_DIR)
	@echo "Clean complete"

# Run the binary with watch command
watch: build
ifeq ($(UNAME_S),Windows_NT)
	@$(BIN_DIR)/procscope.exe watch
else
	@$(BIN_DIR)/procscope watch
endif

# Run the binary with explain command (default to PID 1)
explain: build
ifeq ($(UNAME_S),Windows_NT)
	@$(BIN_DIR)/procscope.exe explain 1
else
	@$(BIN_DIR)/procscope explain 1
endif

# Test the project
test:
	@echo "Running tests..."
	@go test ./...

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Show help
help:
	@echo "procscope build targets:"
	@echo "  make build       - Build the procscope binary for current platform"
	@echo "  make build-all   - Build for Windows, Linux, and macOS"
	@echo "  make clean       - Remove build artifacts"
	@echo "  make watch       - Build and run 'procscope watch'"
	@echo "  make explain     - Build and run 'procscope explain'"
	@echo "  make test        - Run tests"
	@echo "  make deps        - Download and tidy dependencies"
	@echo "  make help        - Show this help message"
