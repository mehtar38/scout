package main

import (
	"fmt"
	"scout/internal/scanner"
)

func main() {
	scan, err := scanner.New(`.`)
	if err != nil {
		fmt.Println("Erroe initializing Scanner: ", err)
	}

	if err := scan.Scan(); err != nil {
		fmt.Print("Error Scanning: ", err)
	}

	for _, results := range scan.Results {
		fmt.Println("Name: ", results.Name)
		// fmt.Println("Mod Time: ", results.ModificationTime)
		// fmt.Println("Extension: ", results.Extension)
		fmt.Println("Size: ", results.Size)
	}
}
