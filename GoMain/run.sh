#!/bin/bash
/edge-orchestration/edge-orchestration &

edge_pid=$!
# Trap to forward SIGTERM to the child edge container
trap 'kill -TERM $edge_pid' TERM
wait $edge_pid
# Restore SIGTERM to its default behavior
trap - TERM
wait $edge_pid
