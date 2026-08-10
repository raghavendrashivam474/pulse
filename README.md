# Pulse

Pulse is a project-awareness CLI for software developers.

## Current Version

v1.0.0 — Sprint 1: Foundation

## What Pulse Is

Pulse is being developed as a developer utility to help quickly understand
the current state of a software project.

## Current Status

Sprint 1 — Foundation.

The CLI is operational. Project analysis will be introduced in Sprint 2.

## Build

```bash
go build -o pulse .
Run
Bash

go run .
pulse
pulse --help
pulse --version
Project Structure
text

pulse/
+-- main.go
+-- go.mod
+-- internal/
    +-- cli/        CLI argument parsing and application entry
    +-- project/    Project identity (Sprint 2)
    +-- scanner/    Project inspection (Sprint 2)
    +-- git/        Git intelligence (Sprint 3)
    +-- snapshot/   Unified project state (Sprint 2+)
    +-- output/     Output rendering (terminal, JSON)
