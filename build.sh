#! /bin/bash

export BASE_DIR=$( cd "$(dirname "$0")" ; pwd )
export BUILD_DATE=$(date +%Y%m%d.%H%M)
export BUILD_TAGS=""
export CONTAINER_VERSION="coconut"

DOCKER_IMAGE="edge-orchestration"
PARAMS=$@

function set_options() {
    echo ""
    echo "-----------------------------------"
    echo " Set build tag and options"
    echo "-----------------------------------"

    ANDROID_TARGET="android"
    for i in $@; do
        if [ $i == "secure" ]; then
            BUILD_TAGS="secure"$BUILD_TAGS
        fi
        if [ $i == "mnedcserver" ]; then
            BUILD_TAGS="mnedcserver"
        fi
        if [ $i == "mnedcclient" ]; then
            BUILD_TAGS="mnedcclient"
        fi
        if [ $i == "x86" ]; then
            ARCH="x86"
            CC="gcc"
            GOARCH="386"
            ANDROID_TARGET="android/386"
            CONTAINER_ARCH="i386"
        fi
        if [ $i == "x86_64" ]; then
            ARCH="x86-64"
	    CC="gcc"
            GOARCH="amd64"
            ANDROID_TARGET="android/amd64"
            CONTAINER_ARCH="amd64"
        fi
        if [ $i == "arm" ]; then
            ARCH="arm"
            CC="arm-linux-gnueabi-gcc"
            GOARM="7"
            GOARCH="arm"
            ANDROID_TARGET="android/arm"
            CONTAINER_ARCH="arm32v7"
        fi
        if [ $i == "arm64" ]; then
            ARCH="aarch64"
            CC="aarch64-linux-gnu-gcc"
            GOARCH="arm64"
            ANDROID_TARGET="android/arm64"
            CONTAINER_ARCH="arm64v8"
        fi
    done
    export BUILD_TAGS=$BUILD_TAGS
}

function go_mod_vendor() {
    echo ""
    echo "-----------------------------------"
    echo " Go Mod Vendor"
    echo "-----------------------------------"
    make go-vendor
}

function build_clean() {
    echo ""
    echo "-----------------------------------"
    echo " Build clean"
    echo "-----------------------------------"
    make clean
}

function build_binary() {
    echo ""
    echo "----------------------------------------"
    echo " Create Executable binary"
    echo "----------------------------------------"

    case $GOARCH in
        386 | amd64 | arm64)
            export GOARCH=$GOARCH CC=$CC ARCH=$ARCH;;
        arm)
            export GOARCH=arm GOARM=7 CC="arm-linux-gnueabi-gcc" ARCH=arm;;
    esac
    export GOPATH=$BASE_DIR/bin:$GOPATH
    make build-binary || exit 1
}

function build_object() {
    echo ""
    echo "----------------------------------------"
    echo " Create Static object of Orchestration for $ARCH"
    echo "----------------------------------------"
    make build-object-c || exit 1
}

function build_test() {
    echo ""
    echo "-------------------------------------"
    echo " Build test to calculate Coverage"
    echo "-------------------------------------"
    export GOARCH=amd64
    export CC="gcc"

    sudo systemctl stop ${SERVICE_FILE}
    sudo systemctl status ${SERVICE_FILE}
    mkdir -p /tmp/foo

    stop_docker_container

    if [[ $1 == "" ]]; then
        DIRS=./internal/...
    else
        DIRS=$1
    fi
    echo "---------------------------------------"
    echo " build test for $DIRS"
    echo "---------------------------------------"
    make test-go TEST_PKG_DIRS=$DIRS
}

function lint_src_code() {
    echo ""
    echo "---------------------------------------"
    echo " Analysis source code golint & go vet"
    echo "---------------------------------------"
    make lint
}

function draw_callvis() {
    echo ""
    echo "**********************************"
    echo " Draw Call Graph "
    echo " 1) Install go-callvis"
    echo "   $ go get -u github.com/TrueFurby/go-callvis"
    echo "   $ cd \$GOPATH/src/github.com/TrueFurby/go-callvis && make"
    echo " 2) To use this, do comment last line of build.sh"
    echo " ex) #build_clean_vendor "
    echo "**********************************"

    export GOPATH=$BASE_DIR/bin:$GOPATH
    go-callvis -http localhost:7010 -group pkg,type -nostd ./cmd/edge-orchestration/main.go &
}

function build_objects() {
    case $GOARCH in
        386 | amd64 | arm64)
            export GOARCH=$GOARCH CC=$CC ARCH=$ARCH
            build_object;;
        arm)
            export GOARCH=arm GOARM=7 CC="arm-linux-gnueabi-gcc" ARCH=arm
            build_object;;
        *)
            export GOARCH=386 CC="gcc" ARCH=x86
            build_object
            export GOARCH=amd64 CC="gcc" ARCH=x86-64
            build_object
            export GOARCH=arm GOARM=7 CC="arm-linux-gnueabi-gcc" ARCH=arm
            build_object
            export GOARCH=arm64 CC="aarch64-linux-gnu-gcc" ARCH=aarch64
            build_object
            ;;
    esac
    export ANDROID_TARGET=$ANDROID_TARGET
    build_android
}

