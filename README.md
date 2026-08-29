# procscope

**Lightweight, cross-platform process behavior observability.**

procscope answers a simple question:

> **What is this process actually doing?**

It collects process metadata, ancestry, network state, and operating-system signals, then correlates them into a single real-time view of behavior. The goal is simple: stop treating process activity, hostnames, and network connections as disconnected events and instead show how they relate to one another.

The current implementation includes a working cross-platform process graph and live monitoring flow, with hostname-context correlation derived from command lines and network metadata rather than full kernel-level DNS event capture.

Instead of looking at processes, DNS, and network connections as separate streams, procscope connects them:

```text
VS Code
  └── node
       └── npm
            └── python
                 ├── DNS → api.example.com
                 ├── TCP → 10.20.4.15:443
                 └── spawned → curl
```

## Why procscope?

Modern operating systems provide plenty of information about what is happening on a machine, but that information is often fragmented.

You can see:

* which processes are running
* which connections exist
* which DNS queries were made
* which processes spawned other processes

But answering **why these events are related** is much harder.

procscope correlates those events into a process-centric view.

```text
Process
   │
   ├── Parent / Children
   │
   ├── DNS Queries
   │
   ├── Network Connections
   │
   ├── File Activity
   │
   └── Other Host Events
```

## Example

Watch activity in real time:

```bash
procscope watch
```

```text
PROCESS       PID    PARENT       DNS                    NETWORK

python        3812   vscode       api.foo.com            10.20.4.15:443
                                  registry.npmjs.org

node          4211   vscode       github.com             github.com:443

curl          4432   python       api.foo.com             10.20.4.15:443
```

Inspect a specific process:

```bash
procscope explain 3812
```

```text
Process:       python
PID:           3812
Parent:        vscode

Children:
  curl

DNS:
  api.foo.com
  registry.npmjs.org

Network:
  10.20.4.15:443
  github.com:443

First observed: 21:42:13
```

The goal is to make relationships visible:

```text
                    ┌──────────────┐
                    │    python    │
                    │    PID 3812  │
                    └──────┬───────┘
                           │
             ┌─────────────┼─────────────┐
             │             │             │
             ▼             ▼             ▼
          spawned        DNS          network
             │             │             │
             ▼             ▼             ▼
           curl       api.foo.com    10.20.4.15:443
```

## Current status

The project currently includes a working cross-platform process collection and correlation layer with heuristic hostname-context enrichment.

### Implemented so far

* Process enumeration and snapshots on Windows, Linux, and macOS
* Process ancestry tracking via parent/child relationships
* Process lifecycle event handling for start and exit behavior
* Network connection collection using OS-native mechanisms
* Hostname-context extraction from command lines and connection metadata
* Real-time process graph generation in memory
* Correlation layer that groups process activity by PID and parent relationship
* `watch` command for live process monitoring
* `explain` command for a detailed snapshot of a single process
* Build scripts and output organization under `bin/`

### Platform implementation

| Platform | Current method | Status |
| -------- | -------------- | ------ |
| Windows | PowerShell `Get-Process` + `netstat` | ✅ Implemented |
| Linux | `/proc` filesystem parsing | ✅ Implemented |
| macOS | `ps` + `lsof` | ✅ Implemented |

The platform-specific collectors normalize the data into a common internal model consumed by the same correlation and graph engine. The current hostname-context signal is intentionally heuristic and designed to enrich process behavior without claiming full DNS capture.

```text
              ┌─────────────────────┐
              │    procscope core   │
              │                     │
              │ Event Model         │
              │ Correlator          │
              │ Process Graph       │
              └──────────┬──────────┘
                         │
          ┌──────────────┼──────────────┐
          │              │              │
        Linux          Windows        macOS
          │              │              │
        /proc        PowerShell + netstat   ps + lsof
```

## Design principles

* **Lightweight** — designed to run continuously with minimal overhead
* **Local-first** — telemetry remains on the host unless explicitly exported
* **Cross-platform** — common event model with OS-specific collectors
* **Real-time** — observe behavior as it happens
* **Correlation over collection** — connect related events instead of producing disconnected logs
* **Process-centric** — the process is the primary unit of observation
* **eBPF where it makes sense** — eBPF is an implementation mechanism, not the product itself

## What procscope is not

procscope is intentionally not:

* an EDR
* a SIEM
* a packet capture tool
* a vulnerability scanner
* a container security platform
* a network firewall

It is a **lightweight process behavior observability tool**.

## Roadmap

### Phase 1 — Complete

* [x] Process lifecycle events
* [x] Process ancestry tracking
* [x] Network connection collection
* [x] Real-time process graph
* [x] `watch` command
* [x] `explain` command
* [x] Cross-platform collection on Windows, Linux, and macOS

