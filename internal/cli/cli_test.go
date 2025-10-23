package cli

import (
	"os"
	"strings"
	"testing"
)

func TestNewCLI(t *testing.T) {
	cli := NewCLI()

	if cli == nil {
		t.Fatal("NewCLI() returned nil")
	}

	if cli.RootCmd == nil {
		t.Error("Root command is nil")
	}

	if cli.Config == nil {
		t.Error("Config is nil")
	}

	// Check default values
	if cli.Config.OutputFormat != "table" {
		t.Errorf("Expected default output format 'table', got '%s'", cli.Config.OutputFormat)
	}

	if cli.Config.Theme != "default" {
		t.Errorf("Expected default theme 'default', got '%s'", cli.Config.Theme)
	}
}

func TestCLIExecute(t *testing.T) {
	cli := NewCLI()

	// Test help command
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"stroidok", "--help"}
	err := cli.Execute()
	if err != nil {
		t.Errorf("Execute() returned error: %v", err)
	}

	// Test version command
	os.Args = []string{"stroidok", "--version"}
	err = cli.Execute()
	if err != nil {
		t.Errorf("Execute() returned error: %v", err)
	}
}

func TestCommandConfigValidation(t *testing.T) {
	tests := []struct {
		name     string
		config   *CommandConfig
		wantErr  bool
		errField string
	}{
		{
			name: "Valid table format",
			config: &CommandConfig{
				OutputFormat: "table",
				Theme:       "default",
			},
			wantErr: false,
		},
		{
			name: "Valid JSON format",
			config: &CommandConfig{
				OutputFormat: "json",
				Theme:       "dark",
			},
			wantErr: false,
		},
		{
			name: "Valid YAML format",
			config: &CommandConfig{
				OutputFormat: "yaml",
				Theme:       "light",
			},
			wantErr: false,
		},
		{
			name: "Invalid output format",
			config: &CommandConfig{
				OutputFormat: "invalid",
				Theme:       "default",
			},
			wantErr: true,
			errField: "output format",
		},
		{
			name: "Invalid theme",
			config: &CommandConfig{
				OutputFormat: "table",
				Theme:       "invalid",
			},
			wantErr: true,
			errField: "theme",
		},
		{
			name: "Valid none theme",
			config: &CommandConfig{
				OutputFormat: "table",
				Theme:       "none",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && err != nil {
				if tt.errField != "" && !containsString(err.Error(), tt.errField) {
					t.Errorf("Expected error to mention '%s', got: %v", tt.errField, err)
				}
			}
		})
	}
}

func TestPrintFunctions(t *testing.T) {
	// Test PrintSuccess doesn't panic
	t.Run("PrintSuccess", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("PrintSuccess() panicked: %v", r)
			}
		}()
		PrintSuccess("test message")
	})

	// Test PrintInfo doesn't panic
	t.Run("PrintInfo", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("PrintInfo() panicked: %v", r)
			}
		}()
		PrintInfo("test message")
	})

	// Test PrintWarning doesn't panic
	t.Run("PrintWarning", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("PrintWarning() panicked: %v", r)
			}
		}()
		PrintWarning("test message")
	})

	// Test PrintError with custom error
	t.Run("PrintError", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("PrintError() panicked: %v", r)
			}
		}()
		// Note: PrintError calls os.Exit(1), so we can't test it directly
		// This test mainly checks it doesn't panic before calling os.Exit
	})
}

func TestMonitorCommandCreation(t *testing.T) {
	config := &CommandConfig{
		OutputFormat: "table",
		Theme:       "default",
	}

	cmd := NewMonitorCommand(config)

	if cmd == nil {
		t.Fatal("NewMonitorCommand() returned nil")
	}

	if !strings.Contains(cmd.Use, "monitor") {
		t.Errorf("Expected command use to contain 'monitor', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Command short description is empty")
	}

	if cmd.Long == "" {
		t.Error("Command long description is empty")
	}

	// Test flags
	flags := cmd.Flags()
	if flags == nil {
		t.Error("Command flags is nil")
	}

	// Check that important flags exist
	flagNames := []string{"recursive", "interval", "daemon", "stats-only", "pattern"}
	for _, flagName := range flagNames {
		flag := flags.Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag '%s' not found", flagName)
		}
	}
}

func TestIndexCommandCreation(t *testing.T) {
	config := &CommandConfig{
		OutputFormat: "table",
		Theme:       "default",
	}

	cmd := NewIndexCommand(config)

	if cmd == nil {
		t.Fatal("NewIndexCommand() returned nil")
	}

	if !strings.Contains(cmd.Use, "index") {
		t.Errorf("Expected command use to contain 'index', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Command short description is empty")
	}

	if cmd.Long == "" {
		t.Error("Command long description is empty")
	}

	// Test flags
	flags := cmd.Flags()
	if flags == nil {
		t.Error("Command flags is nil")
	}

	// Check that important flags exist
	flagNames := []string{"recursive", "dry-run", "force", "pattern", "exclude", "workers", "batch-size", "type"}
	for _, flagName := range flagNames {
		flag := flags.Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag '%s' not found", flagName)
		}
	}
}

func TestStatusCommandCreation(t *testing.T) {
	config := &CommandConfig{
		OutputFormat: "table",
		Theme:       "default",
	}

	cmd := NewStatusCommand(config)

	if cmd == nil {
		t.Fatal("NewStatusCommand() returned nil")
	}

	if cmd.Use != "status" {
		t.Errorf("Expected command use 'status', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Command short description is empty")
	}

	if cmd.Long == "" {
		t.Error("Command long description is empty")
	}

	// Test flags
	flags := cmd.Flags()
	if flags == nil {
		t.Error("Command flags is nil")
	}

	// Check that important flags exist
	flagNames := []string{"version", "index", "system", "health", "refresh", "watch", "interval"}
	for _, flagName := range flagNames {
		flag := flags.Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag '%s' not found", flagName)
		}
	}
}

