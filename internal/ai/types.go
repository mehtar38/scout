package ai

type CommandParser struct {
	Command  string
	Count    int
	Filters  map[string]string
	Response string
}
