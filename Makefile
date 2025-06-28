# Makefile for the a2m2a project

# Get the application name from the current directory.
APP_NAME := $(shell basename "$(CURDIR)")
VERSION ?= 1.0.0
DIST_DIR := dist

# Default target builds an optimized binary for the current system.
.PHONY: all
all: build

.PHONY: build
build:
	@echo "Building optimized binary for local machine..."
	@go build -ldflags="-s -w" -trimpath -o $(APP_NAME) .
	@echo "Build complete: ./$(APP_NAME)"

# Release target cross-compiles for Linux, Windows, and macOS, then packages them.
.PHONY: release
release: clean
	@echo "Creating release builds for version $(VERSION)..."

	# --- Linux (amd64) ---
	@echo "  -> Building for Linux..."
	@GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o $(DIST_DIR)/$(APP_NAME) .
	@echo "  -> Packaging for Linux..."
	@tar -czvf $(DIST_DIR)/$(APP_NAME)-$(VERSION)-linux-amd64.tar.gz -C $(DIST_DIR) $(APP_NAME)
	@rm $(DIST_DIR)/$(APP_NAME)

	# --- Windows (amd64) ---
	@echo "  -> Building for Windows..."
	@GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o $(DIST_DIR)/$(APP_NAME).exe .
	@echo "  -> Packaging for Windows..."
	@zip -j $(DIST_DIR)/$(APP_NAME)-$(VERSION)-windows-amd64.zip $(DIST_DIR)/$(APP_NAME).exe
	@rm $(DIST_DIR)/$(APP_NAME).exe

	# --- Darwin/macOS (amd64) ---
	@echo "  -> Building for macOS..."
	@GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o $(DIST_DIR)/$(APP_NAME) .
	@echo "  -> Packaging for macOS..."
	@zip -j $(DIST_DIR)/$(APP_NAME)-$(VERSION)-darwin-amd64.zip $(DIST_DIR)/$(APP_NAME)
	@rm $(DIST_DIR)/$(APP_NAME)

	@echo "Release builds are in the '$(DIST_DIR)' directory."

# Clean target removes all build artifacts.
.PHONY: clean
clean:
	@echo "Cleaning up..."
	@rm -f $(APP_NAME)
	@rm -rf $(DIST_DIR)
	@mkdir -p $(DIST_DIR) 