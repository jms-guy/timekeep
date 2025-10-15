![Test Status](https://github.com/jms-guy/timekeep/actions/workflows/CI.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/jms-guy/timekeep)](https://goreportcard.com/report/github.com/jms-guy/timekeep)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)


# Timekeep

A process activity tracker, it runs as a background service recording start/stop events for select programs and aggregates active sessions, session history, and lifetime program usage. Now has [WakaTime](https://github.com/jms-guy/timekeep?tab=readme-ov-file#wakatime) integration.

**Linux version currently not working**

## Table of Contents
- [Features](#features)
- [How It Works](#how-it-works)
- [Usage](#usage)
- [Installation](#installation)
- [WakaTime](#wakatime)
- [File Locations](#file-locations)
- [Current Limitations](#current-limitations)
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
- WakaTime integration allows for tracking external program usage alongside your IDE/web-browsing stats

## How It Works
- Windows: Embeds a PowerShell script to subscribe to WMI process start/stop events. Runs a pre-monitoring script to find any tracked programs already running on service start
- Linux: Polls `/proc`, resolves process identity via `/proc/<pid>/exe` (readlink) -> fallback to `/proc/<pid>/cmdline` -> last-resort `/proc/<pid>/comm`, then matches by basename.
- Session model: A session begins when the first process for a tracked program starts. Additional processes (ex. multiple windows) are added to the active session. The session ends only when the last process terminates, giving an accurate picture of total time with that program.

## Usage

**Full command reference:** [Commands](https://github.com/jms-guy/timekeep/blob/main/docs/commands.md)

### Quick Start
```powershell
timekeep add notepad.exe --category notes # Add notepad
timekeep ls               # List currently tracked programs
 • notepad.exe
timekeep info notepad.exe # Basic info for program sessions
 • Category: notes
 • Current Lifetime: 19h 41m
 • Total sessions to date: 4
 • Last Session: 2025-09-26 11:25 - 2025-09-26 11:26 (21 seconds)
 • Average session length: 4h 55m
timekeep history notepad.exe  # Session history for program
  notepad.exe | 2025-09-26 11:25 - 2025-09-26 11:26 | Duration: 21 seconds
  notepad.exe | 2025-09-24 13:49 - 2025-09-24 13:50 | Duration: 39 seconds
  notepad.exe | 2025-09-23 11:18 - 2025-09-23 11:19 | Duration: 56 seconds
  notepad.exe | 2025-09-22 13:08 - 2025-09-23 08:48 | Duration: 19h 39m
```

**Note**: Program category not required for local tracking. Required for WakaTime integration.

## Installation

### Prerequisites
- **Go 1.24+** (if building from source)
- **Windows**: Administrator privileges for service installation
- **Linux**: sudo privileges for systemd service setup

### Method 1: Install script
1. Download latest release ZIP from [Releases](https://github.com/jms-guy/timekeep/releases)
2. Extract ZIP
3. Run the appropriate install script:
  - **Windows**: Double-click 'install.bat'
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
sc.exe create timekeep binPath= "C:\Program Files\Timekeep\timekeep-service.exe" start= auto # Assuming this is the location of service binary
sc.exe start timekeep

# Verify service is running
Get-Service -Name "timekeep"
```

Test using CLI:
```powershell
.\timekeep.exe status # Check if the service is responsive
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

# Database directory
mkdir -p ~/.local/share/timekeep

# Set service capabilities 
sudo setcap cap_dac_read_search,cap_sys_ptrace+ep /usr/local/bin/timekeepd

# Set user/group variables
USER_NAME=$(whoami)
GROUP_NAME=$(id -gn)

# Create socket directory and set permissions
sudo mkdir -p /var/run/timekeep
sudo chown "$USER_NAME":"$GROUP_NAME" /var/run/timekeep
sudo chmod 755 /var/run/timekeep

# Create and set permissions for log directory
sudo mkdir -p /var/log/timekeep
sudo chown "$USER_NAME":"$GROUP_NAME" /var/log/timekeep
sudo chmod 755 /var/log/timekeep

# Create systemd service
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

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable timekeep.service
sudo systemctl start timekeep.service

# Check status
sudo systemctl status timekeep
```

Test using CLI:
```bash
timekeep status # Check if the service is responsive
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

## WakaTime
Timekeep now integrates with [WakaTime](https://wakatime.com), allowing users to track external program usage alongside their IDE and web-browsing stats. **Timekeep does not track activity within these programs, only when these programs are running.**

To enable WakaTime integration, users must:
  1. Have a WakaTime account
  2. Have [wakatime-cli](https://github.com/wakatime/wakatime-cli) installed on their machine

Enable integration through timekeep. Set your WakaTime API key and wakatime-cli path either directly in the Timekeep [config](https://github.com/jms-guy/timekeep?tab=readme-ov-file#file-locations) file, or provide them through flags:

`timekeep wakatime enable --api-key YOUR-KEY --set-path wakatime-cli-PATH`

```json
{
  "wakatime": {
    "enabled": true,
    "api_key": "APIKEY",
    "cli_path": "PATH",
    "global_project": "PROJECT"
  }
}
```

**The wakatime-cli path must be an absolute path.**

Example path: *C:\Users\Guy\.wakatime\wakatime-cli.exe*

### Complete WakaTime setup example

`timekeep wakatime enable --api-key YOUR-KEY --set-path wakatime-cli-PATH`

`timekeep add photoshop.exe --category designing --project "UI Design"`

Check WakaTime current enabled/disabled status:

`timekeep wakatime status`

Disable integration with:

`timekeep wakatime disable`

### Categories
After enabling, wakatime-cli heartbeats will be sent containing tracking data for given programs. Note, that only programs added to Timekeep with a given category will have data sent to WakaTime.

`timekeep add notepad.exe --category notes`

If no category is set for a program, it will still be tracked locally, but no data for it will be sent out.

List of categories accepted(defined [here](https://github.com/wakatime/wakatime-cli/blob/75ed1c3d905fc77a5039817458298c9ac44853a3/cmd/root.go#L74)):
```bash
"Category of this heartbeat activity. Can be \"coding\", \"ai coding\","+
			" \"building\", \"indexing\", \"debugging\", \"learning\", \"notes\","+
			" \"meeting\", \"planning\", \"researching\", \"communicating\", \"supporting\","+
			" \"advising\", \"running tests\", \"writing tests\", \"manual testing\","+
			" \"writing docs\", \"code reviewing\", \"browsing\","+
			" \"translating\", or \"designing\".
```

### Projects
Timekeep has no automatic project detection for WakaTime. Users may set a global project for all programs to use in the config, or via the command:

`timekeep wakatime set-project "YOUR-PROJECT"`

Users can also set project variables on a per-program basis:

`timekeep add notepad.exe --category notes --project Timekeep`

Program-set project variables will take precedence over a set Global Project. If no project variable is set via the global_project config or when adding programs, WakaTime will fall back to default "Unknown Project".

Users can update a program's category or project with the **update** command:

`timekeep update notepad.exe --category planning --project Timekeep2`

## File Locations
- **Logs** 
  - **Windows**: *C:\ProgramData\Timekeep\logs*
  - **Linux**: */var/log/timekeep*

- **Config**
  - **Windows**: *C:\ProgramData\Timekeep\config*
  - **Linux**: *~/.config/timekeep*
  - **Config struct**:

  ```json
  {
    "wakatime": {
      "enabled": true,
      "api_key": "APIKEY",
      "cli_path": "PATH",
      "global_project": "PROJECT"
    }
  }
  ```

- **Database**
  - **Windows**: *C:\ProgramData\Timekeep*
  - **Linux**: *~/.local/share/timekeep*

## Current Limitations
- Linux - Very short-lived processes can be missed by polling (poll interval currently default 1s)
- Linux - Program basenames may collide (different binaries with same name are treated as same program)

## Contributing & Issues
To contribute, clone the repo with ```git clone https://github.com/jms-guy/timekeep```. Please fork the repository and open a pull request to the `main` branch. Run tests from base repo using ```go test ./...```

If you have an issue, please report it [here](https://github.com/jms-guy/timekeep/issues).

## License
Licensed under MIT - see [LICENSE](https://github.com/jms-guy/timekeep/blob/main/LICENSE).