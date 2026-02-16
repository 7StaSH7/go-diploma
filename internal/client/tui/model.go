package tui

import (
	"github.com/7StaSH7/practicum-diploma/internal/client/cli"
	tea "github.com/charmbracelet/bubbletea"
)

func newTUIModel() tuiModel {
	authorized := isAuthorizedSession()
	status := "Выберите действие и нажмите Enter"
	if !authorized {
		status = "Войдите или зарегистрируйтесь для продолжения"
	}

	return tuiModel{
		mode:             tuiModeMenu,
		cursor:           0,
		authorized:       authorized,
		autoSync:         authorized,
		fieldValues:      make(map[string]string),
		selectionFilters: make(map[string]string),
		status:           status,
	}
}

func isAuthorizedSession() bool {
	authorized, err := cli.AuthorizedSession()
	if err != nil {
		return false
	}
	return authorized
}

func (m tuiModel) Init() tea.Cmd {
	cmds := make([]tea.Cmd, 0, 2)
	if m.autoSync {
		cmds = append(cmds, syncTickCmd())
	}
	if m.authorized {
		cmds = append(cmds, runTUIActionCmd("search", map[string]string{}))
	}
	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case operationResultMsg:
		return m.handleOperationResult(msg)
	case secretSelectionLoadedMsg:
		return m.handleSelectionLoaded(msg)
	case syncTickMsg:
		return m.handleSyncTick()
	}
	return m, nil
}
