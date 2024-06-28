# Go commands
GOCMD 	= go
GOCLEAN = $(GOCMD) clean
GODEPS 	= $(GOCMD) mod download
GOTEST 	= $(GOCMD) test
GOBUILD = $(GOCMD) build

# Filepaths
TEST_FOLDER 	= test
BUILD_FOLDER	= bin
COVER_PKG 		= bot
BINARY_NAME 	= $(BUILD_FOLDER)/watchlist
COVERAGE_OUT 	= $(BUILD_FOLDER)/coverage.out
COVERAGE_HTML 	= $(BUILD_FOLDER)/coverage.html


# Default target
default: clean deps build

# Clean target
clean:
	@$(GOCLEAN)
	@rm -rf $(BUILD_FOLDER)

# Install dependencies
deps:
	@$(GODEPS)

# Test target
test:
	@$(GOTEST) ./$(TEST_FOLDER) -v -coverpkg=./$(COVER_PKG) -coverprofile=$(COVERAGE_OUT) ./...
	@go tool cover -html=$(COVERAGE_OUT) -o $(COVERAGE_HTML)

# Build target
build:
	@CGO_ENABLED=1 $(GOBUILD) -o $(BINARY_NAME)

# Development target with hot reloading
dev:
	@air -c .air.toml
 