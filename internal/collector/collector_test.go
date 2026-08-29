package collector

import (
	"testing"
	"time"

	"github.com/bashkimkasa/procscope/internal/models"
)

func TestCollectorDeduplicatesDNSAndNetworkEvents(t *testing.T) {
	t.Parallel()

	c := NewCollector()
	c.processes[123] = &models.ProcessInfo{PID: 123, Name: "worker", ParentPID: 1}
	c.dnsSeen[123] = map[string]bool{"example.com": true}
	c.netSeen[123] = map[string]bool{"TCP:127.0.0.1:1234->example.com:443:ESTABLISHED": true}

	c.emitNewDNSQueries(123)
	c.emitNewNetworkConnections(123)

	if len(c.dnsSeen[123]) != 1 {
		t.Fatalf("expected dnsSeen to remain deduplicated, got %#v", c.dnsSeen[123])
	}
	if len(c.netSeen[123]) != 1 {
		t.Fatalf("expected netSeen to remain deduplicated, got %#v", c.netSeen[123])
	}

	if c.dnsSeen[123]["example.com"] != true {
		t.Fatal("expected example.com to remain tracked")
	}
	if c.netSeen[123]["TCP:127.0.0.1:1234->example.com:443:ESTABLISHED"] != true {
		t.Fatal("expected network event key to remain tracked")
	}

	_ = time.Now()
}