function build_object_result() {
    echo ""
    echo "**********************************"
    echo " Edge-orchestration Archive "
    echo "**********************************"

    make build-result
}

function build_android() {
    echo ""
    echo "**********************************"
    echo " Target Binary is for Android "
    echo "**********************************"
    echo ""
    echo "-------------------------------------------"
    echo " Create Android archive from Java interface"
    echo "-------------------------------------------"

    make build-object-java || exit 1
}

function build_docker_container() {
    echo ""
    echo "**********************************"
    echo " Create Docker container $ARCH"
    echo "**********************************"

    docker rm -f $DOCKER_IMAGE
    docker rmi -f $DOCKER_IMAGE:$CONTAINER_VERSION
    mkdir -p $BASE_DIR/bin/qemu
    case $GOARCH in
        386 | amd64)
            ;;
        arm)
            cp /usr/bin/qemu-arm-static $BASE_DIR/bin/qemu
            ;;
        arm64)
            cp /usr/bin/qemu-aarch64-static $BASE_DIR/bin/qemu
            ;;
        *)
            case "$(uname -m)" in
                "i386"|"i686")
                    CONTAINER_ARCH="i386"
                    ;;
                "x86_64")
                    CONTAINER_ARCH="amd64"
                    ;;
                "armv7l")
                    CONTAINER_ARCH="arm32v7"
                    ;;
                "aarch64")
                    CONTAINER_ARCH="arm64v8"
                    ;;
                *)
                    echo "Target arch isn't supported" && exit 1
                    ;;
            esac
            ;;
    esac

    make build-container CONTAINER_ARCH=$CONTAINER_ARCH || exit 1
}

function run_docker_container() {
    ### Add temporarily, TO DO: move files into container
    echo ""
    echo "--------------------------------------------"
    echo "  Create prerequisite Folder [SuperUser]"
    echo "--------------------------------------------"
    sudo ./tools/create_fs.sh
    echo ""
    echo "**********************************"
    echo " Run Docker container "
    echo "**********************************"
    docker run -it -d \
                --privileged \
                --network="host" \
                --name $DOCKER_IMAGE \
                -v /var/edge-orchestration/:/var/edge-orchestration/:rw \
                -v /var/run/docker.sock:/var/run/docker.sock:rw \
                -v /proc/:/process/:ro \
                $DOCKER_IMAGE:$CONTAINER_VERSION
    docker container ls
}

function stop_docker_container() {
    echo ""
    echo "**********************************"
    echo " Stop Docker container "
    echo "**********************************"
    docker stop $DOCKER_IMAGE
    docker ps -a
}

set_options $PARAMS

case "$1" in
    "container")
        go_mod_vendor
        build_binary
        build_docker_container
        docker save -o $BASE_DIR/bin/edge-orchestration.tar edge-orchestration
        ;;
    "object")
        build_clean
        go_mod_vendor
        build_objects
        build_object_result
        ;;
    "test")
        go_mod_vendor
        build_test $2
        ;;
    "lint")
        go_mod_vendor
        lint_src_code
        ;;
    "callvis")
        go_mod_vendor
        draw_callvis
        ;;
    "clean")
        build_clean
        ;;
    "" | "secure" | "mnedcserver" | "mnedcclient")
        go_mod_vendor
        build_binary
        build_docker_container
        run_docker_container
        ;;
    *)
        echo "build script"
        echo "Usage:"
        echo "-------------------------------------------------------------------------------------------------------------------------------------------"
        echo "  $0                         : build edge-orchestration by default Docker container"
        echo "  $0 secure                  : build edge-orchestration by default Docker container with secure option"
        echo "  $0 container [Arch]        : build Docker container Arch:{x86, x86_64, arm, arm64}"
        echo "  $0 container secure [Arch] : build Docker container  with secure option Arch:{x86, x86_64, arm, arm64}"
        echo "  $0 object [Arch]           : build object (c-object, java-object), Arch:{x86, x86_64, arm, arm64} (default:all)"
        echo "  $0 object secure [Arch]    : build object (c-object, java-object) with secure option, Arch:{x86, x86_64, arm, arm64} (default:all)"
        echo "  $0 mnedcserver             : build edge-orchestration by default container with MNEDC server running option"
        echo "  $0 mnedcserver secure      : build edge-orchestration by default container with MNEDC server running option in secure mode"
        echo "  $0 mnedcclient             : build edge-orchestration by default container with MNEDC client running option"
        echo "  $0 mnedcclient secure      : build edge-orchestration by default container with MNEDC client running option in secure mode"
        echo "  $0 clean                   : build clean"
        echo "  $0 test [PKG_PATH]         : run unittests (optional for PKG_PATH which is a relative path such as './internal/common/commandvalidator')"
        echo "-------------------------------------------------------------------------------------------------------------------------------------------"
        exit 0
        ;;
esac
