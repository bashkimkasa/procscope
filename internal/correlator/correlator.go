package correlator

import (
	"github.com/bashkimkasa/procscope/internal/models"
)

// Correlator correlates events into process-centric relationships
type Correlator struct {
	name         string
	events       []*models.Event
	processGraph *models.ProcessGraph
	eventChan    chan *models.Event
	updatesChan  chan *models.Event
}

// NewCorrelator creates a new correlator instance
func NewCorrelator() *Correlator {
	return &Correlator{
		name:         "EventCorrelator",
		events:       make([]*models.Event, 0),
		processGraph: &models.ProcessGraph{Processes: make(map[int]*models.ProcessNode)},
		eventChan:    make(chan *models.Event, 100),
		updatesChan:  make(chan *models.Event, 100),
	}
}

// Name returns the correlator name
func (c *Correlator) Name() string {
	return c.name
}

// GetEventChannel returns the input event channel
func (c *Correlator) GetEventChannel() chan<- *models.Event {
	return c.eventChan
}

// GetUpdatesChannel returns the output updates channel
func (c *Correlator) GetUpdatesChannel() <-chan *models.Event {
	return c.updatesChan
}

// Start begins correlating events
func (c *Correlator) Start() {
	go c.correlateEvents()
}

// Stop stops the correlator
func (c *Correlator) Stop() {
	close(c.eventChan)
	close(c.updatesChan)
}

// correlateEvents reads events and correlates them
func (c *Correlator) correlateEvents() {
	for event := range c.eventChan {
		c.correlateEvent(event)
	}
}

// CorrelateEvent adds an event and attempts to correlate it
func (c *Correlator) correlateEvent(event *models.Event) {
	c.events = append(c.events, event)

	switch event.Type {
	case "process_start":
		c.handleProcessStart(event)
	case "process_exit":
		c.handleProcessExit(event)
	case "network":
		c.handleNetworkEvent(event)
	case "dns":
		c.handleDNSEvent(event)
	}

	c.updatesChan <- event
}

// handleProcessStart handles process start events
func (c *Correlator) handleProcessStart(event *models.Event) {
	pinfo, ok := event.Data.(*models.ProcessInfo)
	if !ok {
		return
	}

	// Create new process node
	node := &models.ProcessNode{
		ProcessInfo:  pinfo,
		Children:     make([]*models.ProcessNode, 0),
		DNSQueries:   make([]*models.DNSQuery, 0),
		NetworkConns: make([]*models.NetworkConnection, 0),
		ChildrenPIDs: make([]int, 0),
	}

	c.processGraph.Processes[event.ProcessID] = node

	// Link to parent if it exists
	if pinfo.ParentPID > 0 {
		if parent, exists := c.processGraph.Processes[pinfo.ParentPID]; exists {
			node.Parent = parent
			parent.Children = append(parent.Children, node)
			parent.ChildrenPIDs = append(parent.ChildrenPIDs, event.ProcessID)
		}
	}
}

// handleProcessExit handles process exit events
func (c *Correlator) handleProcessExit(event *models.Event) {
	// Remove from graph but keep historical data
	if node, exists := c.processGraph.Processes[event.ProcessID]; exists {
		if node.Parent != nil {
			// Remove from parent's children
			for i, child := range node.Parent.Children {
				if child.ProcessInfo.PID == event.ProcessID {
					node.Parent.Children = append(node.Parent.Children[:i], node.Parent.Children[i+1:]...)
					break
				}
			}
		}
	}
}

// handleNetworkEvent handles network connection events
func (c *Correlator) handleNetworkEvent(event *models.Event) {
	conn, ok := event.Data.(*models.NetworkConnection)
	if !ok {
		return
	}

	if node, exists := c.processGraph.Processes[event.ProcessID]; exists {
		node.NetworkConns = append(node.NetworkConns, conn)
	}
}

// handleDNSEvent handles DNS query events
func (c *Correlator) handleDNSEvent(event *models.Event) {
	query, ok := event.Data.(*models.DNSQuery)
	if !ok {
		return
	}

	if node, exists := c.processGraph.Processes[event.ProcessID]; exists {
		node.DNSQueries = append(node.DNSQueries, query)
	}
}

// GetProcessNode retrieves a process node by PID
func (c *Correlator) GetProcessNode(pid int) *models.ProcessNode {
	return c.processGraph.Processes[pid]
}

// GetAllProcesses returns all process nodes
func (c *Correlator) GetAllProcesses() map[int]*models.ProcessNode {
	return c.processGraph.Processes
}

// GetProcessGraph returns the correlated process graph
func (c *Correlator) GetProcessGraph() *models.ProcessGraph {
	return c.processGraph
}
