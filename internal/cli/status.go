package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/spf13/cobra"
	"github.com/olekukonko/tablewriter"
)

// StatusCommand represents the status command configuration
type StatusCommand struct {
	config         *CommandConfig
	showVersion    bool
	showIndex      bool
	showSystem     bool
	showHealth     bool
	refresh        bool
	watch          bool
	checkInterval  time.Duration
}

// SystemInfo represents system information
type SystemInfo struct {
	OS              string    `json:"os"`
	Architecture    string    `json:"architecture"`
	Hostname        string    `json:"hostname"`
	Uptime          string    `json:"uptime"`
	MemoryUsed      string    `json:"memory_used"`
	MemoryTotal     string    `json:"memory_total"`
	CPUCores        int       `json:"cpu_cores"`
	LoadAverage     []float64 `json:"load_average"`
	Timestamp       time.Time `json:"timestamp"`
}

// IndexInfo represents index information
type IndexInfo struct {
	TotalDocuments    int       `json:"total_documents"`
	IndexedDocuments  int       `json:"indexed_documents"`
	PendingDocuments  int       `json:"pending_documents"`
	IndexSize         string    `json:"index_size"`
	LastIndexed       time.Time `json:"last_indexed"`
	IndexStatus       string    `json:"index_status"`
	IndexHealth       string    `json:"index_health"`
	IndexType         string    `json:"index_type"`
	Timestamp         time.Time `json:"timestamp"`
}

// HealthStatus represents overall health status
type HealthStatus struct {
	Status      string            `json:"status"`
	Components  map[string]string `json:"components"`
	Issues      []string          `json:"issues"`
	Warnings    []string          `json:"warnings"`
	LastCheck   time.Time         `json:"last_check"`
	ResponseTime time.Duration    `json:"response_time"`
}

// StatusReport represents a complete status report
type StatusReport struct {
	Version     string      `json:"version"`
	System      SystemInfo  `json:"system"`
	Index       IndexInfo   `json:"index"`
	Health      HealthStatus `json:"health"`
	Timestamp   time.Time   `json:"timestamp"`
}

// NewStatusCommand creates a new status command
func NewStatusCommand(config *CommandConfig) *cobra.Command {
	sc := &StatusCommand{
		config:        config,
		checkInterval: time.Second * 30, // default check interval
	}

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show system status and information",
		Long: `Status displays comprehensive information about the Stroidex system,
including system resources, index status, and health information.

Examples:
  stroidex status                           # Show basic status
  stroidex status --verbose                 # Show detailed status
  stroidex status --output json            # Output in JSON format
  stroidex status --system --index          # Show system and index info
  stroidex status --health                 # Show health check
  stroidex status --watch                  # Watch status in real-time
  stroidex status --refresh 10s            # Auto-refresh every 10 seconds`,
		RunE: sc.runStatus,
	}

	// Add status-specific flags
	cmd.Flags().BoolVar(&sc.showVersion, "version", false, "Show version information only")
	cmd.Flags().BoolVar(&sc.showIndex, "index", false, "Show index information only")
	cmd.Flags().BoolVar(&sc.showSystem, "system", false, "Show system information only")
	cmd.Flags().BoolVar(&sc.showHealth, "health", false, "Show health check only")
	cmd.Flags().BoolVar(&sc.refresh, "refresh", false, "Refresh status information")
	cmd.Flags().BoolVar(&sc.watch, "watch", false, "Watch status in real-time")
	cmd.Flags().DurationVar(&sc.checkInterval, "interval", time.Second*30, "Check interval for watch mode")

	return cmd
}

// runStatus executes the status command
func (sc *StatusCommand) runStatus(cmd *cobra.Command, args []string) error {
	// If specific flags are set, show only that information
	if sc.showVersion {
		return sc.showVersionInfo()
	}

	if sc.showIndex {
		indexInfo, err := sc.collectIndexInfo()
		if err != nil {
			return fmt.Errorf("failed to collect index info: %w", err)
		}
		return sc.displayIndexInfo(indexInfo)
	}

	if sc.showSystem {
		systemInfo, err := sc.collectSystemInfo()
		if err != nil {
			return fmt.Errorf("failed to collect system info: %w", err)
		}
		return sc.displaySystemInfo(systemInfo)
	}

	if sc.showHealth {
		health, err := sc.checkHealth()
		if err != nil {
			return fmt.Errorf("failed to perform health check: %w", err)
		}
		return sc.displayHealthStatus(health)
	}

	// Show complete status report
	if sc.watch {
		return sc.watchStatus()
	}

	return sc.showStatusReport()
}

