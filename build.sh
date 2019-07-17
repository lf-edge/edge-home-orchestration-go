#! /bin/bash

export BASE_DIR=$( cd "$(dirname "$0")" ; pwd )

DOCKER_IMAGE="edge-orchestration"
BINARY_FILE="edge-orchestration"

PKG_LIST=(
        "common/errormsg"
        "common/logmgr"
        "common/networkhelper"
        "common/networkhelper/detector"
        "common/resourceutil"
        "common/resourceutil/container"
        "common/types/configuremgrtypes"
        "common/types/servicemgrtypes"
        "controller/configuremgr"
        "controller/configuremgr/container"
        "controller/discoverymgr"
        "controller/discoverymgr/wrapper"
        "controller/scoringmgr"
        "controller/servicemgr"
        "controller/servicemgr/executor"
        "controller/servicemgr/executor/containerexecutor"
        "controller/servicemgr/notification"
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

export VERSION=`grep -w "Version" $BASE_DIR/packaging/$RPM_SPEC_FILE | awk -F ':' '{print $2}' | tr -d ' '`
export CONTAINER_VERSION="alpha"
export BUILD_DATE=$(date +%Y%m%d.%H%M)

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
        echo -ne "(${idx}/${#pkg_list[@]}) go get $pkg"
        go get $pkg
        if [ $? -ne 0 ]; then
            echo -e "\n\033[31m"download fail"\033[0m"
            exit 1
        fi
        echo ": Done"
        idx=$((idx+1))
    done
}

function build_clean() {
    echo ""
    echo "-----------------------------------"
    echo " Build clean"
    echo "-----------------------------------"
    make clean
}

function build_clean_all() {
    echo ""
    echo "-----------------------------------"
    echo " Build clean all"
    echo "-----------------------------------"
    arch_list=("x86" "x86-64" "arm" "aarch64")
    for arch in "${arch_list[@]}"; do
        export ARCH=$arch
        make clean
    done
}

function build_binary() {
    echo ""
    echo "----------------------------------------"
    echo " Create Executable binary from GoMain"
    echo "----------------------------------------"

    export GOPATH=$BASE_DIR/GoMain:$GOPATH
    make build-binary || exit 1
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

function build_clean_vendor() {
    echo ""
    echo "-------------------------------------"
    echo " Clean up 3rdParty directory"
    echo "-------------------------------------"
    make clean-tmp-packages
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
    sudo mkdir -p /var/log/$BINARY_FILE
    sudo mkdir -p /var/data/db

    echo ""
    echo "**********************************"
    echo " Run Docker container "
    echo "**********************************"
    docker run -it -d \
                --privileged \
                --network="host" \
                --name $DOCKER_IMAGE \
                -v /var/run/:/var/run/:rw \
                -v /var/log/:/var/log/:rw \
                -v /etc/:/etc/:rw \
                -v /usr/bin/docker:/usr/bin/docker \
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
        install_prerequisite
        install_3rdparty_packages
        build_binary
        build_docker_container
        run_docker_container
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
        build_clean_all
        ;;
    "")
        install_prerequisite
        install_3rdparty_packages
        build_binary
        build_docker_container
        run_docker_container
        ;;
    *)
        echo "build script"
        echo "Usage:"
        echo "---------------------------------------------------------------------------"
        echo "  $0                  : build edge-orchestration by default container"
        echo "  $0 container        : build Docker container as build system environmet"
        echo ""
        echo "  $0 clean            : build clean"
        echo "  $0 test [PKG_NAME]  : run unittests (optional for PKG_NAME)"
        echo "---------------------------------------------------------------------------"
        exit 0
        ;;
esac

build_clean_vendor
