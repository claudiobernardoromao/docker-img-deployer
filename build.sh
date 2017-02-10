#!/bin/sh
# Cross-compile build for production.
# Generated Linux-amd64 binary will be saved to ./apprenda/platform-events/bin

echo $(grep "const version =" main.go | sed 's/.*\"\(.*\)\"/\1/') > VERSION
export GOOS=linux
export GOARCH=amd64
glide install -v
go build -o apprenda/platform-events/bin/docker-image
