package cli

import (
	"testing"
	"time"
)

func TestMonitorStatsCollection(t *testing.T) {
	mc := &MonitorCommand{
		config: &CommandConfig{},
		paths:  []string{"."},
	}

	// Test that stats collection doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("collectStats() panicked: %v", r)
		}
	}()

	stats := mc.collectStats()

	if len(stats) == 0 {
		t.Error("collectStats() returned empty stats")
	}

	// Check expected keys
	expectedKeys := []string{"files", "directories", "total_size", "paths", "patterns"}
	for _, key := range expectedKeys {
		if _, exists := stats[key]; !exists {
			t.Errorf("Expected key '%s' in stats", key)
		}
	}
}

func TestMonitorDetectChanges(t *testing.T) {
	mc := &MonitorCommand{
		config: &CommandConfig{},
	}

	// Test that detectChanges doesn't panic and returns a slice
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("detectChanges() panicked: %v", r)
		}
	}()

	events, err := mc.detectChanges()

	if err != nil {
		t.Errorf("detectChanges() returned error: %v", err)
	}

	if events == nil {
		// detectChanges returns empty slice when no changes, which is correct behavior
		t.Log("detectChanges() returned empty events (expected for monitor)")
	}
}

func TestMonitorProcessEvents(t *testing.T) {
	mc := &MonitorCommand{
		config: &CommandConfig{},
	}

	events := []string{"file1.txt", "file2.md"}

	// Test that processEvents doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("processEvents() panicked: %v", r)
		}
	}()

	err := mc.processEvents(nil, events)
	if err != nil {
		t.Errorf("processEvents() returned error: %v", err)
	}
}

func TestMonitorPrintSummary(t *testing.T) {
	mc := &MonitorCommand{
		config: &CommandConfig{},
	}

	// Test that printSummary doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("printSummary() panicked: %v", r)
		}
	}()

	mc.printSummary(10, time.Now().Add(-time.Minute))
}

// Benchmarks
func BenchmarkMonitorDetectChanges(b *testing.B) {
	mc := &MonitorCommand{
		config: &CommandConfig{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = mc.detectChanges()
	}
}