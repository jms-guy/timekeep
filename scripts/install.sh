#!/bin/bash

if [ -d "linux" ] && [ -f "timekeepd" ]; then
    BINARY_DIR="linux"
elif [ -f "./timekeepd" ]; then
    BINARY_DIR="."
else
    echo "Error: Timekeep binaries not found."
    echo "Please extract the release ZIP and run this script from the extracted directory"
    exit 1
fi

set -e

echo "Installing Timekeep..."

sudo install -m 755 "$BINARY_DIR/timekeepd" /usr/local/bin/
sudo install -m 755 "$BINARY_DIR/timekeep" /usr/local/bin/

sudo tee /etc/systemd/system/timekeep.service > /dev/null <<EOF
[Unit]
Description=TimeKeep Process Tracker
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/timekeepd
Restart=always
User=root

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable timekeep.service
sudo systemctl start timekeep.service

echo "Installation complete. Run 'timekeep ping' to test."