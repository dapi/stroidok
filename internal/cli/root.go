package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// NewRootCommand creates and returns the root command
func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stroidex",
		Short: "Stroidex - Document indexing and monitoring CLI",
		Long: `Stroidex CLI is a powerful command-line interface for document indexing,
monitoring file system changes, and managing the Stroidex engine.

For more information, visit: https://github.com/stroidex/stroidex`,
		Version: "1.0.0",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				_ = cmd.Help()
				os.Exit(0)
			}
			fmt.Printf("Unknown command: %s\n", args[0])
			_ = cmd.Help()
			os.Exit(1)
		},
	}

	// Initialize global flags
	initGlobalFlags(cmd)

	// Add custom help command
	cmd.SetHelpCommand(&cobra.Command{
		Use:    "help [command]",
		Short:  "Help about any command",
		Long:   `Help provides help for any command in the application.`,
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				_ = cmd.Parent().Help()
				return
			}
			parent := cmd.Parent()
			sub, _, err := parent.Find(args)
			if err != nil {
				fmt.Printf("Unknown help topic: %q\n", args[0])
				_ = cmd.Parent().Help()
				return
			}
			_ = sub.Help()
		},
	})

	return cmd
}

// initGlobalFlags initializes global flags for the root command
func initGlobalFlags(cmd *cobra.Command) {
	// Configuration options
	cmd.PersistentFlags().StringP("config", "c", "", "Path to configuration file (default is $HOME/.stroidex.yaml)")
	cmd.PersistentFlags().StringP("workspace", "w", ".", "Working directory path")

	// Output options
	cmd.PersistentFlags().StringP("output", "o", "table", "Output format (table, json, yaml)")
	cmd.PersistentFlags().BoolP("no-color", "", false, "Disable colored output")
	cmd.PersistentFlags().BoolP("quiet", "q", false, "Quiet mode (no output except errors)")
	cmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose output")

	// Theme options
	cmd.PersistentFlags().StringP("theme", "t", "default", "Color theme (default, dark, light, none)")

	// Engine options
	cmd.PersistentFlags().StringP("engine-type", "e", "default", "Engine type (default, experimental, legacy)")
	cmd.PersistentFlags().StringP("log-level", "l", "info", "Log level (debug, info, warn, error)")
}

// addPersistentPreRun adds persistent pre-run functionality
func addPersistentPreRun(cmd *cobra.Command, config *CommandConfig) {
	cmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		// Handle quiet and verbose flags
		if quiet, _ := cmd.Flags().GetBool("quiet"); quiet {
			config.Quiet = true
			config.Verbose = false
		}

		if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
			config.Verbose = true
			config.Quiet = false
		}

		// Handle output format
		if outputFormat, _ := cmd.Flags().GetString("output"); outputFormat != "" {
			config.OutputFormat = outputFormat
		}

		// Handle theme
		if theme, _ := cmd.Flags().GetString("theme"); theme != "" {
			config.Theme = theme
		}

		// Handle config file
		if configFile, _ := cmd.Flags().GetString("config"); configFile != "" {
			config.ConfigFile = configFile
		}

		// Validate configuration
		if err := validateConfig(config); err != nil {
			PrintError(fmt.Errorf("configuration validation failed: %w", err))
		}
	}
}

// validateConfig validates the command configuration
func validateConfig(config *CommandConfig) error {
	// Validate output format
	validFormats := map[string]bool{
		"table": true,
		"json":  true,
		"yaml":  true,
	}

	if !validFormats[config.OutputFormat] {
		return fmt.Errorf("invalid output format: %s (valid: table, json, yaml)", config.OutputFormat)
	}

	// Validate theme
	validThemes := map[string]bool{
		"default": true,
		"dark":    true,
		"light":   true,
		"none":    true,
	}

	if !validThemes[config.Theme] {
		return fmt.Errorf("invalid theme: %s (valid: default, dark, light, none)", config.Theme)
	}

	return nil
}