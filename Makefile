BASE_DIR := ${shell pwd | sed -e 's/ /\\ /g'}

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

VERSION := v$(shell cat VERSION)
CONTAINER_VERSION="latest"
DOCKER_IMAGE="lfedge/edge-home-orchestration-go"

-include $(BASE_DIR)/.config

ifeq ($(CONFIG_MNEDC_SERVER),y)
RUN_OPTIONS += -e MNEDC=server
endif

ifeq ($(CONFIG_MNEDC_CLIENT),y)
RUN_OPTIONS += -e MNEDC=client
endif

ifeq ($(CONFIG_SECURE_MODE),y)
RUN_OPTIONS += -e SECURE=true
endif

ifeq ($(CONFIG_CLOUD_SYNC),y)
RUN_OPTIONS += -e CLOUD_SYNC=true
endif

ifeq ($(CONFIG_WEB_UI),y)
RUN_OPTIONS += -e WEBUI=true
endif

ifeq ($(CONFIG_LOGLEVEL),y)
RUN_OPTIONS += -e LOGLEVEL=$(subst ",,$(CONFIG_LOGLEVEL_VALUE))
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
GO_LDFLAGS		:= -ldflags '-extldflags "-static" -X main.version=$(VERSION) -X main.commitID=$(GO_COMMIT_ID)'
GO_MOBILE_LDFLAGS	:= -ldflags '-X main.version=$(VERSION) -X main.commitID=$(GO_COMMIT_ID)'

# Target parameters
PKG_NAME	:= edge-orchestration

# Go Application target
CMD_DIR 	:= $(BASE_DIR)/cmd
CMD_SRC 	:= $(CMD_DIR)/edge-orchestration/main.go
BIN_DIR 	:= $(BASE_DIR)/bin
BIN_FILE	:= $(PKG_NAME)
WEB_DIR 	:= $(BASE_DIR)/web

# Go Library for C-archive
LIBRARY_NAME		:= liborchestration
HEADER_FILE			:= orchestration.h
LIBRARY_FILE		:= liborchestration.a
CUR_HEADER_FILE 	:= $(LIBRARY_NAME).h
CUR_LIBRARY_FILE 	:= $(LIBRARY_NAME).a
OBJ_SRC_DIR			:= $(CMD_DIR)/edge-orchestration/capi
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
	@ echo "--------------------------------------"
	@ echo " "$1
	@ echo "--------------------------------------"
endef

define go-vendor
	$(call print_header, "Go Mod Vendor")
	$(GOCMD) mod vendor
endef

## edge-orchestration binary build
define build_binary
	$(call print_header, "Create Executable binary")
	GOARM=$(GOARM) GOARCH=$(GOARCH) $(GOBUILD) -a $(GO_LDFLAGS) -o $(BIN_DIR)/$(BIN_FILE) $(CMD_SRC) || exit 1
	$(Q) ls -al $(BIN_DIR)
endef

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
	$(Q) go get golang.org/x/mobile@v0.0.0-20201217150744-e6ae53a27f4f
	$(GOMOBILE) bind $(GO_MOBILE_LDFLAGS) -o $(ANDROID_LIBRARY_OUT_DIR)/$(ANDROID_LIBRARY_FILE) -target=$(ANDROID_TARGET) -androidapi=23 $(ANDROID_SRC_DIR) || exit 1
	$(Q) go mod tidy
	$(Q) ls -al $(ANDROID_LIBRARY_OUT_DIR)

## webui build
build-webui:
	$(call print_header, "Build web contents")
	cd web && npm install && npm run build
	$(Q) ls -al $(WEB_DIR)

## edge-orchestration container build
build_docker_container:
	$(call print_header, "Create Docker container $(CONTAINER_ARCH)")
	-docker rm -f $(PKG_NAME)
	-docker rmi -f $(DOCKER_IMAGE):$(CONTAINER_VERSION)
	$(Q) mkdir -p $(BASE_DIR)/bin/qemu
ifeq ($(CONFIG_ARM),y)
ifneq ($(shell uname -m),armv7l)
	$(Q) cp /usr/bin/qemu-arm-static $(BASE_DIR)/bin/qemu
endif
endif
ifeq ($(CONFIG_ARM64),y)
ifneq ($(shell uname -m),aarch64)
	$(Q) cp /usr/bin/qemu-aarch64-static $(BASE_DIR)/bin/qemu
endif
endif
	$(DOCKER) build --tag $(DOCKER_IMAGE):$(CONTAINER_VERSION) --file $(BASE_DIR)/Dockerfile --build-arg PLATFORM=$(CONTAINER_ARCH) .
	-docker save -o $(BASE_DIR)/bin/edge-orchestration.tar $(DOCKER_IMAGE)

## go test and coverage
test-go:
	$(GOCOVER) test -v ./$(TEST_PKG_DIRS) > coverage.out
	$(GOCOVER) report coverage.out
	$(GOCOVER)-html coverage.out > coverage.html
	$(Q) -rm -rf $(BASE_DIR)/internal/controller/discoverymgr/testDB
	$(Q) firefox coverage.html &

## build clean
clean: go.sum
	$(call print_header, "Build clean")
	$(Q) -rm -rf $(BUILD_VENDOR_DIR)
	$(Q) $(GOCMD) mod tidy
	$(Q) $(GOCLEAN)
	$(Q) -rm -rf $(INTERFACE_OUT_DIR)
	$(Q) -rm -rf $(ANDROID_LIBRARY_OUT_DIR)
	$(Q) -rm -rf $(BASE_DIR)/bin/$(PKG_NAME)*
	$(Q) -rm -rf $(BASE_DIR)/coverage.out
	$(Q) -rm -rf $(BASE_DIR)/coverage.html
	$(Q) make -C examples/native clean

distclean: clean
	$(Q) rm -f .config
	$(Q) rm -f .config.old
	$(Q) rm -f Dockerfile

## check go style and static analysis
lint:
	$(call print_header, "Analysis source code golint & go vet")
	$(GOLINT) ./internal/...
	$(GOVET) -v ./internal/...

staticcheck:
	$(Q) -staticcheck ./...

fuzz:
	$(call print_header, "Run Fuzzer (go test -fuzz)")
	$(Q) ./tools/fuzz-all.sh $1

## format go files
fmt:
	$(Q) make clean
	$(call print_header, "Formatting source code using gofmt")
	$(Q) gofmt -s -w ./internal
	$(Q) gofmt -s -w ./cmd

## show help
help:
	@echo 'Usage: make <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@echo '    all                Build project for current platform.'
	@echo '    clean              Remove binaries, artifacts.'
	@echo '    create_context     Prepare configuration.'
	@echo '    fmt                Run: gofmt -s -w ./ .'
	@echo '    fuzz               Run: go test -fuzz .'
	@echo '    help               Show this help screen.'
	@echo '    lint               Run golint and go vet.'
	@echo '    menuconfig         Change configuration by kconfig-frontends.'
	@echo '    run                Run docker container.'
	@echo '    staticcheck        Run staticcheck.'
	@echo '    stop               Stop and remove docker container.'
	@echo '    test               Run unit tests.'
	@echo ''

## define build target not a file
.PHONY: all binary build clean fmt fuzz help lint run staticcheck stop test

define stop_docker_container
	$(call print_header, "Stop Docker container")
	$(Q) if [ ! -z ${shell docker ps -a --format "{{.Names}}" --filter name=^/$(PKG_NAME)} ]; then \
		docker stop $(PKG_NAME) > /dev/null ; \
	fi
	$(Q) docker ps -a
endef

define rm_docker_container
	$(call print_header, "Remove Docker container")
	$(Q) if [ ! -z ${shell docker ps -a --format "{{.Names}}" --filter name=^/$(PKG_NAME)} ]; then \
		docker rm -f $(PKG_NAME) > /dev/null ; \
	fi
	$(Q) docker ps -a
endef

##
all: check_context
	make clean
	$(call go-vendor)
ifeq ($(CONFIG_CONTAINER),y)
	$(Q) cp configs/defdockerfiles/$(CONFIG_DOCKERFILE) Dockerfile
	$(call build_binary)
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

binary: check_context
	make clean
	$(call go-vendor)
ifeq ($(CONFIG_CONTAINER),y)
	$(call build_binary)
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
ifeq ($(CONFIGFILE),)
	$(Q) echo "" ; echo "CONFIGFILE not been specified:"
	$(Q) echo "  make create_context CONFIGFILE=<configfile>" ; echo "" ; exit 1
endif
	$(Q) cp configs/defconfigs/$(CONFIGFILE) .config

create_context: .config

check_context:
	$(Q) if [ ! -e ${BASE_DIR}/.config ]; then \
		echo "" ; echo "Edge-Orchestration has not been configured:" ; \
		echo "  make create_context CONFIGFILE=<configfile>" ; echo "" ; \
		exit 1 ; \
	fi

go.sum:
	$(Q) $(GOCMD) mod tidy

run:
	$(call print_header, "Run Docker container ")
	$(Q) docker run -it -d \
                --privileged \
                --network="host" \
                --name $(PKG_NAME) \
                $(RUN_OPTIONS) \
                -v /var/edge-orchestration/:/var/edge-orchestration/:rw \
                -v /var/run/docker.sock:/var/run/docker.sock:rw \
                -v /proc/:/process/:ro \
                $(DOCKER_IMAGE):$(CONTAINER_VERSION)
	$(Q) docker container ls

stop:
	$(call stop_docker_container)
	$(call rm_docker_container)

test: go.sum
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
	$(Q) cp -f .config configs/defconfigs/$(CONFIG_CONFIGFILE)

savedefconfig: do_savedefconfig

# Building binaries through multi-stage build
buildx_binary:
	$(call print_header, "Create Executable binary through buildx")
	$(GOBUILD) -a $(GO_LDFLAGS) -o $(BIN_DIR)/$(BIN_FILE) $(CMD_SRC) || exit 1
	$(Q) ls -al $(BIN_DIR)
