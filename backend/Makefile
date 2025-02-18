# Run all checks
all: fmt-check test build

# Check if code is formatted
fmt-check:
	@unformatted_files=$$(gofmt -l .); \
	if [ -n "$$unformatted_files" ]; then \
		echo "The following files are not properly formatted:"; \
		echo "$$unformatted_files"; \
		exit 1; \
	else \
		echo "All files are formatted correctly!"; \
	fi

# Format code 
fmt:
	gofmt -w .

# Run tests
test:
	go test ./...

# Build to verify compilation
build:
	go build ./cmd/main.go
