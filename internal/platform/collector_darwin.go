//go:build darwin
// +build darwin

package platform

import (
	"os/exec"
	"strconv"
	"strings"

	"github.com/bashkimkasa/procscope/internal/models"
)

// DarwinCollector implements PlatformCollector for macOS
type DarwinCollector struct{}

func newPlatformCollector() PlatformCollector {
	return &DarwinCollector{}
}

// GetRunningProcesses gets all running processes on macOS
func (dc *DarwinCollector) GetRunningProcesses() map[int]*models.ProcessInfo {
	processes := make(map[int]*models.ProcessInfo)

	// Use ps command on macOS to get process information
	cmd := exec.Command("ps", "-ax", "-o", "pid,ppid,comm")
	output, err := cmd.Output()
	if err != nil {
		return processes
	}

	lines := strings.Split(string(output), "\n")
	for i, line := range lines {
		if i == 0 { // Skip header
			continue
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		pidStr := fields[0]
		ppidStr := fields[1]
		name := fields[2]

		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			continue
		}

		ppid, err := strconv.Atoi(ppidStr)
		if err != nil {
			ppid = 0
		}

		pinfo := &models.ProcessInfo{
			PID:        pid,
			Name:       name,
			ParentPID:  ppid,
			Executable: "",
		}

		processes[pid] = pinfo
	}

	return processes
}

// GetDNSQueries gets DNS queries for a process on macOS by parsing the process command line.
func (dc *DarwinCollector) GetDNSQueries(pid int) []*models.DNSQuery {
	cmd := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "command=")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}
	return extractDNSQueriesFromCommandLine(string(output), pid)
}

// GetNetworkConnections gets network connections for a process on macOS
func (dc *DarwinCollector) GetNetworkConnections(pid int) []*models.NetworkConnection {
	conns := make([]*models.NetworkConnection, 0)

	// Use lsof command to get open network files
	cmd := exec.Command("lsof", "-p", strconv.Itoa(pid), "-i", "-P", "-n")
	output, err := cmd.Output()
	if err != nil {
		return conns
	}

	lines := strings.Split(string(output), "\n")
	for i, line := range lines {
		if i == 0 { // Skip header
			continue
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Simple parsing of lsof output
		// Format: COMMAND PID USER FD TYPE DEVICE SIZE/OFF NODE NAME
		fields := strings.Fields(line)
		if len(fields) < 9 {
			continue
		}

		nodeInfo := fields[len(fields)-1] // Last field is the name/address

		// Skip if not an IP connection
		if !strings.Contains(nodeInfo, "->") {
			continue
		}

		// Very basic parsing - just capture the connection info
		conn := &models.NetworkConnection{
			Protocol:   "TCP", // Simplified - would need better parsing
			LocalIP:    "",
			LocalPort:  0,
			RemoteIP:   nodeInfo,
			RemotePort: 0,
			State:      "ESTABLISHED",
			PID:        pid,
		}

		conns = append(conns, conn)
	}

	return conns
}

// GetProcessChildren returns child processes on macOS
func (dc *DarwinCollector) GetProcessChildren(parentPID int) []*models.ProcessInfo {
	children := make([]*models.ProcessInfo, 0)
	allProcesses := dc.GetRunningProcesses()

	for _, pinfo := range allProcesses {
		if pinfo.ParentPID == parentPID {
			children = append(children, pinfo)
		}
	}

	return children
}
