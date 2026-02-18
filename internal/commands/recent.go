package commands

import (
	"fmt"
	"strconv"

	"github.com/mehtar38/scout/internal/scanner"
	"github.com/mehtar38/scout/internal/utils"

	"github.com/spf13/cobra"
)

var (
	recentFiles bool
	recentDirs  bool
	recentDays  int
)

var recentCmd = &cobra.Command{
	Use:   "recent [n] [path] --days [n]",
	Short: "Show the recently modified files or directories",
	Long:  `Display the most recently modified files or directories by modifictaion time.`,
	Args:  cobra.MaximumNArgs(2), // n and path are both optional
	Run:   runRecent,
}

func init() {
	recentCmd.Flags().BoolVar(&recentFiles, "files", false, "Show only files")
	recentCmd.Flags().BoolVar(&recentDirs, "dirs", false, "Show only Directories")
	recentCmd.Flags().IntVar(&recentDays, "days", 0, "Show files modified within last N days")

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

	if recentDays > 0 {
		results = scan.GetFilesNewerThan(recentDays)
	} else {
		if recentFiles {
			results = scan.GetNRecentlyModFiles(n)
		} else if recentDirs {
			results = scan.GetNRecentlyModDirs(n)
		} else {
			results = scan.GetNRecentlyModItems(n)
		}
	}

	if recentDays > 0 && len(args) >= 1 && len(results) > n {
		results = results[:n]
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
