package main

import (
	"os"

	"github.com/7StaSH7/practicum-diploma/internal/client/cli"
)

func main() {
	app := cli.New(os.Stdout, os.Stderr)
	os.Exit(app.Run(os.Args[1:]))
}
