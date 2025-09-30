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

mkdir -p ~/.local/share/timekeep


sudo setcap cap_dac_read_search,cap_sys_ptrace+ep /usr/local/bin/timekeepd

CURRENT_USER=$(whoami)
sudo tee /etc/systemd/system/timekeep.service > /dev/null <<EOF
[Unit]
Description=TimeKeep Process Tracker
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/timekeepd
Restart=always
User=$CURRENT_USER
Group=$CURRENT_USER

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable timekeep.service
sudo systemctl start timekeep.service

echo "Installation complete. Run 'timekeep ping' to test."