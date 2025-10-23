package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

// MonitorCommand represents the monitor command configuration
type MonitorCommand struct {
	config     *CommandConfig
	paths      []string
	recursive  bool
	interval   time.Duration
	daemon     bool
	statsOnly  bool
	followMode bool
	patterns   []string
}

// NewMonitorCommand creates a new monitor command
func NewMonitorCommand(config *CommandConfig) *cobra.Command {
	mc := &MonitorCommand{
		config: config,
	}

	cmd := &cobra.Command{
		Use:   "monitor [path...]",
		Short: "Monitor file system changes and document updates",
		Long: `Monitor watches specified paths for file system changes and automatically
triggers document indexing when changes are detected.

Examples:
  stroidex monitor ./docs                    # Monitor docs directory
  stroidex monitor ./src ./docs -r           # Monitor recursively
  stroidex monitor . --interval 5s           # Check every 5 seconds
  stroidex monitor . --daemon                # Run as daemon
  stroidex monitor . --stats-only           # Show stats only
  stroidex monitor . --pattern "*.md,*.txt"  # Monitor specific file patterns`,
		Args: cobra.ArbitraryArgs,
		RunE: mc.runMonitor,
	}

	// Add monitor-specific flags
	cmd.Flags().BoolVarP(&mc.recursive, "recursive", "r", false, "Monitor directories recursively")
	cmd.Flags().DurationVarP(&mc.interval, "interval", "i", time.Second*10, "Monitoring interval (e.g., 1s, 1m, 1h)")
	cmd.Flags().BoolVar(&mc.daemon, "daemon", false, "Run as daemon process")
	cmd.Flags().BoolVar(&mc.statsOnly, "stats-only", false, "Show monitoring statistics without processing")
	cmd.Flags().BoolVarP(&mc.followMode, "follow", "f", false, "Follow file changes in real-time")
	cmd.Flags().StringSliceVarP(&mc.patterns, "pattern", "p", []string{"*"}, "File patterns to monitor (comma-separated)")

	return cmd
}

// runMonitor executes the monitor command
func (mc *MonitorCommand) runMonitor(cmd *cobra.Command, args []string) error {
	// Parse paths
	if len(args) == 0 {
		mc.paths = []string{"."}
	} else {
		mc.paths = args
	}

	// Validate paths
	for _, path := range mc.paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", path)
		}
	}

	// Setup context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start monitoring
	PrintInfo(fmt.Sprintf("Starting monitoring on %d path(s)", len(mc.paths)))
	for _, path := range mc.paths {
		absPath, _ := filepath.Abs(path)
		PrintInfo(fmt.Sprintf("Watching: %s (recursive: %v)", absPath, mc.recursive))
	}

	if mc.statsOnly {
		return mc.runStatsMode(ctx)
	}

	if mc.daemon {
		return mc.runDaemonMode(ctx, sigChan)
	}

	return mc.runInteractiveMode(ctx, sigChan)
}

// runStatsMode runs monitor in statistics-only mode
func (mc *MonitorCommand) runStatsMode(ctx context.Context) error {
	PrintInfo("Running in statistics mode (no processing)")

	stats := mc.collectStats()
	mc.displayStats(stats)

	return nil
}

// runDaemonMode runs monitor as a daemon
func (mc *MonitorCommand) runDaemonMode(ctx context.Context, sigChan chan os.Signal) error {
	PrintInfo("Starting daemon mode...")
	PrintInfo("Use SIGINT (Ctrl+C) or SIGTERM to stop")

	// Create status spinner
	spinner := NewSpinner("Monitoring filesystem")
	spinner.Start()
	defer spinner.Stop()

	// Main daemon loop
	ticker := time.NewTicker(mc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			PrintInfo("Daemon stopped")
			return nil
		case <-sigChan:
			PrintInfo("Received shutdown signal")
			return mc.gracefulShutdown(ctx)
		case <-ticker.C:
			if err := mc.processChanges(ctx); err != nil {
				PrintWarning(fmt.Sprintf("Error processing changes: %v", err))
			}
		}
	}
}

