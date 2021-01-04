#! /bin/bash

export BASE_DIR=$( cd "$(dirname "$0")" ; pwd )
export BUILD_TAGS=""

DOCKER_IMAGE="edge-orchestration"
BINARY_FILE="edge-orchestration"

PKG_LIST=(
        "common/errormsg"
        "common/errors"
        "common/logmgr"
        "common/networkhelper"
        "common/networkhelper/detector"
        "common/resourceutil"
        "common/resourceutil/cpu"
        "common/resourceutil/native"
        "common/resourceutil/types/servicemgrtypes"
        "common/sigmgr"
        "common/types/configuremgrtypes"
        "common/types/servicemgrtypes"
        "controller/configuremgr"
        "controller/configuremgr/container"
        "controller/configuremgr/native"
        "controller/discoverymgr"
        "controller/discoverymgr/wrapper"
        "controller/scoringmgr"
        "controller/securemgr/verifier"
        "controller/securemgr/authenticator"
        "controller/securemgr/authorizer"
        "controller/servicemgr"
        "controller/servicemgr/executor"
        "controller/servicemgr/executor/androidexecutor"
        "controller/servicemgr/executor/containerexecutor"
        "controller/servicemgr/executor/nativeexecutor"
        "controller/servicemgr/notification"
        "controller/discoverymgr/mnedc"
        "controller/discoverymgr/mnedc/client"
        "controller/discoverymgr/mnedc/connectionutil"
        "controller/discoverymgr/mnedc/server"
        "controller/discoverymgr/mnedc/tunmgr"
        "controller/storagemgr"
        "controller/storagemgr/storagedriver"
        "db/bolt/common"
        "db/bolt/configuration"
        "db/bolt/network"
        "db/bolt/resource"
        "db/bolt/service"
        "db/bolt/system"
        "db/bolt/wrapper"
        "db/helper"
        "interfaces/capi"
        "interfaces/javaapi"
        "orchestrationapi"
        "restinterface"
        "restinterface/cipher"
        "restinterface/cipher/dummy"
        "restinterface/cipher/sha256"
        "restinterface/client"
        "restinterface/client/restclient"
        "restinterface/externalhandler"
        "restinterface/internalhandler"
        "restinterface/resthelper"
        "restinterface/route"
)

export CONTAINER_VERSION="coconut"
export BUILD_DATE=$(date +%Y%m%d.%H%M)

function set_secure_option() {
    echo ""
    echo "-----------------------------------"
    echo " Set tags for secure build"
    echo "-----------------------------------"
    export BUILD_TAGS="secure"
}

function set_mnedc_server_option() {
    echo ""
    echo "-----------------------------------"
    echo " Set tags for start mnedc server"
    echo "-----------------------------------"
    if [ "$1" == "secure" ]; then
        export BUILD_TAGS="securemnedcserver"
    else 
        export BUILD_TAGS="mnedcserver"
    fi
}

function set_mnedc_client_option() {
    echo ""
    echo "-----------------------------------"
    echo " Set tags for start mnedc client"
    echo "-----------------------------------"
    if [ "$1" == "secure" ]; then
        export BUILD_TAGS="securemnedcclient"
    else 
        export BUILD_TAGS="mnedcclient"
    fi
}

function go_mod_vendor() {
    echo ""
    echo "-----------------------------------"
    echo " Go Mod Vendor"
    echo "-----------------------------------"

    VENDOR_DIR='vendor'
    make go-vendor
    ###apply patch of zeroconf
    git apply --directory=$VENDOR_DIR/github.com/grandcat/zeroconf $BASE_DIR/src/controller/discoverymgr/wrapper/zeroconfEdgeOrchestration.patch

}

function build_clean() {
    echo ""
    echo "-----------------------------------"
    echo " Build clean"
    echo "-----------------------------------"
    make clean
}

function build_binaries() {
    echo ""
    echo ""
    echo "**********************************"
    echo " Target Binary arch is "$1
    echo "**********************************"
    case $1 in
        x86)
            export GOARCH=386
            export ARCH=x86
            ;;
        x86_64)
            export GOARCH=amd64
            export ARCH=x86-64
            ;;
        arm)
            export GOARCH=arm GOARM=7
            export ARCH=arm
            ;;
        arm64)
            export GOARCH=arm64
            export ARCH=aarch64
            ;;
        *)
            echo "Target arch isn't supported" && exit 1
            ;;
    esac

    build_binary
}

function build_binary() {
    echo ""
    echo "----------------------------------------"
    echo " Create Executable binary from GoMain"
    echo "----------------------------------------"

    export GOPATH=$BASE_DIR/GoMain:$GOPATH
    make build-binary || exit 1
}

