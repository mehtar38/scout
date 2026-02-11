package commands

import (
	"fmt"
	"scout/internal/scanner"
	"strconv"

	"github.com/spf13/cobra"
)

var (
	largestFiles bool
	largestDirs  bool
)

var largestCmd = &cobra.Command{
	Use:   "largest [n] [path]",
	Short: "Show the largest files or directories",
	Long:  `Display the top N largest files or directories by size.`,
	Args:  cobra.MaximumNArgs(2), // n and path are both optional
	Run:   runLargest,
}

func init() {
	largestCmd.Flags().BoolVar(&largestFiles, "files", false, "Show only files")
	largestCmd.Flags().BoolVar(&largestDirs, "dirs", false, "Show only Directories")

	rootCmd.AddCommand(largestCmd)
}

func runLargest(cmd *cobra.Command, args []string) {
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
		results = scan.GetNLargestFilesBySize(n)
	} else if largestDirs {
		results = scan.GetNLargestDirectoriesBySize(n)
	} else {
		results = scan.Results
	}

	for _, elem := range results {
		fmt.Println(elem)
	}
}
