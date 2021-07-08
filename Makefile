BASE_DIR := ${shell pwd | sed -e 's/ /\\ /g'}
ifeq ($(CONFIG_CONTAINER),y)
BASE_DIR := .
endif

# Control build verbosity
#  V=1,2: Enable echo of commands
#  V=2:   Enable bug/verbose options in tools and scripts
ifeq ($(V),1)
export Q :=
else
ifeq ($(V),2)
export Q :=
else
export Q := @
endif
endif

BUILD_DATE=${shell date +%Y%m%d.%H%M}
CONTAINER_VERSION="coconut"

DOCKER_IMAGE="edge-orchestration"

-include $(BASE_DIR)/.config

ifeq ($(CONFIG_MNEDC_SERVER),y)
BUILD_TAGS += mnedcserver
endif

ifeq ($(CONFIG_MNEDC_CLIENT),y)
BUILD_TAGS += mnedcclient
endif

ifeq ($(CONFIG_SECURE_MODE),y)
BUILD_TAGS += secure
endif

# Go parameters
GOCMD		:= GO111MODULE=on go
GOBUILD 	:= $(GOCMD) build
GOCLEAN 	:= $(GOCMD) clean
GOLINT		:= golint
GOVET		:= $(GOCMD) vet
GOCOVER		:= gocov
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

ifeq ($(CONFIG_ARM),y)
	INTERFACE_OUT_INC_DIR		:= $(INTERFACE_OUT_DIR)/inc/linux_arm
	INTERFACE_OUT_BIN_DIR		:= $(INTERFACE_OUT_DIR)/bin/linux_arm
	INTERFACE_OUT_LIB_DIR		:= $(INTERFACE_OUT_DIR)/lib/linux_arm
	CONTAINER_ARCH="arm32v7"
	GOARCH=arm
	CC="arm-linux-gnueabi-gcc"
	GOARM=7
	ANDROID_TARGET="android/arm"
else ifeq ($(CONFIG_ARM64),y)
	INTERFACE_OUT_INC_DIR		:= $(INTERFACE_OUT_DIR)/inc/linux_aarch64
	INTERFACE_OUT_BIN_DIR		:= $(INTERFACE_OUT_DIR)/bin/linux_aarch64
	INTERFACE_OUT_LIB_DIR		:= $(INTERFACE_OUT_DIR)/lib/linux_aarch64
	CONTAINER_ARCH="arm64v8"
	GOARCH=arm64
	CC="aarch64-linux-gnu-gcc"
	ANDROID_TARGET="android/arm64"
else ifeq ($(CONFIG_X86),y)
	INTERFACE_OUT_INC_DIR		:= $(INTERFACE_OUT_DIR)/inc/linux_x86
	INTERFACE_OUT_BIN_DIR		:= $(INTERFACE_OUT_DIR)/bin/linux_x86
	INTERFACE_OUT_LIB_DIR		:= $(INTERFACE_OUT_DIR)/lib/linux_x86
	CONTAINER_ARCH="i386"
	GOARCH=386
	CC="gcc"
	ANDROID_TARGET="android/386"
else ifeq ($(CONFIG_X86_64),y)
	INTERFACE_OUT_INC_DIR		:= $(INTERFACE_OUT_DIR)/inc/linux_x86-64
	INTERFACE_OUT_BIN_DIR		:= $(INTERFACE_OUT_DIR)/bin/linux_x86-64
	INTERFACE_OUT_LIB_DIR		:= $(INTERFACE_OUT_DIR)/lib/linux_x86-64
	CONTAINER_ARCH="amd64"
	GOARCH=amd64
	CC="gcc"
	ANDROID_TARGET="android/amd64"
endif

# Go 3rdParty packages
BUILD_VENDOR_DIR	:= $(BASE_DIR)/vendor

# Go Library for android
ANDROID_LIBRARY_NAME          := liborchestration
ANDROID_LIBRARY_FILE          := liborchestration.aar
ANDROID_JAR_FILE              := liborchestration-sources.jar
ANDROID_SRC_DIR               := $(CMD_DIR)/edge-orchestration/javaapi
ANDROID_LIBRARY_OUT_DIR       := $(BIN_DIR)/javaapi/output

.DEFAULT_GOAL := all

TEST_PKG_DIRS = $(filter-out $@,$(MAKECMDGOALS))
ifeq ($(word 2, $(TEST_PKG_DIRS)),)
TEST_PKG_DIRS="internal/..."
endif

define print_header
	@ echo ""
	@ echo "-----------------------------------"
	@ echo " "$1
	@ echo "-----------------------------------"
endef

define go-vendor
	$(call print_header, "Go Mod Vendor")
	$(GOCMD) mod vendor
endef

## edge-orchestration binary build
build_binary:
	$(call print_header, "Create Executable binary")
	GOARM=$(GOARM) GOARCH=$(GOARCH) $(GOBUILD) -a $(GO_LDFLAGS) -o $(BIN_DIR)/$(BIN_FILE) $(CMD_SRC) || exit 1
	$(Q) ls -al $(BIN_DIR)

## edge-orchestration static archive build
define build-object-c
	$(call print_header, "Create Static object of Orchestration for $(GOARCH)")
	$(Q) mkdir -p $(INTERFACE_OUT_INC_DIR) $(INTERFACE_OUT_LIB_DIR)
	CGO_ENABLED=1 GOARM=$(GOARM) GOARCH=$(GOARCH) CC=$(CC) $(GOBUILD) $(GO_LDFLAGS) -o $(INTERFACE_OUT_LIB_DIR)/$(CUR_LIBRARY_FILE) -buildmode=c-archive $(OBJ_SRC_DIR) || exit 1
	$(Q) mv $(INTERFACE_OUT_LIB_DIR)/$(CUR_HEADER_FILE) $(INTERFACE_OUT_INC_DIR)/$(HEADER_FILE)
	$(Q) ls -al $(INTERFACE_OUT_LIB_DIR)