### Phase 2 — In progress

* [x] Process → hostname-context correlation from command lines and network metadata
* [ ] Deeper OS-level DNS query capture
* [ ] Process → network correlation details
* [ ] Parent → child correlation visualization enhancements
* [ ] File activity tracking
* [ ] Historical process relationships

### Phase 3 — Observability enhancements

* [ ] Interactive terminal UI
* [ ] JSON output mode
* [ ] OpenTelemetry export
* [ ] Event filtering and search
* [ ] Process behavior timelines

### Future

* [ ] Behavioral anomaly detection
* [ ] Network intent analysis
* [ ] Policy visualization
* [ ] Optional enforcement

## Architecture

The project is designed to keep collection separate from correlation.

```text
┌────────────────────────────────────────────────────┐
│                    procscope                       │
│                                                    │
│  ┌──────────────┐      ┌────────────────────────┐  │
│  │ OS Collectors│ ───▶ │    Event Correlator    │  │
│  └──────────────┘      └────────────┬───────────┘  │
│                                     │              │
│                                     ▼              │
│                           ┌─────────────────────┐  │
│                           │    Process Graph    │  │
│                           └──────────┬──────────┘  │
│                                      │             │
│                         ┌────────────┼───────────┐ │
│                         ▼            ▼           ▼ │
│                       CLI          JSON        OTEL│
└────────────────────────────────────────────────────┘
```

The initial implementation is intended to use **Go for the agent and correlation engine**, with eBPF programs providing Linux kernel telemetry.

## Building

### Prerequisites

- Go 1.22 or later
- Windows 10+, Linux (any distribution), or macOS 10.14+
- (Optional) GNU Make for using the Makefile

### Build for Your Platform

**Windows (using PowerShell):**
```bash
.\build.ps1
# or
go build -o bin/procscope.exe .
```

**Linux:**
```bash
make build
# or
go build -o bin/procscope-linux .
```

**macOS:**
```bash
make build
# or
go build -o bin/procscope-mac .
```

### Build for All Platforms

From any platform:

**Using Make:**
```bash
make build-all
```

**Using PowerShell (Windows):**
```bash
.\build.ps1 build-all
```

**Using Go directly:**
```bash
# Windows
GOOS=windows GOARCH=amd64 go build -o bin/procscope.exe .

# Linux
GOOS=linux GOARCH=amd64 go build -o bin/procscope-linux .

# macOS
GOOS=darwin GOARCH=amd64 go build -o bin/procscope-mac .
```

### Using Make

A Makefile is provided for convenience (works on Linux/macOS):

```bash
make build       # Build for current platform
make build-all   # Build for all platforms
make clean       # Remove build artifacts
make test        # Run tests
make deps        # Download dependencies
make help        # Show build targets
```

### Using PowerShell (Windows)

A build script is provided for Windows users:

```powershell
.\build.ps1              # Build for Windows
.\build.ps1 build-all    # Build for all platforms
.\build.ps1 clean        # Remove build artifacts
.\build.ps1 help         # Show help
```

## Usage

### Command: `watch`

Real-time monitoring of process activity:

**Windows:**
```bash
./bin/procscope.exe watch
```

**Linux/macOS:**
```bash
./bin/procscope watch
```

Displays a live table of processes with their associated DNS queries and network connections.

**Example output:**
```
PROCESS                  PID      PARENT     DNS                      NETWORK
powershell               1880     0
chrome                   4156     1880       google.com               142.251.32.14:443
code                     3924     1880       marketplace.visualstud... 13.107.42.14:443
```

### Command: `explain`

Detailed analysis of a specific process by PID:

**Windows:**
```bash
./bin/procscope.exe explain <PID>
```

**Linux/macOS:**
```bash
./bin/procscope explain <PID>
```

Shows comprehensive information about a process including:
- Process name and ID
- Parent process information
- Child processes
- DNS queries
- Network connections

**Example output:**
```
procscope explain - Detailed process behavior for PID 1880

Process:       powershell
PID:           1880
Parent:        0

Children:
  chrome (PID 4156)
  code (PID 3924)

DNS Queries:
  (none)

Network Connections:
  TCP://127.0.0.1:5000 -> 192.168.1.100:443 (ESTABLISHED)
```

### Command: `help`

Display usage information:

**Windows:**
```bash
./bin/procscope.exe help
```

**Linux/macOS:**
```bash
./bin/procscope help
```

## Platform Support

### Current Implementation (Phase 1)

| Platform | Status | Method |
|----------|--------|--------|
| Windows | ✅ Full | PowerShell / netstat |
| Linux | ✅ Full | `/proc` filesystem |
| macOS | ✅ Full | `ps` / `lsof` |

