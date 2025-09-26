![Test Status](https://github.com/jms-guy/timekeep/actions/workflows/CI.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/jms-guy/timekeep)](https://goreportcard.com/report/github.com/jms-guy/timekeep)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)


# Timekeep

A cross-platform process activity tracker written in Go. It runs as a background service, recording start/stop events for selected programs, aggregates active sessions, session history, and lifetime usage.

## Table of Contents
- [Features](#features)
- [How It Works](#how-it-works)
- [Installation](#installation)
- [Usage](#usage)
- [Current Limitations](#current-limitations)
- [To-Do](#to-do)
- [Contributing & Issues](#contributing--issues)
- [License](#license)

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
1. Download latest release from [Releases](https://github.com/jms-guy/timekeep/releases)
2. Extract ZIP file
3. Run the appropriate install script:
  - **Windows**: 'install.ps1' - Right click, run with PowerShell as Administrator
  - **Linux**: ```chmod +x install.sh && sudo ./install.sh```

### Method 2: Build from Source

#### Windows
```powershell
# Clone and build
git clone https://github.com/jms-guy/timekeep
cd timekeep
GOOS=windows go build -o timekeep-service.exe ./cmd/service
GOOS=windows go build -o timekeep.exe ./cmd/cli

# Install and start service (Run as Administrator)
sc.exe create timekeep binPath= "C:\Program Files\Timekeep\timekeep-service.exe" start= auto
sc.exe start timekeep

# Verify service is running
Get-Service -Name "timekeep"
```

#### Linux
```bash
# Clone and build
git clone https://github.com/jms-guy/timekeep
cd timekeep
GOOS=linux go build -o timekeepd ./cmd/service  
GOOS=linux go build -o timekeep ./cmd/cli

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
ExecStart=/usr/local/bin/timekeepd
Restart=always
User=root

[Install]
WantedBy=multi-user.target
EOF

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable timekeep.service
sudo systemctl start timekeep.service

# Check status
sudo systemctl status timekeep
```

### Verify Installation
Test using CLI:
```bash
timekeep ping # Check if the service is responsive
```

## Uninstalling

### Windows
```powershell
sc.exe stop timekeep
sc.exe delete timekeep
```

### Linux
```bash
sudo systemctl disable --now timekeep
sudo rm /etc/systemd/system/timekeep.service
sudo rm /usr/local/bin/timekeepd /usr/local/bin/timekeep
sudo systemctl daemon-reload
```

## Usage

```powershell
timekeep add notepad.exe  # Add notepad
timekeep add code.exe     # Add VS Code
timekeep rm code.exe      # Remove VS Code
timekeep ls               # List currently tracked programs
Programs currently being tracked:
 • notepad.exe
timekeep stats notepad.exe # Basic stats for program sessions
Statistics for notepad.exe:
 • Current Lifetime: 19h 41m
 • Total sessions to date: 4
 • Last Session: 2025-09-26 11:25 - 2025-09-26 11:26 (21 seconds)
 • Average session length: 4h 55m
timekeep history notepad.exe  # Session history for program
Session history for notepad.exe:
  ID: 9 | 2025-09-26 11:25 - 2025-09-26 11:26 | Duration: 21 seconds
  ID: 7 | 2025-09-24 13:49 - 2025-09-24 13:50 | Duration: 39 seconds
  ID: 4 | 2025-09-23 11:18 - 2025-09-23 11:19 | Duration: 56 seconds
  ID: 3 | 2025-09-22 13:08 - 2025-09-23 08:48 | Duration: 19h 39m
```

## Current Limitations
- Linux - Very short-lived processes can be missed by polling (poll interval currently default 1s)
- Linux - Program basenames may collide (different binaries with same name are treated as same program)
- Windows - Processes may be missed if start event happens while service is paused or stopped

## To-Do
- Linux - More accurate start/end time logging
- Linux - Configurable polling interval?
- Windows - Check for running processes on service start
- CLI - commands 
  - show active sessions
  - enhance ping for more service info

## Contributing & Issues
To contribute, clone the repo with ```git clone https://github.com/jms-guy/timekeep```. Please fork the repository and open a pull request to the `main` branch. Tests currently available only for CLI, run tests from base repo using ```go test ./...```

If you have an issue, please report it [here](https://github.com/jms-guy/timekeep/issues).

## License
Licensed under MIT - see LICENSE.