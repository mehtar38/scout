package scanner

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"scout/internal/utils"
	"sort"
	"strings"
	"time"
)

type Metadata struct {
	Name             string
	Size             int64 //OS wal func returned file size
	DirSize          int64 //For the aggregate directory size, the OS Walk func initially gave 0 as dir size
	IsDir            bool
	ModificationTime time.Time
	Location         string
	Extension        string
}

type Scanner struct {
	Results []Metadata
	Errors  []error
	path    string
}

func New(path string) (*Scanner, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("path not found: %w", err) // %w cause  it wraps the whole error unlike %v
	}
	return &Scanner{
		Results: make([]Metadata, 0),
		Errors:  make([]error, 0),
		path:    path,
	}, nil
}

func (s *Scanner) Scan() error {
	return filepath.WalkDir(s.path, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			switch d.Name() {
			case "node_modules", "vendor", "build", "dist":
				return filepath.SkipDir
			}
		}
		if err != nil {
			s.Errors = append(s.Errors, err)
			return nil // Continue walking the tree cause Walk stops the whole tree if it encounters any errors
		}

		info, err := d.Info()
		if err != nil {
			s.Errors = append(s.Errors, err)
			return nil
		}

		entry := Metadata{
			ModificationTime: info.ModTime(),
			Name:             info.Name(),
			Size:             info.Size(),
			DirSize:          info.Size(),
			IsDir:            info.IsDir(),
			Location:         path,
			Extension:        filepath.Ext(info.Name()),
		}

		s.Results = append(s.Results, entry)

		return nil
	})
}

// To normalize all string backslashes to forward slashes for
func normalize(p string) string {
	return filepath.ToSlash(p)
}

func (s *Scanner) ComputeDirSize() {
	for i := range s.Results { //Using index here and not value because value would just be a modified copy and not the actual element
		if s.Results[i].IsDir {
			var SizeOfDir int64
			dirPath := normalize(s.Results[i].Location)

			for _, entry := range s.Results {
				if !entry.IsDir {
					entryPath := normalize(entry.Location)

					if strings.HasPrefix(entryPath, dirPath+"/") {
						SizeOfDir += entry.Size
					}
				}
			}
			s.Results[i].DirSize = SizeOfDir
		}
	}
}

//Util functions - Get all Files and Directoreies (also by size)
// Windows onky feature - last accessed / created

func (s *Scanner) GetFiles() []Metadata {

	files := []Metadata{}

	for _, results := range s.Results {
		if !results.IsDir {
			files = append(files, results)
		}
	}
	return files
}

func (s *Scanner) GetDirectories() []Metadata {
	dires := []Metadata{}

	for _, results := range s.Results {
		if results.IsDir {
			dires = append(dires, results)
		}
	}
	return dires
}

func (s *Scanner) GetAllFilesBySize() []Metadata {
	sortedFiles := s.GetFiles()

	sort.Slice(sortedFiles, func(i, j int) bool { return sortedFiles[i].Size > sortedFiles[j].Size })
	return sortedFiles
}

func (s *Scanner) GetNLargestItems(n int) []Metadata {
	items := make([]Metadata, len(s.Results))
	copy(items, s.Results)

	sort.Slice(items, func(i, j int) bool {
		return items[i].Size > items[j].Size
	})

	if len(items) > n {
		return items[:n]
	}
	return items
}

func (s *Scanner) GetNSmallestItems(n int) []Metadata {
	items := make([]Metadata, len(s.Results))
	copy(items, s.Results)

	sort.Slice(items, func(i, j int) bool {
		return items[i].Size < items[j].Size
	})

	if len(items) > n {
		return items[:n]
	}
	return items
}

func (s *Scanner) GetNLargestFilesBySize(n int) []Metadata {
	sortedFiles := s.GetAllFilesBySize()

	if n > len(sortedFiles) {
		n = len(sortedFiles)
	}

	return sortedFiles[:n]
}

func (s *Scanner) GetNSmallestFilesBySize(n int) []Metadata {
	sortedFiles := s.GetFiles()

	sort.Slice(sortedFiles, func(i, j int) bool { return sortedFiles[i].Size < sortedFiles[j].Size })

	if n > len(sortedFiles) {
		n = len(sortedFiles)
	}

	return sortedFiles[:n]
}

func (s *Scanner) GetAllDirectoriesBySize() []Metadata {
	dires := s.GetDirectories()

	sort.Slice(dires, func(i, j int) bool { return dires[i].DirSize > dires[j].DirSize })
	return dires
}

func (s *Scanner) GetNLargestDirectoriesBySize(n int) []Metadata {
	sortedDires := s.GetAllDirectoriesBySize()

	if n > len(sortedDires) {
		n = len(sortedDires)
	}

	return sortedDires[:n]
}

func (s *Scanner) GetNSmallestDirectoriesBySize(n int) []Metadata {
	dires := s.GetDirectories()

	sort.Slice(dires, func(i, j int) bool { return dires[i].DirSize < dires[j].DirSize })

	if n > len(dires) {
		n = len(dires)
	}

	return dires[:n]
}

func (s *Scanner) GetNRecentlyModItems(n int) []Metadata {
	items := make([]Metadata, len(s.Results))
	copy(items, s.Results)

	sort.Slice(items, func(i, j int) bool {
		return items[i].ModificationTime.After(items[j].ModificationTime)
	})

	if len(items) > n {
		return items[:n]
	}
	return items
}

