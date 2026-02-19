.PHONY: build run clean test fmt lint

# Build the tornado binary
build:
	go build -o bin/tornado ./cmd/tornado

# Run the application directly
run:
	go run ./cmd/tornado

# Run in debug mode (logs to debug.log)
run-debug:
	DEBUG=1 go run ./cmd/tornado

# Clean build artifacts
clean:
	rm -rf bin/

# Run tests
test:
	go test -v ./...

# Format code
fmt:
	go fmt ./...

# Run linter (requires golangci-lint installed)
lint:
	golangci-lint run ./...

# Download dependencies
deps:
	go mod download
	go mod tidy

# Watch and run on file changes (requires air installed)
watch:
	air
