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

USER_NAME=$(whoami)
GROUP_NAME=$(id -gn)

sudo mkdir -p /var/run/timekeep
sudo chown "$USER_NAME":"$GROUP_NAME" /var/run/timekeep
sudo chmod 755 /var/run/timekeep

sudo tee /etc/systemd/system/timekeep.service > /dev/null <<EOF
[Unit]
Description=TimeKeep Process Tracker
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/timekeepd
Restart=always
User=$USER_NAME
Group=$GROUP_NAME

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable timekeep.service
sudo systemctl start timekeep.service

echo "Installation complete. Run 'timekeep status' to test."