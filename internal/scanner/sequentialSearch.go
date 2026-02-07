package scanner

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type SearchResult struct {
	Name        string
	Location    string
	Matches     int
	LineNumbers []int
}

type SearchOptions struct {
	Pattern       string
	Extensions    []string
	UseRegex      bool
	CaseSensitive bool
}

func (s *Scanner) Search(opts SearchOptions) []SearchResult {
	files := s.GetFiles()
	results := []SearchResult{}

	var compiledRegex *regexp.Regexp
	if opts.UseRegex {
		var err error
		compiledRegex, err = regexp.Compile(opts.Pattern)
		if err != nil {
			return results
		}
	}

	for _, file := range files {
		if len(opts.Extensions) > 0 {
			allowed := false
			for _, ext := range opts.Extensions {
				if file.Extension == ext {
					allowed = true
				}
			}
			if !allowed {
				continue
			}
		}

		result := searchInFile(file, opts, compiledRegex)

		if result.Matches > 0 {
			results = append(results, result)
		}
	}

	return results
}

func searchInFile(file Metadata, opts SearchOptions, compiledRegex *regexp.Regexp) SearchResult {
	result := SearchResult{
		Name:        file.Name,
		Location:    file.Location,
		Matches:     0,
		LineNumbers: []int{},
	}

	f, err := os.Open(file.Location)
	if err != nil {
		fmt.Println("Error opening the file: ", err)
		return result //so that f is not nil and defer doesn't break
	}
	defer f.Close() //function runs no matter what when the function returns

	scanner := bufio.NewScanner(f)

	const maxCapacity = 1024 * 1024 // 1MB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if matchesPattern(line, opts, compiledRegex) {
			result.Matches++
			result.LineNumbers = append(result.LineNumbers, lineNum)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error: ", err)
	}

	return result
}

func matchesPattern(line string, opts SearchOptions, compiledRegex *regexp.Regexp) bool {
	text := line
	pattern := opts.Pattern

	if !opts.CaseSensitive {
		text = strings.ToLower(text)
		pattern = strings.ToLower(pattern)
	}

	if opts.UseRegex && compiledRegex != nil {
		return compiledRegex.MatchString(text)
	}

	return strings.Contains(text, pattern)
}
