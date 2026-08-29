package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/bashkimkasa/procscope/internal/collector"
	"github.com/bashkimkasa/procscope/internal/correlator"
	"github.com/bashkimkasa/procscope/internal/graph"
	"github.com/bashkimkasa/procscope/internal/models"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "watch":
		watchCommand()
	case "explain":
		explainCommand()
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func watchCommand() {
	fmt.Println("procscope watch - Real-time process behavior monitoring")
	fmt.Println("\nInitializing collectors...")

	// Initialize collector
	coll := collector.NewCollector()
	fmt.Printf("Collector initialized: %s\n", coll.Name())

	// Initialize correlator
	fmt.Println("Initializing correlator...")
	cor := correlator.NewCorrelator()
	fmt.Printf("Correlator initialized: %s\n", cor.Name())

	// Initialize graph
	fmt.Println("Initializing process graph...")
	g := graph.NewGraph()
	fmt.Printf("Process graph initialized: %s\n", g.Name())

	// Start collector and correlator
	coll.Start()
	cor.Start()
	defer coll.Stop()
	defer cor.Stop()

	// Pipe events from collector to correlator
	go func() {
		for event := range coll.GetEventChannel() {
			cor.GetEventChannel() <- event
		}
	}()

	// Update graph with correlated data
	go func() {
		for range cor.GetUpdatesChannel() {
			g.SetProcessGraph(cor.GetProcessGraph())
		}
	}()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Print header
	fmt.Println("\nWatching for process events (Ctrl+C to stop)...")
	fmt.Println()

	// Display updates
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sigChan:
			fmt.Println("\n\nStopping procscope...")
			return
		case <-ticker.C:
			// Clear screen and print updated data (simplified for Windows)
			// In a real implementation, would use proper terminal control
			if cor.GetAllProcesses() != nil && len(cor.GetAllProcesses()) > 0 {
				// Print the process table
				fmt.Print(g.FormatTable())
				fmt.Println()
				time.Sleep(500 * time.Millisecond)
			}
		}
	}
}

func explainCommand() {
	// Parse PID argument
	var pidStr string

	flag.CommandLine.Parse(os.Args[2:])

	if len(os.Args) > 2 {
		pidStr = os.Args[2]
	} else {
		fmt.Println("Usage: procscope explain <pid>")
		os.Exit(1)
	}

	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		fmt.Printf("Invalid PID: %s\n", pidStr)
		os.Exit(1)
	}

	fmt.Printf("procscope explain - Detailed process behavior for PID %d\n", pid)
	fmt.Println()

	// Initialize collector to get a snapshot of processes
	coll := collector.NewCollector()

	// Get processes synchronously
	processes := coll.GetRunningProcesses()

	// Build process graph directly
	g := graph.NewGraph()
	pg := &models.ProcessGraph{Processes: make(map[int]*models.ProcessNode)}

	for pid, pinfo := range processes {
		node := &models.ProcessNode{
			ProcessInfo:  pinfo,
			Children:     make([]*models.ProcessNode, 0),
			DNSQueries:   make([]*models.DNSQuery, 0),
			NetworkConns: make([]*models.NetworkConnection, 0),
			ChildrenPIDs: make([]int, 0),
		}
		pg.Processes[pid] = node
	}

	// Link parent-child relationships
	for _, node := range pg.Processes {
		if node.ProcessInfo.ParentPID > 0 {
			if parent, exists := pg.Processes[node.ProcessInfo.ParentPID]; exists {
				node.Parent = parent
				parent.Children = append(parent.Children, node)
			}
		}
	}

	// Get network connections and DNS queries for this process
	conns := coll.GetNetworkConnections(pid)
	dnsQueries := coll.GetDNSQueries(pid)
	if node, exists := pg.Processes[pid]; exists {
		node.NetworkConns = conns
		node.DNSQueries = dnsQueries
	}

	// Set graph and display
	g.SetProcessGraph(pg)
	fmt.Print(g.FormatProcessDetail(pid))
}

func printUsage() {
	fmt.Println("procscope - Lightweight, cross-platform process behavior observability")
	fmt.Println("\nUsage:")
	fmt.Println("  procscope watch                 Watch process activity in real-time")
	fmt.Println("  procscope explain <pid>         Explain detailed behavior for a process")
	fmt.Println("  procscope help                  Show this help message")
}
