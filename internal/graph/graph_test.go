package graph

import (
	"strings"
	"testing"

	"github.com/bashkimkasa/procscope/internal/models"
)

func TestFormatProcessDetailShowsParentName(t *testing.T) {
	t.Parallel()

	parent := &models.ProcessNode{
		ProcessInfo: &models.ProcessInfo{PID: 10, Name: "parent"},
	}
	child := &models.ProcessNode{
		ProcessInfo: &models.ProcessInfo{PID: 11, Name: "child", ParentPID: 10},
		Parent:      parent,
	}

	g := NewGraph()
	g.SetProcessGraph(&models.ProcessGraph{Processes: map[int]*models.ProcessNode{
		10: parent,
		11: child,
	}})

	out := g.FormatProcessDetail(11)
	if !strings.Contains(out, "Parent:        parent (PID 10)") {
		t.Fatalf("expected parent name in output, got:\n%s", out)
	}
}

func TestFormatProcessDetailIncludesBehaviorSummary(t *testing.T) {
	t.Parallel()

	parent := &models.ProcessNode{ProcessInfo: &models.ProcessInfo{PID: 1, Name: "parent"}}
	child := &models.ProcessNode{ProcessInfo: &models.ProcessInfo{PID: 7, Name: "child", ParentPID: 1}}
	child.Parent = parent
	child.Children = []*models.ProcessNode{}
	child.DNSQueries = []*models.DNSQuery{{Query: "api.example.com"}}
	child.NetworkConns = []*models.NetworkConnection{{RemoteIP: "10.0.0.4", RemotePort: 443, Protocol: "TCP", State: "ESTABLISHED"}}

	g := NewGraph()
	g.SetProcessGraph(&models.ProcessGraph{Processes: map[int]*models.ProcessNode{
		1: parent,
		7: child,
	}})

	out := g.FormatProcessDetail(7)
	if !strings.Contains(out, "Summary:") {
		t.Fatalf("expected summary block, got:\n%s", out)
	}
	if !strings.Contains(out, "parent") || !strings.Contains(out, "Children: 0") {
		t.Fatalf("expected parent and child count summary, got:\n%s", out)
	}
	if !strings.Contains(out, "Hostnames: 1") || !strings.Contains(out, "Network connections: 1") {
		t.Fatalf("expected hostname and network summary counts, got:\n%s", out)
	}
}