// showStatusReport shows a complete status report
func (sc *StatusCommand) showStatusReport() error {
	report := &StatusReport{
		Version:   "1.0.0",
		Timestamp: time.Now(),
	}

	// Collect system information
	systemInfo, err := sc.collectSystemInfo()
	if err != nil {
		PrintWarning(fmt.Sprintf("Failed to collect system info: %v", err))
	} else {
		report.System = systemInfo
	}

	// Collect index information
	indexInfo, err := sc.collectIndexInfo()
	if err != nil {
		PrintWarning(fmt.Sprintf("Failed to collect index info: %v", err))
	} else {
		report.Index = indexInfo
	}

	// Perform health check
	health, err := sc.checkHealth()
	if err != nil {
		PrintWarning(fmt.Sprintf("Failed to perform health check: %v", err))
	} else {
		report.Health = health
	}

	// Display based on output format
	switch sc.config.OutputFormat {
	case "json":
		return sc.displayStatusJSON(report)
	case "yaml":
		return sc.displayStatusYAML(report)
	default:
		return sc.displayStatusTable(report)
	}
}

// collectSystemInfo collects system information
func (sc *StatusCommand) collectSystemInfo() (SystemInfo, error) {
	// Show progress for system info collection
	pb := NewProgressBar("Collecting system information", 3)
	pb.Start()
	defer pb.Finish()

	// Get hostname
	pb.UpdateTo(1)
	hostname, _ := os.Hostname()

	// Get memory info (placeholder implementation)
	pb.UpdateTo(2)
	memoryTotal := "16GB"
	memoryUsed := "4GB"

	// Get load average (placeholder for non-unix systems)
	pb.UpdateTo(3)
	var loadAverage []float64
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		loadAverage = []float64{0.5, 0.8, 1.2}
	} else {
		loadAverage = []float64{0, 0, 0}
	}

	info := SystemInfo{
		OS:           fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		Architecture: runtime.GOARCH,
		Hostname:     hostname,
		Uptime:       "24h 30m", // placeholder
		MemoryUsed:   memoryUsed,
		MemoryTotal:  memoryTotal,
		CPUCores:     runtime.NumCPU(),
		LoadAverage:  loadAverage,
		Timestamp:    time.Now(),
	}

	return info, nil
}

// collectIndexInfo collects index information
func (sc *StatusCommand) collectIndexInfo() (IndexInfo, error) {
	// Show progress for index info collection
	pb := NewProgressBar("Collecting index information", 3)
	pb.Start()
	defer pb.Finish()

	// This is a placeholder implementation
	// In a real implementation, this would connect to the index engine
	// and get actual statistics

	pb.UpdateTo(1)
	info := IndexInfo{
		TotalDocuments:   1500,
		IndexedDocuments: 1450,
		PendingDocuments: 50,
		IndexSize:        "245MB",
		LastIndexed:      time.Now().Add(-time.Hour * 2),
		IndexStatus:      "active",
		IndexHealth:      "healthy",
		IndexType:        "full-text",
		Timestamp:        time.Now(),
	}

	pb.UpdateTo(3)
	return info, nil
}

// checkHealth performs health checks
func (sc *StatusCommand) checkHealth() (HealthStatus, error) {
	// Show progress for health check
	pb := NewProgressBar("Performing health checks", 5)
	pb.Start()
	defer pb.Finish()

	health := HealthStatus{
		Components: make(map[string]string),
		Issues:      make([]string, 0),
		Warnings:    make([]string, 0),
		LastCheck:   time.Now(),
		ResponseTime: time.Millisecond * 15,
	}

	// Check various components (placeholder implementation)
	pb.UpdateTo(1)
	health.Components["database"] = "healthy"

	pb.UpdateTo(2)
	health.Components["index_engine"] = "healthy"

	pb.UpdateTo(3)
	health.Components["file_system"] = "healthy"

	pb.UpdateTo(4)
	health.Components["memory"] = "ok"
	health.Components["disk_space"] = "warning"

	pb.UpdateTo(5)
	// Add warnings
	health.Warnings = append(health.Warnings, "Disk usage above 80%")

	// Determine overall status
	hasIssues := len(health.Issues) > 0
	hasWarnings := len(health.Warnings) > 0

	switch {
	case hasIssues:
		health.Status = "unhealthy"
	case hasWarnings:
		health.Status = "degraded"
	default:
		health.Status = "healthy"
	}

	return health, nil
}

