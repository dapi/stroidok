package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// IndexCommand represents the index command configuration
type IndexCommand struct {
	config       *CommandConfig
	paths        []string
	recursive    bool
	dryRun       bool
	force        bool
	patterns     []string
	excludePaths []string
	maxWorkers   int
	batchSize    int
	indexType    string
}

// IndexStats represents indexing statistics
type IndexStats struct {
	TotalFiles    int
	ProcessedFiles int
	SkippedFiles   int
	Errors         []error
	Duration       time.Duration
	StartTime      time.Time
	EndTime        time.Time
	FileTypes      map[string]int
}

// NewIndexCommand creates a new index command
func NewIndexCommand(config *CommandConfig) *cobra.Command {
	ic := &IndexCommand{
		config:     config,
		maxWorkers: 4,    // default number of workers
		batchSize:  100,  // default batch size
		indexType:  "full", // default index type
	}

	cmd := &cobra.Command{
		Use:   "index [path...]",
		Short: "Index documents and files",
		Long: `Index processes documents and files in specified paths and adds them
to the Stroidex search index.

Examples:
  stroidex index ./docs                    # Index docs directory
  stroidex index ./src ./docs -r           # Index recursively
  stroidex index . --dry-run               # Show what would be indexed
  stroidex index . --force                 # Force reindex all files
  stroidex index . --pattern "*.md,*.txt"  # Index specific file patterns
  stroidex index . --exclude "*.tmp,*.log" # Exclude specific patterns
  stroidex index . --workers 8              # Use 8 concurrent workers
  stroidex index . --batch-size 200         # Process in batches of 200`,
		Args: cobra.ArbitraryArgs,
		RunE: ic.runIndex,
	}

	// Add index-specific flags
	cmd.Flags().BoolVarP(&ic.recursive, "recursive", "r", true, "Index directories recursively")
	cmd.Flags().BoolVar(&ic.dryRun, "dry-run", false, "Show what would be indexed without processing")
	cmd.Flags().BoolVar(&ic.force, "force", false, "Force reindex all files (ignore existing index)")
	cmd.Flags().StringSliceVarP(&ic.patterns, "pattern", "p", []string{"*"}, "File patterns to index (comma-separated)")
	cmd.Flags().StringSliceVarP(&ic.excludePaths, "exclude", "e", []string{}, "Exclude patterns (comma-separated)")
	cmd.Flags().IntVar(&ic.maxWorkers, "workers", 4, "Number of concurrent workers")
	cmd.Flags().IntVar(&ic.batchSize, "batch-size", 100, "Batch size for processing")
	cmd.Flags().StringVarP(&ic.indexType, "type", "t", "full", "Index type (full, incremental, partial)")

	return cmd
}

// runIndex executes the index command
func (ic *IndexCommand) runIndex(cmd *cobra.Command, args []string) error {
	// Parse paths
	if len(args) == 0 {
		ic.paths = []string{"."}
	} else {
		ic.paths = args
	}

	// Validate paths
	for _, path := range ic.paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", path)
		}
	}

	// Validate configuration
	if err := ic.validateConfig(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Setup context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize statistics
	stats := &IndexStats{
		StartTime:  time.Now(),
		FileTypes:  make(map[string]int),
		Errors:     make([]error, 0),
	}

	PrintInfo(fmt.Sprintf("Starting indexing on %d path(s)", len(ic.paths)))
	for _, path := range ic.paths {
		absPath, _ := filepath.Abs(path)
		PrintInfo(fmt.Sprintf("Indexing: %s (recursive: %v)", absPath, ic.recursive))
	}

	if ic.dryRun {
		return ic.runDryRun(ctx, stats)
	}

	return ic.runFullIndex(ctx, stats)
}

// validateConfig validates the index command configuration
func (ic *IndexCommand) validateConfig() error {
	// Validate workers count
	if ic.maxWorkers < 1 || ic.maxWorkers > 50 {
		return fmt.Errorf("workers count must be between 1 and 50, got: %d", ic.maxWorkers)
	}

	// Validate batch size
	if ic.batchSize < 1 || ic.batchSize > 10000 {
		return fmt.Errorf("batch size must be between 1 and 10000, got: %d", ic.batchSize)
	}

	// Validate index type
	validTypes := map[string]bool{
		"full":        true,
		"incremental": true,
		"partial":     true,
	}

	if !validTypes[ic.indexType] {
		return fmt.Errorf("invalid index type: %s (valid: full, incremental, partial)", ic.indexType)
	}

	return nil
}

