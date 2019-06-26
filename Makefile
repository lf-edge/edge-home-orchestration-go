# Go parameters
GOCMD       := go
GOBUILD     := $(GOCMD) build
GOCLEAN     := $(GOCMD) clean
GOLINT		:= golint
GOVET		:= $(GOCMD) vet
GOCOVER     := gocov
DOCKER		:= docker
GO_COMMIT_ID:= $(shell git rev-parse --short HEAD)
GO_LDFLAGS  := -ldflags '-extldflags "-static" -X main.version=$(VERSION) -X main.commitID=$(GO_COMMIT_ID) -X main.buildTime=$(BUILD_DATE)'

# GoMain target
PKG_NAME		:= edge-orchestration
GOMAIN_DIR		:= $(BASE_DIR)/GoMain
GOMAIN_BIN_DIR	:= $(GOMAIN_DIR)/bin
GOMAIN_SRC_DIR	:= $(GOMAIN_DIR)/src
EXEC_SRC_DIR    := main
GOMAIN_BIN_FILE	:= $(PKG_NAME)

# Go 3rdParty packages
BUILD_VENDOR_DIR	:= $(BASE_DIR)/vendor/
GLIDE_LOCK_FILE		:= $(BASE_DIR)/glide.lock

.DEFAULT_GOAL := help

## edge-orchestration binary build
build-binary:
	$(GOBUILD) -a $(GO_LDFLAGS) -o $(GOMAIN_BIN_DIR)/$(GOMAIN_BIN_FILE) $(EXEC_SRC_DIR) || exit 1
	ls -al $(GOMAIN_BIN_DIR)

## clean 3rdParty packages
clean-tmp-packages:
	-rm -rf $(BUILD_VENDOR_DIR)
	-rm -rf $(GLIDE_LOCK_FILE)

## edge-orchestration container build
build-container:
	$(DOCKER) build --tag $(PKG_NAME):$(CONTAINER_VERSION) --file $(GOMAIN_DIR)/Dockerfile .

## go test and coverage
test-go:
	$(GOCOVER) test -v $(TEST_PKG_DIRS) > coverage.out
	$(GOCOVER) report coverage.out
	$(GOCOVER)-html coverage.out > coverage.html
	firefox coverage.html &

## build clean
clean:
	$(GOCLEAN)

## check go style and static analysis
lint:
	$(GOLINT) ./src/...
	$(GOVET) -v ./src/...

## show help
help:
	@make2help $(MAKEFILE_LIST)

## define build target not a file
.PHONY: all build test clean lint help
