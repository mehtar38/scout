package commands

import (
	"fmt"
	"scout/internal/scanner"
	"strconv"

	"github.com/spf13/cobra"
)

var (
	oldestFiles bool
	oldestDirs  bool
)

var oldestCmd = &cobra.Command{
	Use:   "oldest [n] [path]",
	Short: "Show the oldest files or directories",
	Long:  `Display the files or directories that haven't been modified in a while by modifictaion time.`,
	Args:  cobra.MaximumNArgs(2), // n and path are both optional
	Run:   runOldest,
}

func init() {
	oldestCmd.Flags().BoolVar(&oldestFiles, "files", false, "Show only files")
	oldestCmd.Flags().BoolVar(&oldestDirs, "dirs", false, "Show only DIrectories")

	rootCmd.AddCommand(oldestCmd)
}

func runOldest(cmd *cobra.Command, args []string) {
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
		results = scan.GetNLeastModFiles(n)
	} else if largestDirs {
		results = scan.GetNLeastModDirs(n)
	} else {
		results = scan.Results
	}

	for _, elem := range results {
		fmt.Println(elem)
	}
}
