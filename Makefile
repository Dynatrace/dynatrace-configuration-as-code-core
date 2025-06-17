.PHONY: default setup lint format add-license-headers vet check compile test test-fast update-dependencies

default: test

setup:
	@echo "Installing build tools..."
	@go install github.com/google/addlicense@v1.1.1
	@go install gotest.tools/gotestsum@v1.10.1
	@go install go.uber.org/mock/mockgen@v0.4

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

vet: generate-mocks
	@echo "Vetting files"
	@go vet -tags '!unit' ./...

generate-mocks:
	@echo "Generating mocks"
	@go install go.uber.org/mock/mockgen@v0.4
	@go generate ./...

compile: generate-mocks
	@echo "Compiling sources..."
	@go build ./...
	@echo "Compiling tests..."
	@go test -tags e2e -run "NON_EXISTENT_TEST_TO_ENSURE_NOTHING_RUNS_BUT_ALL_COMPILE" ./...

test: setup generate-mocks
	@echo "Testing $(BINARY_NAME)..."
	@gotestsum ${testopts} -- -v -race ./...

test-fast: setup
	@echo "Testing short tests $(BINARY_NAME)..."
	@gotestsum ${testopts} -- -v -race -short ./...

update-dependencies:
	@echo Update go dependencies
	@go get -u ./...
	@go mod tidy