func (s *Scanner) GetNRecentlyModFiles(n int) []Metadata {
	files := s.GetFiles()

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModificationTime.After(files[j].ModificationTime)
	})

	if n > len(files) {
		n = len(files)
	}

	return files[:n]
}

func (s *Scanner) GetNRecentlyModDirs(n int) []Metadata {
	dirs := s.GetDirectories()

	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].ModificationTime.After(dirs[j].ModificationTime)
	})

	if n > len(dirs) {
		n = len(dirs)
	}

	return dirs[:n]
}

func (s *Scanner) GetNLeastModItems(n int) []Metadata {
	items := make([]Metadata, len(s.Results))
	copy(items, s.Results)

	sort.Slice(items, func(i, j int) bool {
		return items[i].ModificationTime.Before(items[j].ModificationTime)
	})

	if len(items) > n {
		return items[:n]
	}
	return items
}

func (s *Scanner) GetNLeastModFiles(n int) []Metadata {
	files := s.GetFiles()

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModificationTime.Before(files[j].ModificationTime)
	})

	if n > len(files) {
		n = len(files)
	}

	return files[:n]
}

func (s *Scanner) GetNLeastModDirs(n int) []Metadata {
	dirs := s.GetDirectories()

	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].ModificationTime.Before(dirs[j].ModificationTime)
	})

	if n > len(dirs) {
		n = len(dirs)
	}

	return dirs[:n]
}

func (s *Scanner) GetFilesOlderThan(days int) []Metadata {
	filterDate := time.Now().AddDate(0, 0, -days)
	files := []Metadata{}

	for _, file := range s.GetFiles() {
		if file.ModificationTime.Before(filterDate) {
			files = append(files, file)
		}
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModificationTime.Before(files[j].ModificationTime)
	})

	return files
}

func (s *Scanner) GetFilesNewerThan(days int) []Metadata {
	filterDate := time.Now().AddDate(0, 0, -days)
	files := []Metadata{}

	for _, file := range s.GetFiles() {
		if file.ModificationTime.After(filterDate) {
			files = append(files, file)
		}
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModificationTime.After(files[j].ModificationTime)
	})

	return files
}

func (s *Scanner) GetFilesByExtension(ext string) []Metadata {
	files := []Metadata{}

	for _, results := range s.Results {
		if results.Extension == ext {
			files = append(files, results)
		}
	}
	return files
}

func (s *Scanner) GetStats() {

	s.ComputeDirSize()

	files := s.GetFiles()
	dirs := s.GetDirectories()

	fmt.Println("=== System Statistics ===")
	fmt.Printf("Total Items: %d\n", len(files)+len(dirs))
	fmt.Printf("Files: %d\n", len(files))
	fmt.Printf("Directories: %d\n\n", len(dirs))

	// --- Size Stats ---
	var totalFileSize, totalDirSize int64
	for _, f := range files {
		totalFileSize += f.Size
	}
	for _, d := range dirs {
		totalDirSize += d.DirSize
	}

	fmt.Println("=== Size Statistics ===")
	fmt.Printf("Total Size: %s\n", utils.FormatSize(totalFileSize+totalDirSize))
	fmt.Printf("Files: %s\n", utils.FormatSize(totalFileSize))
	fmt.Printf("Directories: %s\n", utils.FormatSize(totalDirSize))

	if len(files) > 0 {
		fmt.Printf("Avg File Size: %s\n", utils.FormatSize(totalFileSize/int64(len(files))))
	}
	if len(dirs) > 0 {
		fmt.Printf("Avg Dir Size: %s\n", utils.FormatSize(totalDirSize/int64(len(dirs))))
	}
	fmt.Println()

	// --- Extremes ---
	if len(files) > 0 {
		largest := s.GetNLargestFilesBySize(1)[0]
		smallest := s.GetNSmallestFilesBySize(1)[0]

		fmt.Println("=== Extremes ===")
		fmt.Printf("Largest File:  %s (%s)\n", largest.Name, utils.FormatSize(largest.Size))
		fmt.Printf("Smallest File: %s (%s)\n\n", smallest.Name, utils.FormatSize(smallest.Size))
	}

	// --- Modification Times ---
	if len(files) > 0 {
		recent := s.GetNRecentlyModFiles(1)[0]
		oldest := s.GetNLeastModFiles(1)[0]

		fmt.Println("=== Modification Times ===")
		fmt.Printf("Most Recent: %s (%s)\n", recent.Name, recent.ModificationTime.Format("2006-01-02 15:04"))
		fmt.Printf("Oldest:      %s (%s)\n\n", oldest.Name, oldest.ModificationTime.Format("2006-01-02 15:04"))
	}

	// --- Extensions ---
	extCounts := make(map[string]int)
	for _, f := range files {
		if f.Extension != "" {
			extCounts[f.Extension]++
		}
	}

	if len(extCounts) > 0 {
		fmt.Println("=== File Extensions ===")

		type extCount struct {
			ext   string
			count int
		}

		var exts []extCount
		for ext, count := range extCounts {
			exts = append(exts, extCount{ext, count})
		}

		sort.Slice(exts, func(i, j int) bool {
			return exts[i].count > exts[j].count
		})

		for _, ec := range exts {
			fmt.Printf("%s: %d files\n", ec.ext, ec.count)
		}
	}
}
