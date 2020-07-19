#! /bin/bash

# Copy service directories

# List of services for edge-orchesrtration 
# srvdirs=(
#     "ls_srv"
# )
# for path in ${srvdirs[@]}
# do 
#     cp -r ${path} /var/edge-orchestration/apps
# done

for path in $(ls -d ./*/) 
do
    cp -r ${path} /var/edge-orchestration/apps
done
