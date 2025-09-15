#!/bin/bash
# build-linux.sh
GOOS=linux go build -o /tmp/Timekeep/timekeepd ./cmd/service
echo "Binary build to /tmp/Timekeep"

GOOS=linux go build -o /tmp/Timekeep/timekeep ./cmd/cli