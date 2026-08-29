package platform

import (
	"github.com/bashkimkasa/procscope/internal/models"
)

// PlatformCollector defines the interface for OS-specific collection
type PlatformCollector interface {
	// GetRunningProcesses returns all running processes
	GetRunningProcesses() map[int]*models.ProcessInfo

	// GetNetworkConnections returns network connections for a specific process
	GetNetworkConnections(pid int) []*models.NetworkConnection

	// GetDNSQueries returns DNS queries issued by a process
	GetDNSQueries(pid int) []*models.DNSQuery

	// GetProcessChildren returns child processes of a given parent
	GetProcessChildren(parentPID int) []*models.ProcessInfo
}

// NewCollector creates a platform-specific collector
func NewCollector() PlatformCollector {
	return newPlatformCollector()
}
