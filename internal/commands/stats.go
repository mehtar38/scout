package commands

import (
	"fmt"
	"scout/internal/scanner"

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

	// 1. Total files vs directories (use GetFiles(), GetDirectories())
	files := scan.GetFiles()
	dirs := scan.GetDirectories()
	total := len(files) + len(dirs)

	fmt.Println("Total items: ", total)
	fmt.Println("Files: ", len(files))
	fmt.Println("Directories: ", len(dirs))

	// 2. Total size (sum all CalculatedSize)
	var fileSize int
	var dirSize int

	for _, results := range scan.Results {
		if results.IsDir {
			dirSize += int(results.DirSize)
		} else {
			fileSize += int(results.Size)
		}
	}

	fmt.Println("Total Size: ", (dirSize + fileSize))
	fmt.Println("Files: ", fileSize)
	fmt.Println("Directories: ", dirSize)
	fmt.Println("Average File Size: ", (fileSize / len(files)))
	fmt.Println("Average Directory Size: ", (dirSize / len(dirs)))

	// 3. Largest/smallest file
	largestFile := scan.GetNLargestFilesBySize(1)
	fmt.Println("Largest File: ", largestFile[0].Name, " ", largestFile[0].Size, " bytes")

	smallestFiles := scan.GetNSmallestFilesBySize(1)
	fmt.Println("Smallest File: ", smallestFiles[0].Name, " ", smallestFiles[0].Size, " bytes")

	// 4. File type counts (count by Extension)
	fmt.Println("File Types: ")
	pdfs := scan.GetFilesByExtension(".pdf")
	pngs := scan.GetFilesByExtension(".png")
	jpgs := scan.GetFilesByExtension(".jpg")
	docx := scan.GetFilesByExtension(".docx")
	txt := scan.GetFilesByExtension(".txt")
	fmt.Println(".pdf: ", len(pdfs), "files")
	fmt.Println(".docx: ", len(docx), "files")
	fmt.Println(".jpg: ", len(jpgs), "files")
	fmt.Println(".png: ", len(pngs), "files")
	fmt.Println(".txt ", len(txt), "files")

	// 5. Most recent/oldest file
	lastMod := scan.GetNRecentlyModFiles(1)
	fmt.Println("Last Modified File: ", lastMod[0].Name)

	oldMod := scan.GetNLeastModFiles(1)
	fmt.Println("Oldest Modified File: ", oldMod[0].Name)

	// TODO: Print nicely formatted output
}