// runDryRun performs a dry run of indexing
func (ic *IndexCommand) runDryRun(ctx context.Context, stats *IndexStats) error {
	PrintInfo("Running in dry-run mode (no processing)")

	files, err := ic.collectFiles(ctx)
	if err != nil {
		return fmt.Errorf("failed to collect files: %w", err)
	}

	stats.TotalFiles = len(files)

	PrintInfo(fmt.Sprintf("Found %d files to index", len(files)))

	// Group files by type
	fileTypes := make(map[string]int)
	for _, file := range files {
		ext := strings.ToLower(filepath.Ext(file))
		if ext == "" {
			ext = "no_extension"
		}
		fileTypes[ext]++
	}

	// Display file type statistics
	PrintInfo("=== File Types ===")
	for ext, count := range fileTypes {
		PrintInfo(fmt.Sprintf("  %s: %d files", ext, count))
	}

	// Show sample files
	if len(files) > 0 {
		PrintInfo("=== Sample Files ===")
		max := 5
		if len(files) < max {
			max = len(files)
		}
		for i := 0; i < max; i++ {
			PrintInfo(fmt.Sprintf("  %s", files[i]))
		}
		if len(files) > max {
			PrintInfo(fmt.Sprintf("  ... and %d more files", len(files)-max))
		}
	}

	return nil
}

// runFullIndex performs full indexing
func (ic *IndexCommand) runFullIndex(ctx context.Context, stats *IndexStats) error {
	PrintInfo(fmt.Sprintf("Running full indexing with %d workers", ic.maxWorkers))

	files, err := ic.collectFiles(ctx)
	if err != nil {
		return fmt.Errorf("failed to collect files: %w", err)
	}

	stats.TotalFiles = len(files)

	if len(files) == 0 {
		PrintWarning("No files found to index")
		return nil
	}

	PrintInfo(fmt.Sprintf("Starting to index %d files...", len(files)))

	// Create overall progress bar
	totalPB := NewProgressBar("Indexing files", int64(len(files)))
	totalPB.Start()
	defer totalPB.Finish()

	// Process files in batches
	processedFiles := 0
	for i := 0; i < len(files); i += ic.batchSize {
		end := i + ic.batchSize
		if end > len(files) {
			end = len(files)
		}

		batch := files[i:end]

		batchProcessed, batchErrors := ic.processBatch(ctx, batch, stats)
		processedFiles += batchProcessed
		stats.Errors = append(stats.Errors, batchErrors...)

		// Update overall progress
		totalPB.UpdateTo(int64(end))

		// Check for context cancellation
		select {
		case <-ctx.Done():
			PrintInfo("Indexing cancelled")
			return ctx.Err()
		default:
		}
	}

	stats.ProcessedFiles = processedFiles
	stats.SkippedFiles = stats.TotalFiles - processedFiles
	stats.EndTime = time.Now()
	stats.Duration = stats.EndTime.Sub(stats.StartTime)

	// Clear progress line and display final statistics
	ClearLine()
	ic.displayStats(stats)

	return nil
}

// collectFiles collects all files to be indexed
func (ic *IndexCommand) collectFiles(ctx context.Context) ([]string, error) {
	var files []string

	for _, path := range ic.paths {
		err := filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
			if err != nil {
				if ic.config.Verbose {
					PrintWarning(fmt.Sprintf("Error accessing %s: %v", walkPath, err))
				}
				return nil // Skip errors
			}

			// Skip directories unless we're at the root
			if info.IsDir() {
				if !ic.recursive && walkPath != path {
					return filepath.SkipDir
				}
				return nil
			}

			// Check if file matches patterns
			if !ic.matchesPattern(walkPath) {
				return nil
			}

			// Check if file should be excluded
			if ic.shouldExclude(walkPath) {
				if ic.config.Verbose {
					PrintInfo(fmt.Sprintf("Excluding: %s", walkPath))
				}
				return nil
			}

			files = append(files, walkPath)
			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("error walking path %s: %w", path, err)
		}
	}

	return files, nil
}

