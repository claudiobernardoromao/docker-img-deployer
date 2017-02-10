#!/bin/bash
# The current directory of this process is already the deployer instance directory
WORKLOAD_PATH=$#BASEPATH#$
PLATFORM_EVENTS_PATH=$(pwd)
CID_FILE_PATH="${PLATFORM_EVENTS_PATH}/container.cid"

echo "NoOp handleWorkloadFailure.sh for PLATFORM_EVENTS_PATH: $PLATFORM_EVENTS_PATH and args: $1 - $2 - $3"
