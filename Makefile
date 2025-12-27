CURRENT_DIRECTORY := $(shell pwd)
TESTS_TO_RUN := $(shell go list ./... | grep -v /integrationTests/ | grep -v /testscommon/ | grep -v mock | grep -v disabled | grep -v defaults)

build:
	go build ./...

clean-test:
	go clean -testcache

test: clean-test
	go test ./...

test-race:
	go test -short -race -v ./...

test-coverage:
	@echo "Running unit tests"
	CURRENT_DIRECTORY=$(CURRENT_DIRECTORY) go test -short -cover -coverprofile=coverage.txt -covermode=atomic -v ${TESTS_TO_RUN}

lint-install:
ifeq (,$(wildcard test -f bin/golangci-lint))
	@echo "Installing golint"
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s
endif

run-lint:
	@echo "Running golint"
	bin/golangci-lint run --max-issues-per-linter 0 --max-same-issues 0 --timeout=2m

lint: lint-install run-lint
