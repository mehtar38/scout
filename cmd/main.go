package main

import (
	"fmt"
	"os"
	"scout/internal/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
}