#### Windows Implementation
- Process enumeration: PowerShell `Get-Process` cmdlet
- Process hierarchy: Parent process relationship via PowerShell
- Network connections: `netstat -ano` command
- Real-time monitoring: Polling every 2 seconds

#### Linux Implementation
- Process enumeration: Reading `/proc/[pid]/stat` and `/proc/[pid]/status`
- Process hierarchy: Parent PID from `/proc/[pid]/status`
- Network connections: Parsing `/proc/net/tcp` and `/proc/net/tcp6`
- Executable paths: Reading `/proc/[pid]/exe` symlink
- Real-time monitoring: Polling every 2 seconds

#### macOS Implementation
- Process enumeration: `ps -ax` command
- Process hierarchy: Using `ppid` from `ps` output
- Network connections: `lsof` command for open files
- Real-time monitoring: Polling every 2 seconds

### Testing Across Platforms

**Windows:** Run directly
```bash
./bin/procscope.exe explain 1880
```

**Linux:** Run directly or via cross-compilation
```bash
./bin/procscope-linux explain 1
```

**macOS:** Run directly or cross-compile
```bash
./bin/procscope-mac explain 1
```

**Testing Linux on Windows:** Use WSL2
```bash
wsl ./bin/procscope-linux explain 1
```

## Project Structure

```
procscope/
├── main.go                      # CLI entry point
├── go.mod                       # Go module file
├── Makefile                     # Build automation (Linux/macOS)
├── build.ps1                    # Build automation (Windows)
├── README.md                    # This file
├── .gitignore                   # Git ignore patterns
├── bin/                         # Build output directory
│   ├── procscope.exe           # Windows executable
│   ├── procscope-linux         # Linux executable
│   └── procscope-mac           # macOS executable
└── internal/
    ├── models/
    │   └── models.go           # Data structures
    ├── collector/
    │   └── collector.go        # Platform-agnostic collector wrapper
    ├── platform/
    │   ├── interface.go        # PlatformCollector interface
    │   ├── collector_windows.go# Windows-specific implementation
    │   ├── collector_linux.go  # Linux-specific implementation
    │   └── collector_darwin.go # macOS-specific implementation
    ├── correlator/
    │   └── correlator.go       # Event correlation engine
    └── graph/
        └── graph.go            # Process graph and visualization
```

## Implementation Status

### Phase 1 ✅ (Complete)
- [x] Process lifecycle events
- [x] Process ancestry tracking
- [x] Network connections collection
- [x] DNS correlation (framework ready)
- [x] Real-time process graph
- [x] `watch` command
- [x] `explain` command
- [x] Cross-platform support (Windows, Linux, macOS)

### Phase 2 (Planned)
- [ ] Process → DNS correlation
- [ ] Process → network correlation
- [ ] Parent → child correlation visualization
- [ ] File activity tracking
- [ ] Historical process relationships

### Phase 3 (Planned)
- [ ] Windows ETW/eBPF integration
- [ ] Linux eBPF-based collection
- [ ] macOS Endpoint Security integration

### Phase 4 (Planned)
- [ ] Interactive terminal UI
- [ ] JSON output format
- [ ] OpenTelemetry export
- [ ] Advanced event filtering

## Development

### Building the Project

```bash
# Clone the repository
git clone https://github.com/bashkimkasa/procscope.git
cd procscope

# Build for your platform (outputs to bin/ directory)
make build          # Linux/macOS
.\build.ps1         # Windows

# Run
./bin/procscope explain 1      # Linux/macOS
./bin/procscope.exe explain 1  # Windows
```

### Building for Multiple Platforms

```bash
# Using Make (Linux/macOS)
make build-all

# Using PowerShell (Windows)
.\build.ps1 build-all

# Using Go directly
GOOS=windows GOARCH=amd64 go build -o bin/procscope.exe .
GOOS=linux GOARCH=amd64 go build -o bin/procscope-linux .
GOOS=darwin GOARCH=amd64 go build -o bin/procscope-mac .
```

### Platform-Specific Code

The project uses Go build tags for platform-specific implementations:

```go
//go:build windows
// +build windows

package platform
// Windows-only code
```

This allows the same source to be built for multiple platforms with platform-specific collectors.

### Adding a New Platform

1. Create `internal/platform/collector_<os>.go`
2. Implement the `PlatformCollector` interface
3. Use build tags: `//go:build <os>`
4. The `Collector` wrapper will automatically use the correct implementation

## Why?

A process opened a connection.

A DNS query happened.

A child process was spawned.

A file was accessed.

These events are easy to collect individually.

The interesting part is understanding how they relate.

procscope exists to make those relationships visible.

> **Don't just see what is running. See what it is doing.**