// displayStatusTable displays status in table format
func (sc *StatusCommand) displayStatusTable(report *StatusReport) error {
	PrintInfo("=== Stroidex Status ===")
	PrintInfo(fmt.Sprintf("Version: %s", report.Version))
	PrintInfo(fmt.Sprintf("Timestamp: %s", report.Timestamp.Format(time.RFC3339)))

	// System information
	if report.System.OS != "" {
		PrintInfo("\n=== System Information ===")
		fmt.Printf("OS:              %s\n", report.System.OS)
		fmt.Printf("Hostname:        %s\n", report.System.Hostname)
		fmt.Printf("CPU Cores:        %d\n", report.System.CPUCores)
		fmt.Printf("Memory:          %s / %s\n", report.System.MemoryUsed, report.System.MemoryTotal)
		fmt.Printf("Uptime:          %s\n", report.System.Uptime)

		if len(report.System.LoadAverage) > 0 {
			fmt.Printf("Load Average:    %.2f, %.2f, %.2f\n",
				report.System.LoadAverage[0],
				report.System.LoadAverage[1],
				report.System.LoadAverage[2])
		}
	}

	// Index information
	if report.Index.TotalDocuments > 0 {
		PrintInfo("\n=== Index Information ===")
		fmt.Printf("Total Documents: %d\n", report.Index.TotalDocuments)
		fmt.Printf("Indexed:         %d\n", report.Index.IndexedDocuments)
		fmt.Printf("Pending:         %d\n", report.Index.PendingDocuments)
		fmt.Printf("Index Size:      %s\n", report.Index.IndexSize)
		fmt.Printf("Last Indexed:    %s\n", report.Index.LastIndexed.Format(time.RFC3339))
		fmt.Printf("Index Status:    %s\n", report.Index.IndexStatus)
		fmt.Printf("Index Health:    %s\n", report.Index.IndexHealth)
		fmt.Printf("Index Type:      %s\n", report.Index.IndexType)
	}

	// Health status
	if report.Health.Status != "" {
		PrintInfo("\n=== Health Status ===")
		fmt.Printf("Overall Status:  %s\n", report.Health.Status)
		fmt.Printf("Response Time:   %v\n", report.Health.ResponseTime)
		fmt.Printf("Last Check:      %s\n", report.Health.LastCheck.Format(time.RFC3339))

		if len(report.Health.Components) > 0 {
			PrintInfo("\nComponents:")
			for component, status := range report.Health.Components {
				fmt.Printf("  %-15s: %s\n", component, status)
			}
		}

		if len(report.Health.Warnings) > 0 {
			PrintWarning("Warnings:")
			for _, warning := range report.Health.Warnings {
				fmt.Printf("  - %s\n", warning)
			}
		}

		if len(report.Health.Issues) > 0 {
			PrintError(fmt.Errorf("Issues detected"))
			for _, issue := range report.Health.Issues {
				fmt.Printf("  - %s\n", issue)
			}
		}
	}

	return nil
}

// displayStatusJSON displays status in JSON format
func (sc *StatusCommand) displayStatusJSON(report *StatusReport) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

// displayStatusYAML displays status in YAML format (placeholder)
func (sc *StatusCommand) displayStatusYAML(report *StatusReport) error {
	// Simple YAML implementation (placeholder)
	fmt.Println("# Stroidex Status")
	fmt.Printf("version: %s\n", report.Version)
	fmt.Printf("timestamp: %s\n", report.Timestamp.Format(time.RFC3339))
	fmt.Println("system:")
	fmt.Printf("  os: %s\n", report.System.OS)
	fmt.Printf("  hostname: %s\n", report.System.Hostname)
	fmt.Printf("  cpu_cores: %d\n", report.System.CPUCores)
	fmt.Printf("  memory_used: %s\n", report.System.MemoryUsed)
	fmt.Printf("  memory_total: %s\n", report.System.MemoryTotal)
	fmt.Println("index:")
	fmt.Printf("  total_documents: %d\n", report.Index.TotalDocuments)
	fmt.Printf("  indexed_documents: %d\n", report.Index.IndexedDocuments)
	fmt.Printf("  index_status: %s\n", report.Index.IndexStatus)
	fmt.Println("health:")
	fmt.Printf("  status: %s\n", report.Health.Status)

	return nil
}

