//go:build windows
// +build windows

package platform

import (
	"os/exec"
	"strconv"
	"strings"

	"github.com/bashkimkasa/procscope/internal/models"
)

// WindowsCollector implements PlatformCollector for Windows
type WindowsCollector struct{}

func newPlatformCollector() PlatformCollector {
	return &WindowsCollector{}
}

// GetRunningProcesses gets all running processes on Windows
func (wc *WindowsCollector) GetRunningProcesses() map[int]*models.ProcessInfo {
	processes := make(map[int]*models.ProcessInfo)

	// Use Win32_Process so ParentProcessId is returned directly instead of relying on
	// Get-Process Parent metadata, which is often empty or unavailable.
	cmd := exec.Command("powershell", "-Command",
		"Get-CimInstance Win32_Process | Select-Object ProcessId, Name, ParentProcessId | ConvertTo-Csv -NoTypeInformation")
	output, err := cmd.Output()
	if err != nil {
		return processes
	}

	lines := strings.Split(string(output), "\n")
	for i, line := range lines {
		if i < 1 { // Skip header
			continue
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse CSV line
		fields := parseCSV(line)
		if len(fields) < 2 {
			continue
		}

		pidStr := fields[0]
		name := fields[1]
		parentIDStr := ""
		if len(fields) > 2 {
			parentIDStr = fields[2]
		}

		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			continue
		}

		ppid := 0
		if parentIDStr != "" && parentIDStr != "0" {
			ppid, _ = strconv.Atoi(parentIDStr)
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

// GetNetworkConnections gets network connections for a process on Windows
func (wc *WindowsCollector) GetNetworkConnections(pid int) []*models.NetworkConnection {
	conns := make([]*models.NetworkConnection, 0)

	// Use netstat on Windows
	cmd := exec.Command("netstat", "-ano")
	output, err := cmd.Output()
	if err != nil {
		return conns
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		pidStr := fields[len(fields)-1]
		connPID, err := strconv.Atoi(pidStr)
		if err != nil || connPID != pid {
			continue
		}

		// Parse protocol and addresses
		protocol := fields[0]
		localAddr := fields[1]
		remoteAddr := fields[2]
		state := fields[3]

		localIP, localPort := parseAddress(localAddr)
		remoteIP, remotePort := parseAddress(remoteAddr)

		conn := &models.NetworkConnection{
			Protocol:   protocol,
			LocalIP:    localIP,
			LocalPort:  localPort,
			RemoteIP:   remoteIP,
			RemotePort: remotePort,
			State:      state,
			PID:        pid,
		}

		conns = append(conns, conn)
	}

	return conns
}

// GetProcessChildren returns child processes
func (wc *WindowsCollector) GetProcessChildren(parentPID int) []*models.ProcessInfo {
	children := make([]*models.ProcessInfo, 0)
	allProcesses := wc.GetRunningProcesses()

	for _, pinfo := range allProcesses {
		if pinfo.ParentPID == parentPID {
			children = append(children, pinfo)
		}
	}

	return children
}

// parseAddress parses an address string like "127.0.0.1:8080"
func parseAddress(addr string) (string, int) {
	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		return addr, 0
	}

	ip := parts[0]
	port, _ := strconv.Atoi(parts[1])
	return ip, port
}

// parseCSV parses a simple CSV line with quoted fields
func parseCSV(line string) []string {
	var fields []string
	var current strings.Builder
	inQuotes := false

	for _, ch := range line {
		if ch == '"' {
			inQuotes = !inQuotes
		} else if ch == ',' && !inQuotes {
			fields = append(fields, strings.TrimSpace(current.String()))
			current.Reset()
		} else {
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 {
		fields = append(fields, strings.TrimSpace(current.String()))
	}

	return fields
}
