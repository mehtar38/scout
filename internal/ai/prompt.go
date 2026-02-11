package ai

import "fmt"

func BuildPrompt(userQuery string) string {
	return fmt.Sprintf(`You are a command parser for Scout, a file system analysis tool.

Available commands:
- "largest": Show largest files by size
- "smallest": Show smallest files by size
- "recent": Show recently modified files
- "oldest": Show least recently modified files
- "search": Search inside files by name/pattern

Available filters (optional):
- "extension": File extensions like ".pdf", ".mp4" (comma-separated if multiple)
- "older_than_days": Files not modified in X days
- "newer_than_days": Files modified in last X days

User query: "%s"

Parse this into a JSON command. Use your best judgment for:
- Extracting the count (default to 10 if not specified)
- Identifying relevant filters
- Writing a brief, friendly explanation

Return ONLY valid JSON in this exact format:
{
  "command": "command_name",
  "count": number,
  "filters": {
    "extension": ".ext",
    "older_than_days": number
  },
  "explanation": "one sentence explaining what you're searching for"
}

If you cannot parse the query into a valid command, return:
{ "command": "none", "explanation": "unable to parse" }

Examples:

Query: "find my 5 biggest files"
Response:
{
  "command": "largest",
  "count": 5,
  "filters": {},
  "explanation": "Finding your 5 largest files"
}

Query: "show videos I haven't watched in forever"
Response:
{
  "command": "oldest",
  "count": 10,
  "filters": {"extension": ".mp4,.avi,.mkv"},
  "explanation": "Finding old video files you haven't opened recently"
}

Query: "what PDFs are eating my disk space"
Response:
{
  "command": "largest",
  "count": 10,
  "filters": {"extension": ".pdf"},
  "explanation": "Finding your largest PDF files"
}`, userQuery)
}