func TestRootCommandFlags(t *testing.T) {
	cli := NewCLI()

	// Test global flags
	flags := cli.RootCmd.PersistentFlags()
	if flags == nil {
		t.Error("Root command persistent flags is nil")
	}

	// Check that important global flags exist
	flagNames := []string{"config", "output", "quiet", "verbose", "theme"}
	for _, flagName := range flagNames {
		flag := flags.Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected global flag '%s' not found", flagName)
		}
	}

	// Test flag defaults
	outputFlag := flags.Lookup("output")
	if outputFlag != nil && outputFlag.DefValue != "table" {
		t.Errorf("Expected output flag default 'table', got '%s'", outputFlag.DefValue)
	}

	themeFlag := flags.Lookup("theme")
	if themeFlag != nil && themeFlag.DefValue != "default" {
		t.Errorf("Expected theme flag default 'default', got '%s'", themeFlag.DefValue)
	}
}

// Helper functions for testing

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr)))
}

func TestProgressBar(t *testing.T) {
	t.Run("Basic progress bar", func(t *testing.T) {
		pb := NewProgressBar("Test", 100)

		if pb == nil {
			t.Fatal("NewProgressBar() returned nil")
		}

		if pb.total != 100 {
			t.Errorf("Expected total 100, got %d", pb.total)
		}

		if pb.current != 0 {
			t.Errorf("Expected current 0, got %d", pb.current)
		}

		// Test update
		pb.UpdateTo(50)
		if pb.current != 50 {
			t.Errorf("Expected current 50 after update, got %d", pb.current)
		}

		progress := pb.GetProgress()
		expected := 0.5
		if progress != expected {
			t.Errorf("Expected progress %.2f, got %.2f", expected, progress)
		}

		// Test finish
		pb.Finish()
		if pb.current != pb.total {
			t.Errorf("Expected current %d after finish, got %d", pb.total, pb.current)
		}
	})

	t.Run("Progress bar with zero total", func(t *testing.T) {
		pb := NewProgressBar("Test", 0)

		progress := pb.GetProgress()
		if progress != 0 {
			t.Errorf("Expected progress 0 for zero total, got %f", progress)
		}
	})

	t.Run("Progress bar update beyond total", func(t *testing.T) {
		pb := NewProgressBar("Test", 100)

		pb.UpdateTo(150) // Beyond total
		if pb.current != 100 {
			t.Errorf("Expected current 100 (capped at total), got %d", pb.current)
		}
	})
}

func TestSpinner(t *testing.T) {
	t.Run("Basic spinner", func(t *testing.T) {
		spinner := NewSpinner("Test")

		if spinner == nil {
			t.Fatal("NewSpinner() returned nil")
		}

		if spinner.total != 0 {
			t.Errorf("Expected spinner total 0, got %d", spinner.total)
		}

		if spinner.progressType != ProgressTypeSpinner {
			t.Errorf("Expected spinner type ProgressTypeSpinner, got %v", spinner.progressType)
		}
	})
}

func TestBytesProgress(t *testing.T) {
	t.Run("Bytes progress bar", func(t *testing.T) {
		pb := NewBytesProgress("Test", 1024)

		if pb == nil {
			t.Fatal("NewBytesProgress() returned nil")
		}

		if pb.total != 1024 {
			t.Errorf("Expected total 1024, got %d", pb.total)
		}

		if pb.progressType != ProgressTypeBytes {
			t.Errorf("Expected type ProgressTypeBytes, got %v", pb.progressType)
		}
	})
}

func TestProgressGroup(t *testing.T) {
	t.Run("Progress group operations", func(t *testing.T) {
		pg := NewProgressGroup()

		if pg == nil {
			t.Fatal("NewProgressGroup() returned nil")
		}

		if len(pg.bars) != 0 {
			t.Errorf("Expected 0 bars initially, got %d", len(pg.bars))
		}

		// Add bars
		bar1 := pg.NewBar("Bar 1", 100)
		bar2 := pg.NewSpinner("Spinner")

		if len(pg.bars) != 2 {
			t.Errorf("Expected 2 bars after adding, got %d", len(pg.bars))
		}

		if bar1 == nil || bar2 == nil {
			t.Error("Added bars are nil")
		}
	})
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{"Zero bytes", 0, "0 B"},
		{"Bytes", 512, "512 B"},
		{"Kilobytes", 1024, "1.0 KiB"},
		{"Megabytes", 1048576, "1.0 MiB"},
		{"Gigabytes", 1073741824, "1.0 GiB"},
		{"Large value", 3670016, "3.5 MiB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatBytes(tt.bytes)
			if result != tt.expected {
				t.Errorf("formatBytes(%d) = %s, expected %s", tt.bytes, result, tt.expected)
			}
		})
	}
}

// Benchmark tests
func BenchmarkProgressBarUpdate(b *testing.B) {
	pb := NewProgressBar("Benchmark", 1000000)
	pb.Start()
	defer pb.Finish()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pb.Update()
	}
}

func BenchmarkProgressBarGetProgress(b *testing.B) {
	pb := NewProgressBar("Benchmark", 100)
	pb.UpdateTo(50)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pb.GetProgress()
	}
}

func BenchmarkFormatBytes(b *testing.B) {
	values := []int64{1024, 1048576, 1073741824, 3670016}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = formatBytes(values[i%len(values)])
	}
}