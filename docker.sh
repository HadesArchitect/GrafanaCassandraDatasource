#!/bin/bash

REPO="docker-karma"

HOST_MOUNT="$PWD"
CONTAINER_MOUNT="/opt/karma"

HOST_PORT="9876"
CONTAINER_PORT="9876"

function setup {
    if [[ -z "$MOUNT" && -n "$HOST_MOUNT" && -n "$CONTAINER_MOUNT" ]]; then
        MOUNT="-v $HOST_MOUNT:$CONTAINER_MOUNT"
    fi

    if [[ -z "$PORT" && -n "$HOST_PORT" && -n "$CONTAINER_PORT" ]]; then
        PORT="-p $HOST_PORT:$CONTAINER_PORT"
    fi
}

function build {
    setup
    docker build -rm -t $REPO .
}

function run {
    setup
    docker run -rm $PORT $MOUNT -i -t $REPO
}

function cleanup {
    # Remove stopped containers
    CONTAINERS=$(docker ps  -a | grep 'Exit' | awk '{print $1}')
    if [[ -n $CONTAINERS ]]; then
        docker rm $CONTAINERS
    fi

    # Remove untagged images
    IMAGES=$(docker images | grep "^<none>" | awk '{print $3}')
    if [[ -n $IMAGES ]]; then
        docker rmi $IMAGES
    fi
}

function remove {
    IMAGE=$(docker images | grep $REPO | awk '{print $3}')
    if [[ -n $IMAGE ]]; then
        docker rmi $IMAGE
    fi
}

function usage {
    echo "Usage: $0 [build | run | cleanup | remove]"
}

MOUNT=""
PORT=""
while [[ -n "$1" ]]; do
    case $1 in
        -v ) shift
            MOUNT="-v $1"
            ;;
        -p ) shift
            PORT="-p $1"
            ;;
        build ) build
            ;;
        run ) run
            ;;
        cleanup ) cleanup
            ;;
        remove ) remove
            ;;
        * ) usage
            exit 1
            ;;
    esac
    shift
done
