#! /bin/bash

export BASE_DIR=$( cd "$(dirname "$0")" ; pwd )

DOCKER_IMAGE="build"
DOCKER_VERSION="0.1"
DOCKERFILE="Dockerfile"

function build_container() {
    echo ""
    echo "**********************************"
    echo " Remove previous container "
    echo "**********************************"    
    docker rmi -f $DOCKER_IMAGE:$DOCKER_VERSION

    echo ""
    echo "**********************************"
    echo " Create Build container "
    echo "**********************************"    
    docker build --rm --tag $DOCKER_IMAGE:$DOCKER_VERSION --file $DOCKERFILE .
    docker images
}

function start_container() {
    echo ""
    echo "**********************************"
    echo " Stop Build container "
    echo "**********************************"
    docker rm -f $DOCKER_IMAGE
    docker ps -a

    echo ""
    echo "**********************************"
    echo " Start Build container "
    echo "**********************************"
    docker run -it \
                --privileged \
                --name $DOCKER_IMAGE \
                -v /var/run:/var/run:rw \
                -v /usr/bin/docker:/usr/bin/docker \
                $DOCKER_IMAGE:$DOCKER_VERSION \
                /bin/bash
}


case "$1" in
    "build")
        build_container
        ;;
    "start")
        start_container
        ;;
    "all")
        build_container
        start_container
        ;;
    *)
        echo "build script"
        echo "Usage:"
        echo "-----------------------------------------------------------------------"
        echo "  $0 build        : build container"
        echo "  $0 start        : start container"
        echo "  $0 all          : build & start container"
        echo "-----------------------------------------------------------------------"
        exit 0
        ;;
esac