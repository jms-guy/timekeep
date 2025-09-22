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