function build_object() {
    echo ""
    echo "----------------------------------------"
    echo " Create Static object of Orchestration"
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
        echo ""
        echo "---------------------------"
        echo " build test for ALL pkgs"
        echo "---------------------------"
        make test-go TEST_PKG_DIRS=./src/...
    else
        idx=0
        for pkg in "${PKG_LIST[@]}"; do
            if [[ "$pkg" == *"$1"* ]]; then
                break
            fi
            idx=$((idx+1))
        done
        if [ $idx -ge ${#PKG_LIST[@]} ]; then
            echo ""
            echo " ERROR!!! There is no package for $1"
        else
            echo "---------------------------------------"
            echo " build test for ${PKG_LIST[$idx]}"
            echo "---------------------------------------"
            make test-go TEST_PKG_DIRS=./src/${PKG_LIST[$idx]}
        fi
    fi
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

    export GOPATH=$BASE_DIR/GoMain:$GOPATH
    go-callvis -http localhost:7010 -group pkg,type -nostd ./GoMain/src/main/main.go &
}

function build_objects() {
    case $1 in
        x86)
            build_object_x86
            export ANDROID_TARGET="android/386";;
        x86_64)
            build_object_x86-64
            export ANDROID_TARGET="android/amd64";;
        arm)
            build_object_arm
            export ANDROID_TARGET="android/arm";;
        arm64)
            build_object_aarch64
            export ANDROID_TARGET="android/arm64";;
        *)
            build_object_x86
            build_object_x86-64
            build_object_arm
            build_object_aarch64
            export ANDROID_TARGET="android";;
    esac

    build_android
}


function build_object_x86() {
    echo ""
    echo ""
    echo "**********************************"
    echo " Target Binary arch is i386 "
    echo "**********************************"
    export GOARCH=386
    export CC="gcc"
    export ARCH=x86

    build_object
}

function build_object_x86-64() {
    echo ""
    echo ""
    echo "**********************************"
    echo " Target Binary arch is amd64 "
    echo "**********************************"
    export GOARCH=amd64
    export CC="gcc"
    export ARCH=x86-64

    build_object
}

function build_object_arm() {
    echo ""
    echo ""
    echo "**********************************"
    echo " Target Binary arch is armv7 "
    echo "**********************************"
    export GOARCH=arm GOARM=7
    export CC="arm-linux-gnueabi-gcc"
    export ARCH=arm

    build_object
}

function build_object_aarch64() {
    echo ""
    echo ""
    echo "**********************************"
    echo " Target Binary arch is arm64 "
    echo "**********************************"
    export GOARCH=arm64
    export CC="aarch64-linux-gnu-gcc"
    export ARCH=aarch64

    build_object
}

function build_object_result() {
    echo ""
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
    echo ""
    echo "**********************************"
    echo " Create Docker container "
    echo "**********************************"

    docker rm -f $DOCKER_IMAGE
    docker rmi -f $DOCKER_IMAGE:$CONTAINER_VERSION
    mkdir -p $BASE_DIR/GoMain/bin/qemu
    case $1 in
        x86)
            CONTAINER_ARCH="i386"
            ;;
        x86_64)
            CONTAINER_ARCH="amd64"
            ;;
        arm)
            CONTAINER_ARCH="arm32v7"
            cp /usr/bin/qemu-arm-static $BASE_DIR/GoMain/bin/qemu
            ;;
        arm64)
            CONTAINER_ARCH="arm64v8"
            cp /usr/bin/qemu-aarch64-static $BASE_DIR/GoMain/bin/qemu
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
    sudo mkdir -p /var/edge-orchestration/log
    sudo mkdir -p /var/edge-orchestration/apps
    sudo mkdir -p /var/edge-orchestration/data/db
    sudo mkdir -p /var/edge-orchestration/data/cert
    sudo mkdir -p /var/edge-orchestration/user
    sudo mkdir -p /var/edge-orchestration/device
    sudo mkdir -p /var/edge-orchestration/mnedc
    sudo mkdir -p /var/edge-orchestration/datastorage

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


case "$1" in
    "container")
        go_mod_vendor
        if [ "$2" == "secure" ]; then
            set_secure_option
            build_binaries $3
            build_docker_container $3
            docker save -o $BASE_DIR/GoMain/bin/edge-orchestration.tar edge-orchestration
            if [ "$3" == "x86_64" ]; then
                run_docker_container
            fi
        else
            build_binaries $2
            build_docker_container $2
            docker save -o $BASE_DIR/GoMain/bin/edge-orchestration.tar edge-orchestration
            if [ "$2" == "x86_64" ]; then
                run_docker_container
            fi
        fi
        ;;
    "object")
        case "$2" in
            "secure")
                if [ "$3" == "mnedcserver" ]; then
                    set_mnedc_server_option $2
                elif [ "$3" == "mnedcclient" ]; then
                    set_mnedc_client_option $2
                elif [ "$3" == "" ]; then
                    set_secure_option
                fi
                ;;
            "mnedcserver")
                set_mnedc_server_option $3
                ;;
            "mnedcclient")
                set_mnedc_client_option $3
                ;;
        esac
        go_mod_vendor
        build_clean
        if [ "$2" == "secure" ]; then
            build_objects $3
        else
            build_objects $2
        fi
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
    "")
        go_mod_vendor
        build_binary
        build_docker_container
        run_docker_container
        ;;
    "secure")
        set_secure_option
        go_mod_vendor
        build_binary
        build_docker_container
        run_docker_container
        ;;
    "mnedcserver")
        set_mnedc_server_option $2
        go_mod_vendor
        build_binary
        build_docker_container
        run_docker_container
        ;;
    "mnedcclient")
        set_mnedc_client_option $2
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
        echo "  $0 test [PKG_NAME]         : run unittests (optional for PKG_NAME)"
        echo "-------------------------------------------------------------------------------------------------------------------------------------------"
        exit 0
        ;;
esac
