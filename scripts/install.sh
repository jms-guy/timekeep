#!/bin/bash

if [ -d "linux" ] && [ -f "linux/timekeepd" ]; then
    BINARY_DIR="linux"
elif [ -d "timekeep-release/linux" ] && [ -f "timekeep-release/linux/timekeepd" ]; then
    BINARY_DIR="timekeep-release/linux"
elif [ -f "./timekeepd" ]; then
    BINARY_DIR="."
else
    echo "Error: Timekeep binaries not found."
    echo "Current directory: $(pwd)"
    echo "Available files:"
    ls -la
    echo "Please run this script from the extracted release directory"
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
User=%U
Group=%G

[Install]
WantedBy=multi-user.target
EOF

sudo setcap cap_dac_read_search,cap_sys_ptrace+ep /usr/local/bin/timekeepd

sudo systemctl daemon-reload
sudo systemctl enable timekeep.service
sudo systemctl start timekeep.service

echo "Installation complete. Run 'timekeep ping' to test."