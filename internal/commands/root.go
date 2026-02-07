package commands

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "scout",
	Short: "Your file organizer and management assistant",
	Long:  "Scout helps you organize, analyze, and clean up your file system with smart filtering and AI-powered insights.",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// subcommands
}
