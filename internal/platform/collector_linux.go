//go:build linux
// +build linux

package platform

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bashkimkasa/procscope/internal/models"
)

// LinuxCollector implements PlatformCollector for Linux using /proc filesystem
type LinuxCollector struct{}

func newPlatformCollector() PlatformCollector {
	return &LinuxCollector{}
}

// GetRunningProcesses gets all running processes on Linux
func (lc *LinuxCollector) GetRunningProcesses() map[int]*models.ProcessInfo {
	processes := make(map[int]*models.ProcessInfo)

	procDir := "/proc"
	entries, err := os.ReadDir(procDir)
	if err != nil {
		return processes
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Check if directory name is a valid PID (all digits)
		pidStr := entry.Name()
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			continue
		}

		pinfo := lc.getProcessInfo(pid)
		if pinfo != nil {
			processes[pid] = pinfo
		}
	}

	return processes
}

// getProcessInfo reads process info from /proc/[pid]/stat and /proc/[pid]/status
func (lc *LinuxCollector) getProcessInfo(pid int) *models.ProcessInfo {
	pidStr := strconv.Itoa(pid)

	// Read /proc/[pid]/stat for basic info
	statPath := filepath.Join("/proc", pidStr, "stat")
	statContent, err := os.ReadFile(statPath)
	if err != nil {
		return nil
	}

	// Read /proc/[pid]/status for parent PID and comm
	statusPath := filepath.Join("/proc", pidStr, "status")
	statusContent, err := os.ReadFile(statusPath)
	if err != nil {
		return nil
	}

	// Parse status file for Name and PPid
	name := ""
	ppid := 0

	statusLines := strings.Split(string(statusContent), "\n")
	for _, line := range statusLines {
		if strings.HasPrefix(line, "Name:") {
			name = strings.TrimSpace(strings.TrimPrefix(line, "Name:"))
		} else if strings.HasPrefix(line, "PPid:") {
			ppidStr := strings.TrimSpace(strings.TrimPrefix(line, "PPid:"))
			ppid, _ = strconv.Atoi(ppidStr)
		}
	}

	if name == "" {
		// Fallback: parse from stat file (second field in parentheses)
		statStr := string(statContent)
		if idx := strings.LastIndex(statStr, ")"); idx > 0 {
			start := strings.Index(statStr, "(") + 1
			name = statStr[start:idx]
		}
	}

	if name == "" {
		name = fmt.Sprintf("pid-%d", pid)
	}

	return &models.ProcessInfo{
		PID:        pid,
		Name:       name,
		ParentPID:  ppid,
		Executable: lc.getExecutablePath(pid),
	}
}

// getExecutablePath gets the executable path from /proc/[pid]/exe
func (lc *LinuxCollector) getExecutablePath(pid int) string {
	pidStr := strconv.Itoa(pid)
	exePath := filepath.Join("/proc", pidStr, "exe")

	target, err := os.Readlink(exePath)
	if err != nil {
		return ""
	}

	return target
}

// GetDNSQueries gets DNS queries for a process on Linux by parsing the process command line.
func (lc *LinuxCollector) GetDNSQueries(pid int) []*models.DNSQuery {
	cmdlinePath := filepath.Join("/proc", strconv.Itoa(pid), "cmdline")
	content, err := os.ReadFile(cmdlinePath)
	if err != nil {
		return nil
	}
	return extractDNSQueriesFromCommandLine(string(content), pid)
}

// GetNetworkConnections gets network connections for a process on Linux
func (lc *LinuxCollector) GetNetworkConnections(pid int) []*models.NetworkConnection {
	conns := make([]*models.NetworkConnection, 0)

	pidStr := strconv.Itoa(pid)

	// Check /proc/[pid]/fd for network sockets
	fdPath := filepath.Join("/proc", pidStr, "fd")
	entries, err := os.ReadDir(fdPath)
	if err != nil {
		return conns
	}

	for _, entry := range entries {
		// Read the symlink target to get socket inode
		symPath := filepath.Join(fdPath, entry.Name())
		target, err := os.Readlink(symPath)
		if err != nil || !strings.Contains(target, "socket:[") {
			continue
		}

		// Extract inode number
		inode := extractInode(target)
		if inode == "" {
			continue
		}

		// Parse /proc/net/tcp and /proc/net/tcp6 to find connections
		conn := lc.parseNetConnection(inode, "tcp", pid)
		if conn != nil {
			conns = append(conns, conn)
		} else {
			conn = lc.parseNetConnection(inode, "tcp6", pid)
			if conn != nil {
				conns = append(conns, conn)
			}
		}
	}

	return conns
}

// parseNetConnection parses a connection from /proc/net/tcp or /proc/net/tcp6
func (lc *LinuxCollector) parseNetConnection(inode string, proto string, pid int) *models.NetworkConnection {
	netPath := filepath.Join("/proc/net", proto)
	content, err := os.ReadFile(netPath)
	if err != nil {
		return nil
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if !strings.Contains(line, inode) {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		// Parse local address (hex format)
		localAddr := parseHexAddress(fields[1])
		remoteAddr := parseHexAddress(fields[2])
		state := fields[3]

		if localAddr.IP == "" {
			continue
		}

		return &models.NetworkConnection{
			Protocol:   strings.ToUpper(proto),
			LocalIP:    localAddr.IP,
			LocalPort:  localAddr.Port,
			RemoteIP:   remoteAddr.IP,
			RemotePort: remoteAddr.Port,
			State:      state,
			PID:        pid,
		}
	}

	return nil
}

// HexAddr represents a parsed hex address
type HexAddr struct {
	IP   string
	Port int
}

// parseHexAddress parses a hex encoded address:port
func parseHexAddress(hexAddr string) HexAddr {
	parts := strings.Split(hexAddr, ":")
	if len(parts) != 2 {
		return HexAddr{"", 0}
	}

	ip := hexToIP(parts[0])
	port, _ := strconv.ParseInt(parts[1], 16, 32)

	return HexAddr{ip, int(port)}
}

// hexToIP converts hex string to IP address
func hexToIP(hex string) string {
	if len(hex) != 8 {
		return ""
	}

	// Convert from hex to bytes (little-endian)
	b := make([]byte, 4)
	for i := 0; i < 4; i++ {
		val, _ := strconv.ParseInt(hex[i*2:i*2+2], 16, 8)
		b[3-i] = byte(val)
	}

	return fmt.Sprintf("%d.%d.%d.%d", b[0], b[1], b[2], b[3])
}

// extractInode extracts the inode number from socket:[INODE]
func extractInode(sockStr string) string {
	start := strings.Index(sockStr, "[")
	end := strings.Index(sockStr, "]")
	if start == -1 || end == -1 {
		return ""
	}
	return sockStr[start+1 : end]
}

// GetProcessChildren returns child processes on Linux
func (lc *LinuxCollector) GetProcessChildren(parentPID int) []*models.ProcessInfo {
	children := make([]*models.ProcessInfo, 0)
	allProcesses := lc.GetRunningProcesses()

	for _, pinfo := range allProcesses {
		if pinfo.ParentPID == parentPID {
			children = append(children, pinfo)
		}
	}

	return children
}
