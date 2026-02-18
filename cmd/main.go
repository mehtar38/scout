package main

import (
	"fmt"
	"os"

	"github.com/mehtar38/scout/internal/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
}
