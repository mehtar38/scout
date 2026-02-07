package commands

import (
	"fmt"
	"scout/internal/scanner"

	"github.com/spf13/cobra"
)

var (
	filesOnly bool
	dirsOnly  bool
)

var listCmd = &cobra.Command{
	Use:   "list [path]",
	Short: "List all files and directories",
	Long:  `List all files and directories in the specified path with metadata like size and modification time.`,
	Args:  cobra.MaximumNArgs(1),
	Run:   runList,
}

func init() {
	// flags
	listCmd.Flags().BoolVar(&filesOnly, "files", false, "Show only files")
	listCmd.Flags().BoolVar(&dirsOnly, "dirs", false, "Show only directories")

	rootCmd.AddCommand(listCmd) //Adding command to root
}

func runList(cmd *cobra.Command, args []string) {

	path := "." // Default to current directory
	if len(args) > 0 {
		path = args[0]
	}

	//Create scanner and scan

	scan, err := scanner.New(path)
	if err != nil {
		fmt.Println("Error creating Scanner ", err)
		return
	}

	err = scan.Scan()
	if err != nil {
		fmt.Println("Error scanning the files ", err)
		return
	}

	scan.ComputeDirSize()

	var results []scanner.Metadata

	if filesOnly {
		results = scan.GetFiles()
	} else if dirsOnly {
		results = scan.GetDirectories()
	} else {
		results = scan.Results
	}

	for _, elem := range results {
		fmt.Println(elem)
	}
}
