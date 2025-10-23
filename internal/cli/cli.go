package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// CLI represents the main CLI structure
type CLI struct {
 RootCmd *cobra.Command
 Config  *CommandConfig
}

// CommandConfig holds configuration for CLI commands
type CommandConfig struct {
	ConfigFile string
	Verbose    bool
	Quiet      bool
	OutputFormat string
	Theme      string
}

// NewCLI creates a new CLI instance
func NewCLI() *CLI {
	config := &CommandConfig{
		OutputFormat: "table", // default output format
		Theme:        "default", // default theme
	}

	cli := &CLI{
		Config: config,
	}

	cli.RootCmd = cli.createRootCommand()
	cli.addCommands()

	return cli
}

// createRootCommand creates the root cobra command
func (cli *CLI) createRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stroidex",
		Short: "Stroidex - Document indexing and monitoring CLI",
		Long: `Stroidex CLI is a powerful command-line interface for document indexing,
monitoring file system changes, and managing the Stroidex engine.`,
		Version: "1.0.0",
	}

	// Global flags
	cmd.PersistentFlags().StringVar(&cli.Config.ConfigFile, "config", "", "config file path")
	cmd.PersistentFlags().BoolVarP(&cli.Config.Verbose, "verbose", "v", false, "verbose output")
	cmd.PersistentFlags().BoolVarP(&cli.Config.Quiet, "quiet", "q", false, "quiet mode")
	cmd.PersistentFlags().StringVarP(&cli.Config.OutputFormat, "output", "o", "table", "output format (table, json, yaml)")
	cmd.PersistentFlags().StringVar(&cli.Config.Theme, "theme", "default", "color theme (default, dark, light, none)")

	// Add custom help and version commands
	// cmd.SetHelpCommand(cmd.HelpCommand())
	cmd.SetVersionTemplate(`{{with .Name}}{{printf "%s " .}}{{end}}{{printf "version %s" .Version}}
`)

	return cmd
}

// addCommands adds all subcommands to the root command
func (cli *CLI) addCommands() {
	cli.RootCmd.AddCommand(NewMonitorCommand(cli.Config))
	cli.RootCmd.AddCommand(NewIndexCommand(cli.Config))
	cli.RootCmd.AddCommand(NewStatusCommand(cli.Config))
	// cli.RootCmd.AddCommand(cli.NewConfigCommand())
}

// Execute executes the CLI
func (cli *CLI) Execute() error {
	return cli.RootCmd.Execute()
}

// PrintError prints formatted error message
func PrintError(err error) {
	fmt.Printf("Error: %v\n", err)
	os.Exit(1)
}

// PrintSuccess prints formatted success message
func PrintSuccess(message string) {
	fmt.Printf("✓ %s\n", message)
}

// PrintInfo prints formatted info message
func PrintInfo(message string) {
	fmt.Printf("ℹ %s\n", message)
}

// PrintWarning prints formatted warning message
func PrintWarning(message string) {
	fmt.Printf("⚠ %s\n", message)
}