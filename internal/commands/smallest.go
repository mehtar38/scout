package commands

import (
	"fmt"
	"scout/internal/scanner"
	"strconv"

	"github.com/spf13/cobra"
)

var (
	smallestFiles bool
	smallestDirs  bool
)

var smallestCmd = &cobra.Command{
	Use:   "smallest [n] [path]",
	Short: "Show the smallest files or directories",
	Long:  `Display the top N smallest files or directories by size.`,
	Args:  cobra.MaximumNArgs(2), // n and path are both optional
	Run:   runSmallest,
}

func init() {
	smallestCmd.Flags().BoolVar(&smallestFiles, "files", false, "Show only files")
	smallestCmd.Flags().BoolVar(&smallestDirs, "dirs", false, "Show only DIrectories")

	rootCmd.AddCommand(smallestCmd)
}

func runSmallest(cmd *cobra.Command, args []string) {
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

	if smallestFiles {
		results = scan.GetNSmallestFilesBySize(n)
	} else if smallestDirs {
		results = scan.GetNSmallestDirectoriesBySize(n)
	} else {
		results = scan.Results
	}

	for _, elem := range results {
		fmt.Println(elem)
	}
}
