package graph

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bashkimkasa/procscope/internal/models"
)

// Graph provides visualization and analysis of the process graph
type Graph struct {
	name         string
	processGraph *models.ProcessGraph
}

// NewGraph creates a new process graph
func NewGraph() *Graph {
	return &Graph{
		name:         "ProcessGraph",
		processGraph: &models.ProcessGraph{Processes: make(map[int]*models.ProcessNode)},
	}
}

// Name returns the graph name
func (g *Graph) Name() string {
	return g.name
}

// SetProcessGraph sets the underlying process graph
func (g *Graph) SetProcessGraph(pg *models.ProcessGraph) {
	g.processGraph = pg
}

// GetProcessGraph returns the process graph
func (g *Graph) GetProcessGraph() *models.ProcessGraph {
	return g.processGraph
}

// FormatTable formats the process graph as a table for display
func (g *Graph) FormatTable() string {
	var output string
	output += fmt.Sprintf("%-20s %-8s %-10s %-25s %-20s\n", "PROCESS", "PID", "PARENT", "HOSTNAME", "NETWORK")
	output += fmt.Sprintf("%-20s %-8s %-10s %-25s %-20s\n", strings.Repeat("-", 19), strings.Repeat("-", 7), strings.Repeat("-", 9), strings.Repeat("-", 24), strings.Repeat("-", 19))

	// Sort by PID for consistent output
	var pids []int
	for pid := range g.processGraph.Processes {
		pids = append(pids, pid)
	}
	sort.Ints(pids)

	for _, pid := range pids {
		node := g.processGraph.Processes[pid]
		if node.ProcessInfo == nil {
			continue
		}

		// Collect DNS domains (first 2)
		dnsStr := ""
		for i, query := range node.DNSQueries {
			if i >= 2 {
				dnsStr += "..."
				break
			}
			if i > 0 {
				dnsStr += ", "
			}
			dnsStr += query.Query
		}

		// Collect network connections (first 1)
		networkStr := ""
		for _, conn := range node.NetworkConns {
			if networkStr != "" {
				networkStr += "..."
				break
			}
			networkStr = fmt.Sprintf("%s:%d", conn.RemoteIP, conn.RemotePort)
		}

		parentLabel := fmt.Sprintf("%d", node.ProcessInfo.ParentPID)
		if node.Parent != nil && node.Parent.ProcessInfo != nil {
			parentLabel = fmt.Sprintf("%s (%d)", node.Parent.ProcessInfo.Name, node.Parent.ProcessInfo.PID)
		}

		output += fmt.Sprintf("%-20s %-8d %-18s %-25s %-20s\n",
			node.ProcessInfo.Name,
			node.ProcessInfo.PID,
			parentLabel,
			dnsStr,
			networkStr,
		)
	}

	return output
}

// FormatProcessDetail formats detailed information about a specific process
func (g *Graph) FormatProcessDetail(pid int) string {
	node, exists := g.processGraph.Processes[pid]
	if !exists || node.ProcessInfo == nil {
		return fmt.Sprintf("Process with PID %d not found\n", pid)
	}

	pinfo := node.ProcessInfo
	output := fmt.Sprintf("Process:       %s\n", pinfo.Name)
	output += fmt.Sprintf("PID:           %d\n", pinfo.PID)

	parentLabel := fmt.Sprintf("%d", pinfo.ParentPID)
	if node.Parent != nil && node.Parent.ProcessInfo != nil {
		parentLabel = fmt.Sprintf("%s (PID %d)", node.Parent.ProcessInfo.Name, node.Parent.ProcessInfo.PID)
	}
	output += fmt.Sprintf("Parent:        %s\n", parentLabel)

	if pinfo.Executable != "" {
		output += fmt.Sprintf("Executable:    %s\n", pinfo.Executable)
	}

	output += "\nSummary:\n"
	childCount := len(node.Children)
	output += fmt.Sprintf("  Parent: %s\n", parentLabel)
	output += fmt.Sprintf("  Children: %d\n", childCount)
	output += fmt.Sprintf("  Hostnames: %d\n", len(node.DNSQueries))
	output += fmt.Sprintf("  Network connections: %d\n", len(node.NetworkConns))

	if len(node.Children) > 0 {
		output += "\nChildren:\n"
		for _, child := range node.Children {
			if child.ProcessInfo != nil {
				output += fmt.Sprintf("  %s (PID %d)\n", child.ProcessInfo.Name, child.ProcessInfo.PID)
			}
		}
	} else {
		output += "\nChildren:\n"
		output += "  (none)\n"
	}

	if len(node.DNSQueries) > 0 {
		output += "\nHostname Context:\n"
		for _, query := range node.DNSQueries {
			output += fmt.Sprintf("  %s (%s)\n", query.Query, query.QueryType)
		}
	} else {
		output += "\nHostname Context:\n"
		output += "  (none)\n"
	}

	if len(node.NetworkConns) > 0 {
		output += "\nNetwork Connections:\n"
		for _, conn := range node.NetworkConns {
			output += fmt.Sprintf("  %s://%s:%d -> %s:%d (%s)\n",
				conn.Protocol, conn.LocalIP, conn.LocalPort,
				conn.RemoteIP, conn.RemotePort, conn.State)
		}
	} else {
		output += "\nNetwork Connections:\n"
		output += "  (none)\n"
	}

	return output
}