// showVersionInfo shows version information only
func (sc *StatusCommand) showVersionInfo() error {
	PrintInfo("Stroidex CLI")
	fmt.Printf("Version: %s\n", "1.0.0")
	fmt.Printf("Build:    %s\n", "development")
	fmt.Printf("Go:       %s\n", runtime.Version())
	fmt.Printf("OS/Arch:  %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("Built:    %s\n", "2024-01-01T00:00:00Z")

	return nil
}

// displaySystemInfo displays detailed system information
func (sc *StatusCommand) displaySystemInfo(info SystemInfo) error {
	if sc.config.OutputFormat == "table" {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Property", "Value"})
		table.SetAlignment(tablewriter.ALIGN_LEFT)

		data := [][]string{
			{"OS", info.OS},
			{"Architecture", info.Architecture},
			{"Hostname", info.Hostname},
			{"Uptime", info.Uptime},
			{"CPU Cores", fmt.Sprintf("%d", info.CPUCores)},
			{"Memory Used", info.MemoryUsed},
			{"Memory Total", info.MemoryTotal},
			{"Timestamp", info.Timestamp.Format(time.RFC3339)},
		}

		if len(info.LoadAverage) > 0 {
			loadStr := fmt.Sprintf("%.2f, %.2f, %.2f",
				info.LoadAverage[0], info.LoadAverage[1], info.LoadAverage[2])
			data = append(data, []string{"Load Average", loadStr})
		}

		table.AppendBulk(data)
		table.Render()
	} else {
		// Use JSON format for other output types
		data, _ := json.MarshalIndent(info, "", "  ")
		fmt.Println(string(data))
	}

	return nil
}

// displayIndexInfo displays detailed index information
func (sc *StatusCommand) displayIndexInfo(info IndexInfo) error {
	if sc.config.OutputFormat == "table" {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Property", "Value"})
		table.SetAlignment(tablewriter.ALIGN_LEFT)

		completionRate := float64(info.IndexedDocuments) / float64(info.TotalDocuments) * 100

		data := [][]string{
			{"Total Documents", fmt.Sprintf("%d", info.TotalDocuments)},
			{"Indexed Documents", fmt.Sprintf("%d", info.IndexedDocuments)},
			{"Pending Documents", fmt.Sprintf("%d", info.PendingDocuments)},
			{"Completion Rate", fmt.Sprintf("%.1f%%", completionRate)},
			{"Index Size", info.IndexSize},
			{"Last Indexed", info.LastIndexed.Format(time.RFC3339)},
			{"Index Status", info.IndexStatus},
			{"Index Health", info.IndexHealth},
			{"Index Type", info.IndexType},
			{"Timestamp", info.Timestamp.Format(time.RFC3339)},
		}

		table.AppendBulk(data)
		table.Render()
	} else {
		data, _ := json.MarshalIndent(info, "", "  ")
		fmt.Println(string(data))
	}

	return nil
}

// displayHealthStatus displays health status information
func (sc *StatusCommand) displayHealthStatus(health HealthStatus) error {
	if sc.config.OutputFormat == "table" {
		fmt.Printf("Overall Status: %s\n", health.Status)
		fmt.Printf("Response Time:  %v\n", health.ResponseTime)
		fmt.Printf("Last Check:     %s\n", health.LastCheck.Format(time.RFC3339))

		if len(health.Components) > 0 {
			PrintInfo("\nComponents:")
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Component", "Status"})
			table.SetAlignment(tablewriter.ALIGN_LEFT)

			for component, status := range health.Components {
				table.Append([]string{component, status})
			}

			table.Render()
		}

		if len(health.Warnings) > 0 {
			PrintWarning("\nWarnings:")
			for _, warning := range health.Warnings {
				fmt.Printf("  - %s\n", warning)
			}
		}

		if len(health.Issues) > 0 {
			PrintError(fmt.Errorf("Issues detected"))
			for _, issue := range health.Issues {
				fmt.Printf("  - %s\n", issue)
			}
		}
	} else {
		data, _ := json.MarshalIndent(health, "", "  ")
		fmt.Println(string(data))
	}

	return nil
}

// watchStatus watches status in real-time
func (sc *StatusCommand) watchStatus() error {
	PrintInfo(fmt.Sprintf("Watching status (refresh every %v)...", sc.checkInterval))
	PrintInfo("Press Ctrl+C to stop")

	ticker := time.NewTicker(sc.checkInterval)
	defer ticker.Stop()

	// Clear screen initially
	fmt.Print("\033[H\033[2J")

	for {
		select {
		case <-ticker.C:
			// Clear screen and update status
			fmt.Print("\033[H\033[2J")
			fmt.Printf("Last update: %s\n\n", time.Now().Format(time.RFC3339))

			if err := sc.showStatusReport(); err != nil {
				PrintWarning(fmt.Sprintf("Error updating status: %v", err))
			}

			// Add countdown timer
			for i := int(sc.checkInterval.Seconds()); i > 0; i-- {
				fmt.Printf("\rNext update in %2ds...", i)
				time.Sleep(time.Second)
			}
		}
	}
}