// runInteractiveMode runs monitor in interactive mode
func (mc *MonitorCommand) runInteractiveMode(ctx context.Context, sigChan chan os.Signal) error {
	PrintInfo("Starting interactive monitoring...")
	PrintInfo("Press Ctrl+C to stop")

	// Create monitoring spinner
	spinner := NewSpinner("Monitoring for changes")
	spinner.Start()
	defer spinner.Stop()

	// Interactive monitoring loop
	ticker := time.NewTicker(mc.interval)
	defer ticker.Stop()

	eventCount := 0
	startTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			mc.printSummary(eventCount, startTime)
			return nil
		case <-sigChan:
			mc.printSummary(eventCount, startTime)
			return mc.gracefulShutdown(ctx)
		case <-ticker.C:
			events, err := mc.detectChanges()
			if err != nil {
				PrintWarning(fmt.Sprintf("Error detecting changes: %v", err))
				continue
			}

			if len(events) > 0 {
				// Stop spinner temporarily to show events
				spinner.Stop()

				eventCount += len(events)
				PrintSuccess(fmt.Sprintf("Detected %d change(s)", len(events)))

				if err := mc.processEvents(ctx, events); err != nil {
					PrintWarning(fmt.Sprintf("Error processing events: %v", err))
				}

				// Restart spinner
				spinner.Start()
			}
		}
	}
}

// collectStats collects monitoring statistics
func (mc *MonitorCommand) collectStats() map[string]interface{} {
	stats := make(map[string]interface{})

	fileCount := 0
	dirCount := 0
	totalSize := int64(0)

	for _, path := range mc.paths {
		filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip errors
			}

			if !mc.recursive && walkPath != path && info.IsDir() {
				return filepath.SkipDir
			}

			if info.IsDir() {
				dirCount++
			} else {
				fileCount++
				totalSize += info.Size()
			}

			return nil
		})
	}

	stats["files"] = fileCount
	stats["directories"] = dirCount
	stats["total_size"] = totalSize
	stats["paths"] = len(mc.paths)
	stats["patterns"] = mc.patterns

	return stats
}

// displayStats displays monitoring statistics
func (mc *MonitorCommand) displayStats(stats map[string]interface{}) {
	PrintInfo("=== Monitoring Statistics ===")
	PrintInfo(fmt.Sprintf("Paths monitored: %v", stats["paths"]))
	PrintInfo(fmt.Sprintf("Files found: %v", stats["files"]))
	PrintInfo(fmt.Sprintf("Directories found: %v", stats["directories"]))
	PrintInfo(fmt.Sprintf("Total size: %v bytes", stats["total_size"]))
	PrintInfo(fmt.Sprintf("File patterns: %v", stats["patterns"]))
}

// detectChanges detects file system changes
func (mc *MonitorCommand) detectChanges() ([]string, error) {
	// This is a placeholder implementation
	// In a real implementation, this would use file system watchers
	// or compare file modification times

	// For now, return empty slice
	return []string{}, nil
}

// processEvents processes detected events
func (mc *MonitorCommand) processEvents(ctx context.Context, events []string) error {
	for _, event := range events {
		if mc.config.Verbose {
			PrintInfo(fmt.Sprintf("Processing: %s", event))
		}

		// Process the event (placeholder)
		// In a real implementation, this would trigger indexing
	}

	return nil
}

// processChanges processes file system changes
func (mc *MonitorCommand) processChanges(ctx context.Context) error {
	// Placeholder implementation
	// In a real implementation, this would scan for changes
	// and trigger appropriate indexing actions

	if mc.config.Verbose {
		PrintInfo("Scanning for changes...")
	}

	return nil
}

// gracefulShutdown performs graceful shutdown
func (mc *MonitorCommand) gracefulShutdown(ctx context.Context) error {
	PrintInfo("Performing graceful shutdown...")

	// Perform cleanup operations
	// In a real implementation, this would:
	// - Stop all file watchers
	// - Complete in-progress indexing
	// - Save state

	PrintSuccess("Shutdown complete")
	return nil
}

// printSummary prints monitoring summary
func (mc *MonitorCommand) printSummary(eventCount int, startTime time.Time) {
	duration := time.Since(startTime)
	PrintInfo("=== Monitoring Summary ===")
	PrintInfo(fmt.Sprintf("Duration: %v", duration.Round(time.Second)))
	PrintInfo(fmt.Sprintf("Total events: %d", eventCount))

	if duration > 0 {
		rate := float64(eventCount) / duration.Seconds()
		PrintInfo(fmt.Sprintf("Event rate: %.2f events/second", rate))
	}
}