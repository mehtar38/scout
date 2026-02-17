package ai

type CommandParser struct {
	Command  string
	Count    int
	Days     int
	Filters  map[string]string
	Response string
	Path     string
	Pattern  string
	Flags    map[string]bool
	Chain    []string
}
