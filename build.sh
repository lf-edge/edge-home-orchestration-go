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
        "common/types/configuremgrtypes"
        "common/types/servicemgrtypes"
        "controller/configuremgr"
        "controller/configuremgr/container"
        "controller/configuremgr/native"
        "controller/discoverymgr"
        "controller/discoverymgr/wrapper"
        "controller/scoringmgr"
        "controller/servicemgr"
        "controller/servicemgr/executor"
        "controller/servicemgr/executor/androidexecutor"
        "controller/servicemgr/executor/containerexecutor"
        "controller/servicemgr/executor/nativeexecutor"
        "controller/servicemgr/notification"
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

export CONTAINER_VERSION="alpha"
export BUILD_DATE=$(date +%Y%m%d.%H%M)

function set_secure_option() {
    echo ""
    echo "-----------------------------------"
    echo " Set tags for secure build"
    echo "-----------------------------------"
    export BUILD_TAGS="secure"
}

function install_3rdparty_packages() {
    echo ""
    echo "-----------------------------------"
    echo " Install 3rd-party packages"
    echo "-----------------------------------"
    ### Install glide on Ubuntu
    ### sudo add-apt-repository -y ppa:masterminds/glide && sudo apt-get update
    ### sudo apt-get install -y glide

    ### glide update
    glide install

    ### Set GOPATH and link of vendor directory
    export GOPATH=$BASE_DIR:$BASE_DIR/vendor:$GOPATH
    ORG_VENDOR_DIR='vendor'
    CUR_VENDOR_DIR='vendor/src'

    if [ ! -d $BASE_DIR/$CUR_VENDOR_DIR ]; then
        ln -s $BASE_DIR/$ORG_VENDOR_DIR $BASE_DIR/$CUR_VENDOR_DIR
    fi

    ###apply patch of zeroconf
    git apply --directory=$ORG_VENDOR_DIR/github.com/grandcat/zeroconf $BASE_DIR/src/controller/discoverymgr/wrapper/zeroconfEdgeOrchestration.patch

    ### Exception case from package dependencies
    rm -rf $BASE_DIR/vendor/github.com/docker/distribution/vendor/github.com/opencontainers
    rm -rf $BASE_DIR/vendor/github.com/docker/docker/vendor/github.com/docker/go-connections/nat
}

function install_prerequisite() {
    echo ""
    echo "-----------------------------------"
    echo " Install prerequisite packages"
    echo "-----------------------------------"
    pkg_list=(
        "github.com/axw/gocov/gocov"
        "github.com/matm/gocov-html"
        "golang.org/x/lint/golint"
        "github.com/Songmu/make2help/cmd/make2help"
        "golang.org/x/mobile/cmd/gomobile"
        "golang.org/x/mobile/cmd/gobind"
    )
    idx=1
    for pkg in "${pkg_list[@]}"; do
        echo -ne "(${idx}/${#pkg_list[@]}) go get -u $pkg"
        go get -u $pkg
        if [ $? -ne 0 ]; then
            echo -e "\n\033[31m"download fail"\033[0m"
            exit 1
        fi
        echo ": Done"
        idx=$((idx+1))
    done

    # Rebase gomobile [ Needed due to issues in latest gomobile ]
    cd $GOPATH/src/golang.org/x/mobile
    git reset --hard 30c70e3810e97d051f18b66d59ae242540c0c391
    go install ./cmd/...
    cd $BASE_DIR
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
            make test-go TEST_PKG_DIRS=${PKG_LIST[$idx]}
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

function build_clean_vendor() {
    echo ""
    echo "-------------------------------------"
    echo " Clean up 3rdParty directory"
    echo "-------------------------------------"
    make clean-tmp-packages
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
    make build-container || exit 1
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
        if [ "$2" == "secure" ]; then
            set_secure_option
        fi
        install_prerequisite
        install_3rdparty_packages
        build_binary
        build_docker_container
        run_docker_container
        ;;
    "object")
        if [ "$2" == "secure" ]; then
            set_secure_option
        fi
        install_prerequisite
        install_3rdparty_packages
        build_clean
        build_android
        build_object_x86
        build_object_aarch64
        build_object_arm
        build_object_x86-64
        build_object_result
        ;;
    "test")
        install_prerequisite
        install_3rdparty_packages
        build_test $2
        ;;
    "lint")
        install_prerequisite
        install_3rdparty_packages
        lint_src_code
        ;;
    "callvis")
        install_3rdparty_packages
        draw_callvis
        ;;
    "clean")
        build_clean
        ;;
    "")
        install_prerequisite
        install_3rdparty_packages
        build_binary
        build_docker_container
        run_docker_container
        ;;
    "secure")
        set_secure_option
        install_prerequisite
        install_3rdparty_packages
        build_binary
        build_docker_container
        run_docker_container
        ;;
    *)
        echo "build script"
        echo "Usage:"
        echo "----------------------------------------------------------------------------------------------"
        echo "  $0                  : build edge-orchestration by default container"
        echo "  $0 secure           : build edge-orchestration by default container with secure option"
        echo "  $0 container        : build Docker container as build system environmet"
        echo "  $0 container secure : build Docker container as build system environmet with secure option"
        echo "  $0 object           : build object (c-object, java-object)"
        echo "  $0 object secure    : build object (c-object, java-object) with secure option"
        echo "  $0 clean            : build clean"
        echo "  $0 test [PKG_NAME]  : run unittests (optional for PKG_NAME)"
        echo "----------------------------------------------------------------------------------------------"
        exit 0
        ;;
esac

build_clean_vendor
