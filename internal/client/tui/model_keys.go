package tui

import (
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m tuiModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}

	switch m.mode {
	case tuiModeMenu:
		return m.handleMenuKey(msg)
	case tuiModeSelect:
		return m.handleSelectKey(msg)
	case tuiModeConfirmDelete:
		return m.handleDeleteConfirmKey(msg)
	default:
		return m.handleFormKey(msg)
	}
}

func (m tuiModel) handleMenuKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	actions := m.visibleActions()
	m.ensureCursor(actions)

	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		idx, _ := strconv.Atoi(msg.String())
		if idx >= 1 && idx <= len(actions) {
			m.cursor = idx - 1
			m.status = "[INFO] Выбрано: " + actions[m.cursor].Title + " (Enter для запуска)"
		}
		return m, nil
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil
	case "down", "j":
		if m.cursor < len(actions)-1 {
			m.cursor++
		}
		return m, nil
	case "enter":
		if len(actions) == 0 {
			return m, nil
		}
		action := actions[m.cursor]
		cmd := m.startAction(action)
		return m, cmd
	}

	return m, nil
}

func (m tuiModel) handleSelectKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.ensureSelectionCursor()

	switch msg.String() {
	case "esc":
		m.mode = tuiModeMenu
		m.clearSelectionState()
		m.status = "[INFO] Выбор секрета отменен"
		return m, nil
	case "up", "k":
		if m.selectionCursor > 0 {
			m.selectionCursor--
		}
		return m, nil
	case "down", "j":
		if m.selectionCursor < len(m.selectionItems)-1 {
			m.selectionCursor++
		}
		return m, nil
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		idx, _ := strconv.Atoi(msg.String())
		if idx >= 1 && idx <= len(m.selectionItems) {
			m.selectionCursor = idx - 1
			m.status = "[INFO] Выбран вариант №" + strconv.Itoa(idx) + " (Enter для подтверждения)"
		}
		return m, nil
	case "enter":
		if len(m.selectionItems) == 0 {
			return m, nil
		}
		return m.applySelectedSecret(m.selectionItems[m.selectionCursor])
	}

	return m, nil
}

func (m tuiModel) handleDeleteConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "n", "N":
		m.mode = tuiModeMenu
		m.selectedSecret = secretOutputItem{}
		m.status = "[INFO] Удаление отменено"
		return m, nil
	case "enter", "y", "Y":
		secretID := strings.TrimSpace(m.selectedSecret.ID)
		if secretID == "" {
			m.mode = tuiModeMenu
			m.status = "[ERR] Не выбран секрет для удаления"
			return m, nil
		}
		m.mode = tuiModeMenu
		m.clearFormState()
		m.clearSelectionState()
		m.status = "[INFO] Выполняю удаление..."
		return m, runTUIActionCmd("delete", map[string]string{"id": secretID})
	}

	return m, nil
}

func (m tuiModel) handleFormKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = tuiModeMenu
		m.clearFormState()
		m.clearSelectionState()
		m.status = "[INFO] Действие отменено"
		return m, nil
	case "up", "k", "shift+tab":
		if m.fieldIndex > 0 {
			currentField := m.currentAction.Fields[m.fieldIndex]
			m.fieldValues[currentField.Key] = strings.TrimSpace(m.input)
			m.fieldIndex--
			prev := m.currentAction.Fields[m.fieldIndex]
			m.input = m.fieldValues[prev.Key]
		}
		return m, nil
	case "down", "j", "tab":
		field := m.currentAction.Fields[m.fieldIndex]
		value := strings.TrimSpace(m.input)
		if err := validateField(field, value); err != nil {
			m.status = "[ERR] " + err.Error()
			return m, nil
		}
		m.fieldValues[field.Key] = value
		m.input = ""
		if m.fieldIndex < len(m.currentAction.Fields)-1 {
			m.fieldIndex++
			next := m.currentAction.Fields[m.fieldIndex]
			m.input = m.fieldValues[next.Key]
			return m, nil
		}
		return m.submitCurrentAction()
	case "ctrl+g":
		currentField := m.currentAction.Fields[m.fieldIndex]
		m.fieldValues[currentField.Key] = strings.TrimSpace(m.input)
		for _, field := range m.currentAction.Fields {
			value := strings.TrimSpace(m.fieldValues[field.Key])
			if err := validateField(field, value); err != nil {
				m.status = "[ERR] " + err.Error()
				return m, nil
			}
		}
		return m.submitCurrentAction()
	case "backspace":
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
		return m, nil
	case "ctrl+u":
		m.input = ""
		return m, nil
	case "enter":
		return m.handleFormKey(tea.KeyMsg{Type: tea.KeyTab})
	default:
		if len(msg.Runes) > 0 {
			m.input += string(msg.Runes)
		}
		return m, nil
	}
}
