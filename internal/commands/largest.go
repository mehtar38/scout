package commands

import (
	"fmt"
	"strconv"

	"github.com/mehtar38/scout/internal/scanner"
	"github.com/mehtar38/scout/internal/utils"

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
			fmt.Println("Printing 10 results by default: ")
			n = 10
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
		results = scan.GetNLargestItems(n)
	}

	if len(results) == 0 {
		fmt.Println("No results found")
		return
	}

	var lines []string
	for _, elem := range results {
		icon := "📄"
		if elem.IsDir {
			icon = "📁"
		}
		lines = append(lines, fmt.Sprintf("%s %-50s %12s  %s",
			icon,
			elem.Name,
			utils.FormatSize(elem.Size),
			elem.ModificationTime.Format("2006-01-02 15:04")))
	}

	utils.Paginate(lines, 20)
}
