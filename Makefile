APP_NAME=ratatoskr
BIN_DIR=bin

# Platform-specific variables
LINUX_AMD64=$(BIN_DIR)/linux-amd64
DARWIN_AMD64=$(BIN_DIR)/darwin-amd64
LINUX_ARM64=$(BIN_DIR)/linux-arm64

.PHONY: dev build-linux build-macos build-rpi build-all clean help

main:
	go build -o $(BIN_DIR)/$(APP_NAME) cmd/$(APP_NAME)/main.go

dev:
	air 

build-linux: 
	GOOS=linux GOARCH=amd64 go build -o $(LINUX_AMD64)/$(APP_NAME) cmd/$(APP_NAME)/main.go

build-macos: 
	GOOS=darwin GOARCH=amd64 go build -o $(DARWIN_AMD64)/$(APP_NAME) cmd/$(APP_NAME)/main.go

build-rpi: 
	GOOS=linux GOARCH=arm64 go build -o $(LINUX_ARM64)/$(APP_NAME) cmd/$(APP_NAME)/main.go

build-all: build-linux build-macos build-rpi

clean:
	rm -rf $(BIN_DIR)

help:
    @echo "Usage:"
	@echo "  make main        - Build the main binary for the current platform"
	@echo "  make dev         - Run the application for development"
	@echo "  make build-linux - Build the binary for Linux (amd64)"
	@echo "  make build-macos - Build the binary for macOS (amd64)"
	@echo "  make build-rpi   - Build the binary for Raspberry Pi (arm64)"
	@echo "  make build-all   - Build binaries for all platforms"
	@echo "  make clean       - Remove all build artifacts"

pushpi:
	ssh $(PI) "mkdir -p ~/src/"
	rsync -av --progress $(LINUX_ARM64) $(PI):~/src/$(APP_NAME)/bin
	rsync -av --progress scripts $(PI):~/src/$(APP_NAME)
	rsync -av --progress Makefile $(PI):~/src/$(APP_NAME)
	rsync -av --progress docker-compose.yml $(PI):~/src/$(APP_NAME)

install:
	# stop the service if it already exists
	systemctl stop ratatoskr || true
	systemctl disable ratatoskr || true
	# delete the old service file if it exists
	rm /etc/systemd/system/ratatoskr.service || true
	cp scripts/ratatoskr.service /etc/systemd/system/
	systemctl enable ratatoskr
	systemctl start ratatoskr


setup:
	go install github.com/air-verse/air@latest
	air init

