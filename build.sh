#!/bin/bash
# build-windows.sh
GOOS=windows go build -o /mnt/c/temp/timekeep_service.exe ./cmd/service
echo "Binary built to C:\temp\timekeep_service.exe"