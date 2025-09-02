#!/bin/bash
# build-windows.sh
GOOS=windows GOARCH=amd64 go build -o /mnt/c/temp/processtrack.exe .
echo "Binary built to C:\temp\processtrack.exe"