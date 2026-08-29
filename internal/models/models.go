package models

import "time"

// ProcessInfo represents information about a process
type ProcessInfo struct {
	PID        int
	Name       string
	ParentPID  int
	StartTime  time.Time
	Executable string
	Args       []string
}

// NetworkConnection represents a network connection
type NetworkConnection struct {
	Protocol   string // TCP, UDP
	LocalIP    string
	LocalPort  int
	RemoteIP   string
	RemotePort int
	State      string
	PID        int
}

// DNSQuery represents a DNS query
type DNSQuery struct {
	Query     string
	QueryType string // A, AAAA, CNAME, MX, etc.
	Response  string
	Timestamp time.Time
	PID       int
}

// Event represents a normalized event
type Event struct {
	Type      string // "process_start", "process_exit", "network", "dns"
	Timestamp time.Time
	ProcessID int
	Data      interface{} // ProcessInfo, NetworkConnection, DNSQuery, etc.
}

// ProcessGraph represents the process graph
type ProcessGraph struct {
	Processes map[int]*ProcessNode
}

// ProcessNode represents a node in the process graph
type ProcessNode struct {
	ProcessInfo  *ProcessInfo
	Children     []*ProcessNode
	Parent       *ProcessNode
	DNSQueries   []*DNSQuery
	NetworkConns []*NetworkConnection
	ChildrenPIDs []int
}
