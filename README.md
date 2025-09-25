![Test Status](https://github.com/jms-guy/timekeep/actions/workflows/CI.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/jms-guy/greed)](https://goreportcard.com/report/github.com/jms-guy/timekeep)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)


# Timekeep

A cross-platform process activity tracker written in Go. It records start/stop events for selected programs, aggregates active sessions, session history, and lifetime usage. Runs as a Windows service, Linux functionality is currently being built.

## Table of Contents
- [Features](#features)
- [How It Works](#how-it-works)

## Features
- Track programs by executable basename (e.g., `notepad.exe`, `code`, `bash`)
- Start/stop detection:
  - Windows: WMI PowerShell subscription
  - Linux: /proc polling with exe/cmdline-based identity
- Active session aggregation across multiple PIDs
- Session history and total lifetime durations
- CLI for managing tracked programs

## How It Works
- Windows: embeds a PowerShell script to subscribe to WMI process start/stop events.
- Linux: polls `/proc`, resolves process identity via `/proc/<pid>/exe` (readlink) -> fallback to `/proc/<pid>/cmdline` -> last-resort `/proc/<pid>/comm`, then matches by basename.
- Session model: first PID for a program starts a session; additional PIDs join it; last PID exit ends the session.

## Installation

### Prerequisites
- **Go 1.24+** (if building from source)
- **Windows**: Administrator privileges for service installation
- **Linux**: sudo privileges for systemd service setup

### Method 1: Install script
Download pre-built binaries from the [Releases](https://github.com/jms-guy/timekeep/releases) page, and run install.ps1(Windows) or install.sh(Linux) inside extracted ZIP.

### Method 2: Build from Source

#### Windows
```powershell
# Clone and build
git clone https://github.com/jms-guy/timekeep
cd timekeep
go build -o timekeep-service.exe ./cmd/service
go build -o timekeep.exe ./cmd/cli

# Install and start service (Run as Administrator)
.\timekeep-service.exe install
.\timekeep-service.exe start

# Verify service is running
Get-Service -Name "timekeep"
```

#### Linux
```bash
# Clone and build
git clone https://github.com/jms-guy/timekeep
cd timekeep
go build -o timekeepd ./cmd/service  
go build -o timekeep ./cmd/cli

# Install binaries
sudo install -m 755 timekeepd /usr/local/bin/
sudo install -m 755 timekeep /usr/local/bin/

# Create systemd service
sudo tee /etc/systemd/system/timekeep.service > /dev/null <<EOF
[Unit]
Description=TimeKeep Process Tracker
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/timekeep-service
Restart=always
User=root

[Install]
WantedBy=multi-user.target
EOF

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable timekeep
sudo systemctl start timekeep

# Check status
sudo systemctl status timekeep
```

#### Verify Installation
Test using CLI:
```bash
timekeep ping # Check if the service is responsive
```