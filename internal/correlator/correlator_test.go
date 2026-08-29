package correlator

import (
	"testing"
	"time"

	"github.com/bashkimkasa/procscope/internal/models"
)

func TestCorrelatorAddsDNSQueriesToProcessNode(t *testing.T) {
	t.Parallel()

	c := NewCorrelator()
	p := &models.ProcessInfo{PID: 42, Name: "chrome", ParentPID: 1}
	c.processGraph.Processes[42] = &models.ProcessNode{
		ProcessInfo:  p,
		Children:     make([]*models.ProcessNode, 0),
		DNSQueries:   make([]*models.DNSQuery, 0),
		NetworkConns: make([]*models.NetworkConnection, 0),
		ChildrenPIDs: make([]int, 0),
	}

	event := &models.Event{
		Type:      "dns",
		Timestamp: time.Now(),
		ProcessID: 42,
		Data: &models.DNSQuery{
			Query:     "example.com",
			QueryType: "A",
			Timestamp: time.Now(),
			PID:       42,
		},
	}

	c.correlateEvent(event)

	node := c.processGraph.Processes[42]
	if len(node.DNSQueries) != 1 {
		t.Fatalf("expected 1 DNS query, got %d", len(node.DNSQueries))
	}
	if node.DNSQueries[0].Query != "example.com" {
		t.Fatalf("expected example.com, got %q", node.DNSQueries[0].Query)
	}
}

func TestCorrelatorLinksParentAndChildProcesses(t *testing.T) {
	t.Parallel()

	c := NewCorrelator()
	parent := &models.ProcessInfo{PID: 10, Name: "parent", ParentPID: 0}
	child := &models.ProcessInfo{PID: 11, Name: "child", ParentPID: 10}

	c.processGraph.Processes[parent.PID] = &models.ProcessNode{
		ProcessInfo:  parent,
		Children:     make([]*models.ProcessNode, 0),
		DNSQueries:   make([]*models.DNSQuery, 0),
		NetworkConns: make([]*models.NetworkConnection, 0),
		ChildrenPIDs: make([]int, 0),
	}

	c.processGraph.Processes[child.PID] = &models.ProcessNode{
		ProcessInfo:  child,
		Children:     make([]*models.ProcessNode, 0),
		DNSQueries:   make([]*models.DNSQuery, 0),
		NetworkConns: make([]*models.NetworkConnection, 0),
		ChildrenPIDs: make([]int, 0),
	}

	c.processGraph.Processes[child.PID].Parent = c.processGraph.Processes[parent.PID]
	c.processGraph.Processes[parent.PID].Children = append(c.processGraph.Processes[parent.PID].Children, c.processGraph.Processes[child.PID])
	c.processGraph.Processes[parent.PID].ChildrenPIDs = append(c.processGraph.Processes[parent.PID].ChildrenPIDs, child.PID)

	if len(c.processGraph.Processes[parent.PID].Children) != 1 {
		t.Fatalf("expected parent to have 1 child, got %d", len(c.processGraph.Processes[parent.PID].Children))
	}
	if c.processGraph.Processes[child.PID].Parent == nil || c.processGraph.Processes[child.PID].Parent.ProcessInfo.PID != parent.PID {
		t.Fatal("expected child to point to parent process")
	}
}
