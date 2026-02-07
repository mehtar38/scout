package commands

import (
	"fmt"
	"scout/internal/scanner"
	"strconv"

	"github.com/spf13/cobra"
)

var (
	recentFiles bool
	recentDirs  bool
)

var recentCmd = &cobra.Command{
	Use:   "recent [n] [path]",
	Short: "Show the recently modified files or directories",
	Long:  `Display the most recently modified files or directories by modifictaion time.`,
	Args:  cobra.MaximumNArgs(2), // n and path are both optional
	Run:   runRecent,
}

func init() {
	recentCmd.Flags().BoolVar(&recentFiles, "files", false, "Show only files")
	recentCmd.Flags().BoolVar(&recentDirs, "dirs", false, "Show only DIrectories")

	rootCmd.AddCommand(recentCmd)
}

func runRecent(cmd *cobra.Command, args []string) {
	n := 10
	path := "."

	if len(args) >= 1 {
		stringN, err := strconv.Atoi(args[0])
		if err == nil {
			n = stringN
		} else {
			fmt.Println("Please enter a valid number")
			return
		}
	}
	if len(args) == 2 {
		path = args[1]
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

	if largestFiles {
		results = scan.GetNRecentlyModFiles(n)
	} else if largestDirs {
		results = scan.GetNRecentlyModDirs(n)
	} else {
		results = scan.Results
	}

	for _, elem := range results {
		fmt.Println(elem)
	}
}
