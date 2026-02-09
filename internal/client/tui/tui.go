package tui

import (
	"fmt"
	"io"

	"github.com/7StaSH7/practicum-diploma/internal/version"
	tea "github.com/charmbracelet/bubbletea"
)

type TUIApp struct {
	stdout io.Writer
	stderr io.Writer
}

func NewTUI(stdout, stderr io.Writer) *TUIApp {
	return &TUIApp{
		stdout: stdout,
		stderr: stderr,
	}
}

func (a *TUIApp) Run(args []string) int {
	if len(args) > 0 {
		switch args[0] {
		case "version":
			_, _ = fmt.Fprint(a.stdout, version.Info())
			return 0
		case "help", "-h", "--help":
			_, _ = fmt.Fprintln(a.stdout, "PKeeper TUI")
			_, _ = fmt.Fprintln(a.stdout, "Запустите без аргументов для интерактивного режима.")
			return 0
		default:
			_, _ = fmt.Fprintf(a.stderr, "неизвестный аргумент: %s\n", args[0])
			return 2
		}
	}

	p := tea.NewProgram(newTUIModel(), tea.WithOutput(a.stdout))
	if _, err := p.Run(); err != nil {
		_, _ = fmt.Fprintf(a.stderr, "ошибка tui: %v\n", err)
		return 1
	}
	return 0
}
