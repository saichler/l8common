#!/usr/bin/env bash

set -e

# Shared types
docker run --user "$(id -u):$(id -g)" -e PROTO=l8common.proto --mount type=bind,source="$PWD",target=/home/proto/ -i saichler/protoc:latest

# Move the generated bindings to the types directory and clean up
rm -rf ../go/types
mkdir -p ../go/types
mv ./types/* ../go/types/.
rm -rf ./types

rm -rf *.rs

cd ../go
find . -name "*.go" -type f -exec sed -i 's|"./types/l8common"|"github.com/saichler/l8common/go/types/l8common"|g' {} +