// matchesPattern checks if file matches inclusion patterns
func (ic *IndexCommand) matchesPattern(filePath string) bool {
	if len(ic.patterns) == 1 && ic.patterns[0] == "*" {
		return true
	}

	fileName := filepath.Base(filePath)
	for _, pattern := range ic.patterns {
		matched, err := filepath.Match(pattern, fileName)
		if err != nil {
			continue
		}
		if matched {
			return true
		}
	}

	return false
}

// shouldExclude checks if file should be excluded
func (ic *IndexCommand) shouldExclude(filePath string) bool {
	fileName := filepath.Base(filePath)
	for _, pattern := range ic.excludePaths {
		matched, err := filepath.Match(pattern, fileName)
		if err != nil {
			continue
		}
		if matched {
			return true
		}
	}
	return false
}

// processBatch processes a batch of files
func (ic *IndexCommand) processBatch(ctx context.Context, files []string, stats *IndexStats) (int, []error) {
	processed := 0
	var errors []error

	// Create progress bar for this batch
	batchNum := (len(files) + ic.batchSize - 1) / ic.batchSize
	pb := NewProgressBar(fmt.Sprintf("Processing batch %d", batchNum), int64(len(files)))
	pb.Start()
	defer pb.Finish()

	for _, file := range files {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return processed, errors
		default:
		}

		// Process file (placeholder implementation)
		err := ic.processFile(file, stats)
		if err != nil {
			errors = append(errors, fmt.Errorf("error processing %s: %w", file, err))
			if ic.config.Verbose {
				PrintWarning(fmt.Sprintf("Error processing %s: %v", file, err))
			}
			continue
		}

		processed++

		// Update file type statistics
		ext := strings.ToLower(filepath.Ext(file))
		if ext == "" {
			ext = "no_extension"
		}
		stats.FileTypes[ext]++

		// Update progress bar
		pb.Update()
	}

	return processed, errors
}

// processFile processes a single file (placeholder)
func (ic *IndexCommand) processFile(filePath string, stats *IndexStats) error {
	// In a real implementation, this would:
	// 1. Read file content
	// 2. Extract text and metadata
	// 3. Analyze content
	// 4. Add to search index

	if ic.config.Verbose {
		PrintInfo(fmt.Sprintf("Processing: %s", filePath))
	}

	// Simulate processing time
	time.Sleep(time.Millisecond * 10)

	return nil
}

// displayStats displays indexing statistics
func (ic *IndexCommand) displayStats(stats *IndexStats) {
	PrintInfo("=== Indexing Summary ===")
	PrintInfo(fmt.Sprintf("Total files found: %d", stats.TotalFiles))
	PrintInfo(fmt.Sprintf("Files processed: %d", stats.ProcessedFiles))
	PrintInfo(fmt.Sprintf("Files skipped: %d", stats.SkippedFiles))
	PrintInfo(fmt.Sprintf("Processing time: %v", stats.Duration.Round(time.Millisecond)))

	if len(stats.Errors) > 0 {
		PrintWarning(fmt.Sprintf("Errors encountered: %d", len(stats.Errors)))
		if ic.config.Verbose {
			for _, err := range stats.Errors {
				PrintWarning(fmt.Sprintf("  %v", err))
			}
		}
	}

	if stats.TotalFiles > 0 && stats.Duration > 0 {
		rate := float64(stats.ProcessedFiles) / stats.Duration.Seconds()
		PrintInfo(fmt.Sprintf("Processing rate: %.2f files/second", rate))
	}

	PrintInfo("=== File Types Processed ===")
	for ext, count := range stats.FileTypes {
		PrintInfo(fmt.Sprintf("  %s: %d files", ext, count))
	}

	successRate := float64(stats.ProcessedFiles) / float64(stats.TotalFiles) * 100
	PrintInfo(fmt.Sprintf("Success rate: %.1f%%", successRate))

	if len(stats.Errors) == 0 {
		PrintSuccess("Indexing completed successfully!")
	} else {
		PrintWarning("Indexing completed with errors")
	}
}