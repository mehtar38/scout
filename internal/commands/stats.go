package commands

import (
	"fmt"

	"github.com/mehtar38/scout/internal/scanner"

	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats [path]",
	Short: "Show statistics about files and directories",
	Long:  `Display summary statistics including counts, sizes, and file type distribution.`,
	Args:  cobra.MaximumNArgs(1),
	Run:   runStats,
}

func init() {
	rootCmd.AddCommand(statsCmd)
}

func runStats(cmd *cobra.Command, args []string) {
	path := "."
	if len(args) > 0 {
		path = args[0]
	}

	scan, err := scanner.New(path)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	err = scan.Scan()
	if err != nil {
		fmt.Println("Error scanning:", err)
		return
	}

	scan.ComputeDirSize()

	scan.GetStats()
}
