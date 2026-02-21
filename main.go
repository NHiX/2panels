package main

import (
	"2panels/ui"
	"fmt"
	"os"
)

func main() {
	app := ui.NewApp()
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
		os.Exit(1)
	}
}
