.PHONY: default setup lint format add-license-headers vet check compile install test test-fast update-dependencies

default: test

setup:
	@echo "Installing build tools..."
	@go install github.com/google/addlicense@v1.1.1
	@go install gotest.tools/gotestsum@v1.10.1

lint: setup
ifeq ($(OS),Windows_NT)
	@.\tools\check-format.cmd
else
	@go install github.com/google/addlicense@v1
	@sh ./tools/check-format.sh
	@sh ./tools/check-license-headers.sh
	@go mod tidy
endif

format:
	@gofmt -w .

add-license-headers:
ifeq ($(OS),Windows_NT)
	@echo "This is currently not supported on windows"
	@exit 1
else
	@sh ./tools/add-missing-license-headers.sh
endif

vet:
	@echo "Vetting files"
	@go vet -tags '!unit' ./...

check:
	@echo "Static code analysis"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1
	@golangci-lint run ./...

compile:
	@echo "Compiling sources..."
	@go build ./...
	@echo "Compiling tests..."
	@go test -run "NON_EXISTENT_TEST_TO_ENSURE_NOTHING_RUNS_BUT_ALL_COMPILE" ./...

install:
	@go install ./...

test: setup
	@echo "Testing $(BINARY_NAME)..."
	@gotestsum ${testopts} -- -v -race ./...

test-fast: setup
	@echo "Testing short tests $(BINARY_NAME)..."
	@gotestsum ${testopts} -- -v -race -short ./...

update-dependencies:
	@echo Update go dependencies
	@go get -u ./...
	@go mod tidy
