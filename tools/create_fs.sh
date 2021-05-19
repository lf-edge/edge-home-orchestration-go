#! /bin/bash

# List of directories for edge-orchesrtration
fs=(
    "/var/edge-orchestration/log"
    "/var/edge-orchestration/apps"
    "/var/edge-orchestration/data/db"
    "/var/edge-orchestration/data/cert"
    "/var/edge-orchestration/device"
    "/var/edge-orchestration/user"
    "/var/edge-orchestration/mnedc"
    "/var/edge-orchestration/datastorage"
)

# Create file system for edge-orchestration
for path in ${fs[@]} 
do
    mkdir -p $path
done

FILEUSERID="/var/edge-orchestration/user/orchestration_userID.txt"
[ ! -e ${FILEUSERID} ] && echo "Hello world" > ${FILEUSERID}
