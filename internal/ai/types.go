package ai

type CommandParser struct {
	Command  string
	Count    int
	Filters  map[string]string
	Response string
	Path     string
	Pattern  string
	Flags    map[string]bool
	Chain    []string
}
