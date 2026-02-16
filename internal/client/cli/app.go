package cli

import (
	"fmt"
	"io"

	"github.com/7StaSH7/practicum-diploma/internal/version"
)

type App struct {
	stdout io.Writer
	stderr io.Writer
}

func New(stdout, stderr io.Writer) *App {
	return &App{
		stdout: stdout,
		stderr: stderr,
	}
}

func (a *App) Run(args []string) int {
	return Execute(args, a.stdout, a.stderr)
}

func Execute(args []string, stdout, stderr io.Writer) int {
	return run(args, stdout, stderr)
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printHelp(stdout)
		return 0
	}

	var err error
	switch args[0] {
	case "version":
		_, err = fmt.Fprint(stdout, version.Info())
	case "signup":
		err = runSignup(args[1:], stdout)
	case "signin":
		err = runSignin(args[1:], stdout)
	case "refresh":
		err = runRefresh(args[1:], stdout)
	case "secrets":
		err = runSecrets(args[1:], stdout)
	case "help", "-h", "--help":
		printHelp(stdout)
	default:
		_, _ = fmt.Fprintf(stderr, "unknown command: %s\n", args[0])
		printHelp(stderr)
		return 2
	}

	if err != nil {
		_, _ = fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	return 0
}

func printHelp(w io.Writer) {
	_, _ = fmt.Fprintln(w, "PKeeper CLI")
	_, _ = fmt.Fprintln(w, "Commands:")
	_, _ = fmt.Fprintln(w, "  version")
	_, _ = fmt.Fprintln(w, "  signup [--server URL] --login LOGIN --password PASSWORD")
	_, _ = fmt.Fprintln(w, "  signin [--server URL] --login LOGIN --password PASSWORD")
	_, _ = fmt.Fprintln(w, "  refresh [--server URL]")
	_, _ = fmt.Fprintln(w, "  secrets list [--server URL] [--since RFC3339]")
	_, _ = fmt.Fprintln(w, "  secrets sync [--server URL] [--since RFC3339] [--once]")
	_, _ = fmt.Fprintln(w, "  secrets get [--server URL] --id UUID")
	_, _ = fmt.Fprintln(w, "  secrets create [--server URL] --type TYPE --ciphertext BASE64 [--title TEXT] [--tags a,b] [--site URL]")
	_, _ = fmt.Fprintln(w, "  secrets update [--server URL] --id UUID --type TYPE --ciphertext BASE64 [--title TEXT] [--tags a,b] [--site URL]")
	_, _ = fmt.Fprintln(w, "  secrets delete [--server URL] --id UUID")
}
