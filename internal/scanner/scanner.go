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

func (s *Scanner) ComputeDirSize() {
	for i := range s.Results { //Using index here and not value because value would just be a modified copy and not the actual element
		if s.Results[i].IsDir {
			var SizeOfDir int64
			dirPath := s.Results[i].Location

			for _, entry := range s.Results {
				if !entry.IsDir && strings.HasPrefix(entry.Location, dirPath+"/") {
					SizeOfDir += entry.Size
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
	sortedFiles := make([]Metadata, len(s.Results))
	copy(sortedFiles, s.Results)

	sort.Slice(sortedFiles, func(i, j int) bool { return sortedFiles[i].Size < sortedFiles[j].Size })
	return sortedFiles
}

func (s *Scanner) GetNLargestFilesBySize(n int) []Metadata {
	sortedFiles := s.GetAllFilesBySize()

	if n > len(sortedFiles) {
		n = len(sortedFiles)
	}

	return sortedFiles[:n]
}

func (s *Scanner) GetAllDirectoriesBySize() []Metadata {
	dires := s.GetDirectories()

	sort.Slice(dires, func(i, j int) bool { return dires[i].DirSize < dires[j].DirSize })
	return dires
}

func (s *Scanner) GetNLargestDirectoriesBySize(n int) []Metadata {
	sortedDires := s.GetAllDirectoriesBySize()

	if n > len(sortedDires) {
		n = len(sortedDires)
	}

	return sortedDires[:n]
}

func (s *Scanner) SortByModTimeDesc() []Metadata {
	sortedFiles := make([]Metadata, len(s.Results))
	copy(sortedFiles, s.Results)

	sort.Slice(sortedFiles, func(i, j int) bool { return sortedFiles[i].ModificationTime.After(sortedFiles[j].ModificationTime) })
	return sortedFiles
}

func (s *Scanner) SortByModTimeAsc() []Metadata {
	sortedFiles := make([]Metadata, len(s.Results))
	copy(sortedFiles, s.Results)

	sort.Slice(sortedFiles, func(i, j int) bool { return sortedFiles[i].ModificationTime.Before(sortedFiles[j].ModificationTime) })
	return sortedFiles
}

func (s *Scanner) GetNRecentlyModFiles(n int) []Metadata {
	files := s.SortByModTimeDesc()
	if n > len(files) {
		n = len(files)
	}

	return files[:n]
}

func (s *Scanner) GetNLeastModFiles(n int) []Metadata {
	files := s.SortByModTimeAsc()
	if n > len(files) {
		n = len(files)
	}

	return files[:n]
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