endef

## edge-orchestration archive
define build-result
	tree $(INTERFACE_OUT_DIR)
	tree $(ANDROID_LIBRARY_OUT_DIR)
endef

## edge-orchestration android library build
build-object-java:
	$(Q) mkdir -p $(ANDROID_LIBRARY_OUT_DIR)
	$(Q) rm -rf $(BUILD_VENDOR_DIR)
	$(GOMOBILE) bind $(GO_MOBILE_LDFLAGS) -o $(ANDROID_LIBRARY_OUT_DIR)/$(ANDROID_LIBRARY_FILE) -target=$(ANDROID_TARGET) -androidapi=23 $(ANDROID_SRC_DIR) || exit 1
	$(Q) ls -al $(ANDROID_LIBRARY_OUT_DIR)

## edge-orchestration container build
build_docker_container:
	$(call print_header, "Create Docker container $(CONTAINER_ARCH)")
	-docker rm -f $(DOCKER_IMAGE)
	-docker rmi -f $(DOCKER_IMAGE):$(CONTAINER_VERSION)
	$(Q) mkdir -p $(BASE_DIR)/bin/qemu
ifeq ($(CONFIG_ARM),y)
	$(Q) cp /usr/bin/qemu-arm-static $(BASE_DIR)/bin/qemu
endif
ifeq ($(CONFIG_ARM64),y)
	$(Q) cp /usr/bin/qemu-aarch64-static $(BASE_DIR)/bin/qemu
endif
	$(DOCKER) build --tag $(PKG_NAME):$(CONTAINER_VERSION) --file $(BASE_DIR)/Dockerfile --build-arg PLATFORM=$(CONTAINER_ARCH) .
	-docker save -o $(BASE_DIR)/bin/edge-orchestration.tar edge-orchestration

## go test and coverage
test-go:
	$(GOCOVER) test -v ./$(TEST_PKG_DIRS) > coverage.out
	$(GOCOVER) report coverage.out
	$(GOCOVER)-html coverage.out > coverage.html
	$(Q) -rm -rf $(BASE_DIR)/internal/controller/discoverymgr/testDB
	$(Q) firefox coverage.html &

## build clean
clean:
	$(call print_header, "Build clean")
	$(Q) $(GOCLEAN)
	$(Q) -rm -rf $(BUILD_VENDOR_DIR)
	$(Q) -rm -rf $(INTERFACE_OUT_DIR)
	$(Q) -rm -rf $(ANDROID_LIBRARY_OUT_DIR)
	$(Q) -rm -rf $(BASE_DIR)/bin/$(PKG_NAME)*
	$(Q) -rm -rf $(BASE_DIR)/coverage.out
	$(Q) make -C examples/native clean

distclean: clean
	$(Q) rm -f .config
	$(Q) rm -f .config.old

## check go style and static analysis
lint:
	$(call print_header, "Analysis source code golint & go vet")
	$(GOLINT) ./internal/...
	$(GOVET) -v ./internal/...

## show help
help:
	@echo 'Usage: make <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@echo '    create_context     Prepare configuration.'
	@echo '    help               Show this help screen.'
	@echo '    clean              Remove binaries, artifacts.'
	@echo '    test               Run unit tests.'
	@echo '    lint               Run golint and go vet.'
	@echo '    all                Build project for current platform.'
	@echo '    menuconfig         Change configuration by kconfig-frontends.'
	@echo ''

## define build target not a file
.PHONY: all build test clean lint help

define stop_docker_container
	$(call print_header, "Stop Docker container")
	-docker stop $(DOCKER_IMAGE)
	$(Q) docker ps -a
endef

##
all:
	make clean
	$(call go-vendor)
ifeq ($(CONFIG_CONTAINER),y)
	make build_docker_container
else ifeq ($(CONFIG_NATIVE),y)
	$(call build-object-c)
	$(call build-result)
else ifeq ($(CONFIG_ANDROID),y)
	$(call print_header, "Target Binary is for Android")
	$(call print_header, "Create Android archive from Java interface")
	make build-object-java
	$(call build-result)
endif

.config:
	$(Q) cp configs/defconfigs/$(CONFIGURATION_FILE_NAME) .config

create_context: .config

run:
	$(call print_header, "Run Docker container ")
	docker run -it -d \
                --privileged \
                --network="host" \
                --name $(DOCKER_IMAGE) \
                -v /var/edge-orchestration/:/var/edge-orchestration/:rw \
                -v /var/run/docker.sock:/var/run/docker.sock:rw \
                -v /proc/:/process/:ro \
                $(DOCKER_IMAGE):$(CONTAINER_VERSION)
	$(Q) docker container ls

test:
	$(call print_header, "Build test to calculate Coverage")
	$(Q) -sudo systemctl stop ${SERVICE_FILE}
	$(Q) -sudo systemctl status ${SERVICE_FILE}
	$(Q) mkdir -p /tmp/foo
	$(call stop_docker_container)
	$(call print_header, "build test for $(TEST_PKG_DIRS)")
	make test-go TEST_PKG_DIRS=$(TEST_PKG_DIRS)

do_menuconfig:
	$(Q) kconfig-mconf Kconfig

menuconfig: do_menuconfig

do_savedefconfig:
	$(Q) cp -f .config configs/defconfigs/$(CONFIG_CONFIGURATION_FILE_NAME)

savedefconfig: do_savedefconfig
