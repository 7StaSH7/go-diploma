package main

import (
	"os"

	"github.com/7StaSH7/practicum-diploma/internal/client/tui"
)

func main() {
	app := tui.NewTUI(os.Stdout, os.Stderr)
	os.Exit(app.Run(os.Args[1:]))
}
