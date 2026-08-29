package collector

import (
	"time"

	"github.com/bashkimkasa/procscope/internal/models"
	"github.com/bashkimkasa/procscope/internal/platform"
)

// Collector wraps the platform-specific collector
type Collector struct {
	name      string
	platform  platform.PlatformCollector
	processes map[int]*models.ProcessInfo
	eventChan chan *models.Event
}

// NewCollector creates a new collector instance
func NewCollector() *Collector {
	return &Collector{
		name:      "OSCollector",
		platform:  platform.NewCollector(),
		processes: make(map[int]*models.ProcessInfo),
		eventChan: make(chan *models.Event, 100),
	}
}

// Name returns the collector name
func (c *Collector) Name() string {
	return c.name
}

// GetEventChannel returns the event channel
func (c *Collector) GetEventChannel() <-chan *models.Event {
	return c.eventChan
}

// Start begins collecting events
func (c *Collector) Start() error {
	go c.collectProcesses()
	return nil
}

// Stop stops the collector
func (c *Collector) Stop() error {
	close(c.eventChan)
	return nil
}

// collectProcesses collects process information
func (c *Collector) collectProcesses() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		processes := c.GetRunningProcesses()

		// Detect new processes
		for pid, pinfo := range processes {
			if _, exists := c.processes[pid]; !exists {
				c.eventChan <- &models.Event{
					Type:      "process_start",
					Timestamp: time.Now(),
					ProcessID: pid,
					Data:      pinfo,
				}
			}
		}

		// Detect exited processes
		for pid := range c.processes {
			if _, exists := processes[pid]; !exists {
				c.eventChan <- &models.Event{
					Type:      "process_exit",
					Timestamp: time.Now(),
					ProcessID: pid,
					Data:      nil,
				}
			}
		}

		c.processes = processes
	}
}

// GetRunningProcesses delegates to the platform-specific collector
func (c *Collector) GetRunningProcesses() map[int]*models.ProcessInfo {
	return c.platform.GetRunningProcesses()
}

// GetNetworkConnections delegates to the platform-specific collector
func (c *Collector) GetNetworkConnections(pid int) []*models.NetworkConnection {
	return c.platform.GetNetworkConnections(pid)
}

// GetProcessChildren delegates to the platform-specific collector
func (c *Collector) GetProcessChildren(parentPID int) []*models.ProcessInfo {
	return c.platform.GetProcessChildren(parentPID)
}
