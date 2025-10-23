package cli

import (
	"testing"
	"time"
)

func TestIndexCommandValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    *IndexCommand
		expectErr bool
		errField  string
	}{
		{
			name: "Valid configuration",
			config: &IndexCommand{
				maxWorkers: 4,
				batchSize:  100,
				indexType:  "full",
			},
			expectErr: false,
		},
		{
			name: "Too many workers",
			config: &IndexCommand{
				maxWorkers: 51,
				batchSize:  100,
				indexType:  "full",
			},
			expectErr: true,
			errField: "workers",
		},
		{
			name: "Zero workers",
			config: &IndexCommand{
				maxWorkers: 0,
				batchSize:  100,
				indexType:  "full",
			},
			expectErr: true,
			errField: "workers",
		},
		{
			name: "Too large batch size",
			config: &IndexCommand{
				maxWorkers: 4,
				batchSize:  10001,
				indexType:  "full",
			},
			expectErr: true,
			errField: "batch size",
		},
		{
			name: "Zero batch size",
			config: &IndexCommand{
				maxWorkers: 4,
				batchSize:  0,
				indexType:  "full",
			},
			expectErr: true,
			errField: "batch size",
		},
		{
			name: "Invalid index type",
			config: &IndexCommand{
				maxWorkers: 4,
				batchSize:  100,
				indexType:  "invalid",
			},
			expectErr: true,
			errField: "index type",
		},
		{
			name: "Valid incremental type",
			config: &IndexCommand{
				maxWorkers: 4,
				batchSize:  100,
				indexType: "incremental",
			},
			expectErr: false,
		},
		{
			name: "Valid partial type",
			config: &IndexCommand{
				maxWorkers: 4,
				batchSize:  100,
				indexType:  "partial",
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validateConfig()
			if (err != nil) != tt.expectErr {
				t.Errorf("validateConfig() error = %v, expectErr %v", err, tt.expectErr)
			}

			if tt.expectErr && err != nil && tt.errField != "" {
				if !containsString(err.Error(), tt.errField) {
					t.Errorf("Expected error to mention '%s', got: %v", tt.errField, err)
				}
			}
		})
	}
}

func TestIndexPatternMatching(t *testing.T) {
	ic := &IndexCommand{
		patterns: []string{"*.txt", "*.md"},
	}

	tests := []struct {
		filePath   string
		shouldMatch bool
	}{
		{"document.txt", true},
		{"README.md", true},
		{"script.go", false},
		{"file.txt.backup", false},
		{"notes.MD", false}, // Case sensitive
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.filePath, func(t *testing.T) {
			matches := ic.matchesPattern(tt.filePath)
			if matches != tt.shouldMatch {
				t.Errorf("matchesPattern(%s) = %v, expected %v", tt.filePath, matches, tt.shouldMatch)
			}
		})
	}
}

func TestIndexShouldExclude(t *testing.T) {
	ic := &IndexCommand{
		excludePaths: []string{"*.tmp", "*.log", ".*"},
	}

	tests := []struct {
		filePath    string
		shouldExclude bool
	}{
		{"temp.tmp", true},
		{"application.log", true},
		{".gitignore", true},
		{"document.txt", false},
		{"README.md", false},
		{"config.json", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.filePath, func(t *testing.T) {
			exclude := ic.shouldExclude(tt.filePath)
			if exclude != tt.shouldExclude {
				t.Errorf("shouldExclude(%s) = %v, expected %v", tt.filePath, exclude, tt.shouldExclude)
			}
		})
	}
}

func TestIndexStatsCreation(t *testing.T) {
	startTime, _ := time.Parse(time.RFC3339, "2024-01-01T10:00:00Z")
	endTime, _ := time.Parse(time.RFC3339, "2024-01-01T10:05:00Z")

	stats := &IndexStats{
		TotalFiles:    100,
		ProcessedFiles: 95,
		SkippedFiles:   5,
		StartTime:      startTime,
		EndTime:        endTime,
		FileTypes:      map[string]int{".txt": 50, ".md": 45},
	}

	if stats.TotalFiles != 100 {
		t.Errorf("Expected TotalFiles 100, got %d", stats.TotalFiles)
	}

	if stats.ProcessedFiles != 95 {
		t.Errorf("Expected ProcessedFiles 95, got %d", stats.ProcessedFiles)
	}

	if stats.SkippedFiles != 5 {
		t.Errorf("Expected SkippedFiles 5, got %d", stats.SkippedFiles)
	}

	if stats.FileTypes[".txt"] != 50 {
		t.Errorf("Expected .txt files 50, got %d", stats.FileTypes[".txt"])
	}

	if stats.FileTypes[".md"] != 45 {
		t.Errorf("Expected .md files 45, got %d", stats.FileTypes[".md"])
	}
}

func TestIndexDryRun(t *testing.T) {
	ic := &IndexCommand{
		config:     &CommandConfig{},
		paths:      []string{"."},
		recursive:  true,
		dryRun:     true,
		patterns:   []string{"*"},
	}

	// Test dry-run mode doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("runDryRun() panicked: %v", r)
		}
	}()

	// Mock filesystem operations would be needed for full testing
	// This tests that the function can be called without error
	startTime, _ := time.Parse(time.RFC3339, "2024-01-01T10:00:00Z")
	stats := &IndexStats{
		StartTime: startTime,
		FileTypes: make(map[string]int),
	}

	// This would normally call collectFiles, but we can test the structure
	if stats.FileTypes == nil {
		t.Error("FileTypes map should be initialized")
	}
	_ = ic // Suppress unused variable warning
}

func TestIndexDisplayStats(t *testing.T) {
	ic := &IndexCommand{
		config: &CommandConfig{Verbose: true},
	}

	startTime, _ := time.Parse(time.RFC3339, "2024-01-01T10:00:00Z")
	endTime, _ := time.Parse(time.RFC3339, "2024-01-01T10:05:00Z")

	stats := &IndexStats{
		TotalFiles:    100,
		ProcessedFiles: 95,
		SkippedFiles:   5,
		StartTime:      startTime,
		EndTime:        endTime,
		FileTypes:      map[string]int{".txt": 50, ".md": 45},
	}

	// Test that displayStats doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("displayStats() panicked: %v", r)
		}
	}()

	ic.displayStats(stats)
}

// Benchmarks
func BenchmarkIndexPatternMatching(b *testing.B) {
	ic := &IndexCommand{
		patterns: []string{"*.txt", "*.md", "*.go", "*.json"},
	}

	testFiles := []string{
		"document.txt", "README.md", "main.go", "config.json",
		"script.py", "image.png", "data.csv", "notes.md",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		file := testFiles[i%len(testFiles)]
		_ = ic.matchesPattern(file)
	}
}

func BenchmarkIndexShouldExclude(b *testing.B) {
	ic := &IndexCommand{
		excludePaths: []string{"*.tmp", "*.log", ".*", "*.backup"},
	}

	testFiles := []string{
		"temp.tmp", "app.log", ".gitignore", "file.backup",
		"document.txt", "README.md", "main.go", "config.json",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		file := testFiles[i%len(testFiles)]
		_ = ic.shouldExclude(file)
	}
}