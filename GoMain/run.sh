#!/bin/bash
DS_CONFIG=/var/edge-orchestration/datastorage/configuration.toml
if [ -f "$DS_CONFIG" ]; then
    cp -rf /var/edge-orchestration/datastorage/* /edge-orchestration/res/
fi

/edge-orchestration/edge-orchestration &

edge_pid=$!
# Trap to forward SIGTERM to the child edge container
trap 'kill -TERM $edge_pid' TERM
wait $edge_pid
# Restore SIGTERM to its default behavior
trap - TERM
wait $edge_pid
