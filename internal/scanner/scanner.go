package scanner

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
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

	// scan, err := scanner.New(path)
	// if err != nil {
	// 	fmt.Println("Error:", err)
	// 	return
	// }

	// err = scan.Scan()
	// if err != nil {
	// 	fmt.Println("Error scanning:", err)
	// 	return
	// }

	s.ComputeDirSize()

	// 1. Total files vs directories
	files := s.GetFiles()
	dirs := s.GetDirectories()
	total := len(files) + len(dirs)

	fmt.Println("Total items: ", total)
	fmt.Println("Files: ", len(files))
	fmt.Println("Directories: ", len(dirs))

	// 2. Total size
	var fileSize int
	var dirSize int

	for _, results := range s.Results {
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
	largestFile := s.GetNLargestFilesBySize(1)
	fmt.Println("Largest File: ", largestFile[0].Name, " ", largestFile[0].Size, " bytes")

	smallestFiles := s.GetNSmallestFilesBySize(1)
	fmt.Println("Smallest File: ", smallestFiles[0].Name, " ", smallestFiles[0].Size, " bytes")

	// 4. File type counts (count by Extension)
	fmt.Println("File Types: ")
	pdfs := s.GetFilesByExtension(".pdf")
	pngs := s.GetFilesByExtension(".png")
	jpgs := s.GetFilesByExtension(".jpg")
	docx := s.GetFilesByExtension(".docx")
	txt := s.GetFilesByExtension(".txt")
	fmt.Println(".pdf: ", len(pdfs), "files")
	fmt.Println(".docx: ", len(docx), "files")
	fmt.Println(".jpg: ", len(jpgs), "files")
	fmt.Println(".png: ", len(pngs), "files")
	fmt.Println(".txt ", len(txt), "files")

	// 5. Most recent/oldest file
	lastMod := s.GetNRecentlyModFiles(1)
	fmt.Println("Last Modified File: ", lastMod[0].Name)

	oldMod := s.GetNLeastModFiles(1)
	fmt.Println("Oldest Modified File: ", oldMod[0].Name)

}
