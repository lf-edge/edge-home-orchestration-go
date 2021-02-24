# Go parameters
GOCMD		:= GO111MODULE=on go
GOBUILD 	:= $(GOCMD) build
GOCLEAN 	:= $(GOCMD) clean
GOLINT		:= golint
GOVET		:= $(GOCMD) vet
GOCOVER     	:= gocov
GOMOBILE	:= gomobile
DOCKER		:= docker
GO_COMMIT_ID	:= $(shell git rev-parse --short HEAD)
GO_LDFLAGS		:= -ldflags '-extldflags "-static" -X main.version=$(VERSION) -X main.commitID=$(GO_COMMIT_ID) -X main.buildTime=$(BUILD_DATE) -X main.buildTags=$(BUILD_TAGS)'
GO_MOBILE_LDFLAGS	:= -ldflags '-X main.version=$(VERSION) -X main.commitID=$(GO_COMMIT_ID) -X main.buildTime=$(BUILD_DATE) -X main.buildTags=$(BUILD_TAGS)'

# Target parameters
PKG_NAME	:= edge-orchestration

# Go Application target
CMD_DIR 	:= $(BASE_DIR)/cmd
CMD_SRC 	:= $(CMD_DIR)/edge-orchestration/main.go
BIN_DIR 	:= $(BASE_DIR)/bin
BIN_FILE	:= $(PKG_NAME)

# Go Library for C-archive
LIBRARY_NAME		:= liborchestration
HEADER_FILE		:= orchestration.h
LIBRARY_FILE		:= liborchestration.a
CUR_HEADER_FILE 	:= $(LIBRARY_NAME).h
CUR_LIBRARY_FILE 	:= $(LIBRARY_NAME).a
OBJ_SRC_DIR		:= $(CMD_DIR)/edge-orchestration/capi
INTERFACE_OUT_DIR	:= $(BIN_DIR)/capi/output
ifeq ($(ARCH), arm)
	INTERFACE_OUT_INC_DIR		:= $(INTERFACE_OUT_DIR)/inc/linux_arm
	INTERFACE_OUT_BIN_DIR		:= $(INTERFACE_OUT_DIR)/bin/linux_arm
	INTERFACE_OUT_LIB_DIR		:= $(INTERFACE_OUT_DIR)/lib/linux_arm
else ifeq ($(ARCH), aarch64)
	INTERFACE_OUT_INC_DIR		:= $(INTERFACE_OUT_DIR)/inc/linux_aarch64
	INTERFACE_OUT_BIN_DIR		:= $(INTERFACE_OUT_DIR)/bin/linux_aarch64
	INTERFACE_OUT_LIB_DIR		:= $(INTERFACE_OUT_DIR)/lib/linux_aarch64
else ifeq ($(ARCH), x86)
	INTERFACE_OUT_INC_DIR		:= $(INTERFACE_OUT_DIR)/inc/linux_x86
	INTERFACE_OUT_BIN_DIR		:= $(INTERFACE_OUT_DIR)/bin/linux_x86
	INTERFACE_OUT_LIB_DIR		:= $(INTERFACE_OUT_DIR)/lib/linux_x86
else
	INTERFACE_OUT_INC_DIR		:= $(INTERFACE_OUT_DIR)/inc/linux_x86-64
	INTERFACE_OUT_BIN_DIR		:= $(INTERFACE_OUT_DIR)/bin/linux_x86-64
	INTERFACE_OUT_LIB_DIR		:= $(INTERFACE_OUT_DIR)/lib/linux_x86-64
endif

# Go 3rdParty packages
BUILD_VENDOR_DIR	:= $(BASE_DIR)/vendor

# Go Library for android
ANDROID_LIBRARY_NAME          := liborchestration
ANDROID_LIBRARY_FILE          := liborchestration.aar
ANDROID_JAR_FILE              := liborchestration-sources.jar
ANDROID_SRC_DIR               := $(CMD_DIR)/edge-orchestration/javaapi
ANDROID_LIBRARY_OUT_DIR       := $(BIN_DIR)/javaapi/output

.DEFAULT_GOAL := help

go-vendor:
	$(GOCMD) mod vendor

## edge-orchestration binary build
build-binary:
	$(GOBUILD) -a $(GO_LDFLAGS) -o $(BIN_DIR)/$(BIN_FILE) $(CMD_SRC) || exit 1
	ls -al $(BIN_DIR)

## edge-orchestration static archive build
build-object-c:
	mkdir -p $(INTERFACE_OUT_INC_DIR) $(INTERFACE_OUT_LIB_DIR)
	CGO_ENABLED=1 $(GOBUILD) $(GO_LDFLAGS) -o $(INTERFACE_OUT_LIB_DIR)/$(CUR_LIBRARY_FILE) -buildmode=c-archive $(OBJ_SRC_DIR) || exit 1
	mv $(INTERFACE_OUT_LIB_DIR)/$(CUR_HEADER_FILE) $(INTERFACE_OUT_INC_DIR)/$(HEADER_FILE)
	ls -al $(INTERFACE_OUT_LIB_DIR)

## edge-orchestration archive
build-result:
	tree $(INTERFACE_OUT_DIR)
	tree $(ANDROID_LIBRARY_OUT_DIR)

## edge-orchestration android library build
build-object-java:
	mkdir -p $(ANDROID_LIBRARY_OUT_DIR)
	rm -rf $(BUILD_VENDOR_DIR)
	$(GOMOBILE) bind $(GO_MOBILE_LDFLAGS) -o $(ANDROID_LIBRARY_OUT_DIR)/$(ANDROID_LIBRARY_FILE) -target=$(ANDROID_TARGET) -androidapi=23 $(ANDROID_SRC_DIR) || exit 1
	ls -al $(ANDROID_LIBRARY_OUT_DIR)

## edge-orchestration container build
build-container:
	$(DOCKER) build --tag $(PKG_NAME):$(CONTAINER_VERSION) --file $(BASE_DIR)/Dockerfile --build-arg PLATFORM=$(CONTAINER_ARCH) .

## go test and coverage
test-go:
	$(GOCOVER) test -v $(TEST_PKG_DIRS) > coverage.out
	$(GOCOVER) report coverage.out
	$(GOCOVER)-html coverage.out > coverage.html
	-rm -rf $(BASE_DIR)/internal/controller/discoverymgr/testDB
	firefox coverage.html &

## build clean
clean:
	$(GOCLEAN)
	-rm -rf $(BUILD_VENDOR_DIR)
	-rm -rf $(INTERFACE_OUT_DIR)
	-rm -rf $(ANDROID_LIBRARY_OUT_DIR)

## check go style and static analysis
lint:
	$(GOLINT) ./internal/...
	$(GOVET) -v ./internal/...

## show help
help:
	@make2help $(MAKEFILE_LIST)

## define build target not a file
.PHONY: all build test clean lint help
