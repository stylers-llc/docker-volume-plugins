#!/bin/bash -e
error() {
  printf '\E[31m'; echo "$@"; printf '\E[0m'
}

if [[ $EUID -ne 0 ]]; then
    error "This script should be run using sudo or as the root user"
    exit 1
fi

TAG=$1
build() {
    docker plugin rm -f stylers-llc/glusterfs-volume-plugin || true
    docker rmi -f rootfsimage || true
    docker build -t rootfsimage -f glusterfs-volume-plugin/Dockerfile .
    id=$(docker create rootfsimage true) # id was cd851ce43a403 when the image was created
    rm -rf build/rootfs
    mkdir -p build/rootfs
    docker export "$id" | tar -x -C build/rootfs
    docker rm -vf "$id"
    cp glusterfs-volume-plugin/config.json build
#    if [ -z "$TAG" ]
#    then
        docker plugin create stylers-llc/glusterfs-volume-plugin build
#    else
#        docker plugin create stylers-llc/glusterfs-volume-plugin:$TAG build
#        docker plugin push stylers-llc/glusterfs-volume-plugin:$TAG
#    fi
}

build
