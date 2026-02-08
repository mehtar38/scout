package scanner

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
)

type SearchResult struct {
	Name        string
	Location    string
	Matches     int
	LineNumbers []int
}

// struct in case we want to add options without breaking the pattern later (4+ parameters = messy)
type SearchOptions struct {
	Pattern       string
	Extensions    []string
	UseRegex      bool
	CaseSensitive bool
}

func (s *Scanner) Search(opts SearchOptions) []SearchResult {
	files := s.GetFiles()

	jobs := make(chan Metadata, 100)
	results := make(chan SearchResult, 100)
	var waitGroup sync.WaitGroup
	workerPool := 30

	// compile regex
	var regex *regexp.Regexp
	if opts.UseRegex {
		var err error
		regex, err = regexp.Compile(opts.Pattern)
		if err != nil {
			return []SearchResult{}
		}
	}

	for range workerPool {
		waitGroup.Add(1)

		go func() {
			defer waitGroup.Done() //attaches to the goroutines, runs after the goroutine exits (no task tracking but once the channel is empty and routine exits)

			for file := range jobs { // no index cause its a channel not a slice
				results <- searchInFile(file, opts, regex)
			}
		}()
	}

	go func() {
		for _, file := range files {
			if len(opts.Extensions) > 0 {
				allowed := false
				for _, ext := range opts.Extensions {
					if file.Extension == ext {
						allowed = true
						break
					}
				}
				if !allowed {
					continue
				}
			}
			jobs <- file
		}
		close(jobs)
	}()

	go func() {
		waitGroup.Wait()
		close(results)
	}()

	var finalResults []SearchResult
	for res := range results {
		if res.Matches > 0 {
			finalResults = append(finalResults, res)
		}
	}

	return finalResults
}

func searchInFile(file Metadata, opts SearchOptions, regex *regexp.Regexp) SearchResult {
	result := SearchResult{
		Name:        file.Name,
		Location:    file.Location,
		Matches:     0,
		LineNumbers: []int{},
	}

	f, err := os.Open(file.Location)
	if err != nil {
		// fmt.Println("Error opening the file: ", err)
		return result //so that f is not nil and defer doesn't break
	}
	defer f.Close() //function runs no matter what when the function(searchInFile in this case) returns

	scanner := bufio.NewScanner(f)

	const maxCapacity = 1024 * 1024 // 1MB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if matchesPattern(line, opts, regex) {
			result.Matches++
			result.LineNumbers = append(result.LineNumbers, lineNum)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error: ", err)
	}

	return result
}

func matchesPattern(line string, opts SearchOptions, regex *regexp.Regexp) bool {
	text := line
	pattern := opts.Pattern

	if !opts.CaseSensitive {
		text = strings.ToLower(text)
		pattern = strings.ToLower(pattern)
	}

	if opts.UseRegex && regex != nil {
		return regex.MatchString(text)
	}

	return strings.Contains(text, pattern)
}
