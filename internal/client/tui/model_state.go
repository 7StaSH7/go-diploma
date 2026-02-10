package tui

import (
	"errors"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func isAuthAction(actionID string) bool {
	return actionID == "signup" || actionID == "signin"
}

func requiresSecretSelection(actionID string) bool {
	return actionID == "update" || actionID == "delete"
}

func (m tuiModel) visibleActions() []tuiAction {
	actions := make([]tuiAction, 0, len(tuiActions))
	for _, action := range tuiActions {
		authAction := isAuthAction(action.ID)
		if !m.authorized && !authAction {
			continue
		}
		if m.authorized && authAction {
			continue
		}
		actions = append(actions, action)
	}
	return actions
}

func (m *tuiModel) ensureCursor(actions []tuiAction) {
	if len(actions) == 0 {
		m.cursor = 0
		return
	}
	if m.cursor < 0 || m.cursor >= len(actions) {
		m.cursor = 0
	}
}

func (m *tuiModel) refreshAuthorizationState() {
	m.authorized = isAuthorizedSession()
	m.autoSync = m.authorized
	if !m.authorized {
		m.syncInFlight = false
	}
	m.ensureCursor(m.visibleActions())
	m.ensureSelectionCursor()
}

func (m *tuiModel) ensureSelectionCursor() {
	if len(m.selectionItems) == 0 {
		m.selectionCursor = 0
		return
	}
	if m.selectionCursor < 0 || m.selectionCursor >= len(m.selectionItems) {
		m.selectionCursor = 0
	}
}

func (m *tuiModel) clearSelectionState() {
	m.selectionItems = nil
	m.selectionCursor = 0
	m.selectionAction = ""
	m.selectionFilters = nil
}

func (m *tuiModel) clearFormState() {
	m.currentAction = tuiAction{}
	m.fieldIndex = 0
	m.fieldValues = make(map[string]string)
	m.input = ""
}

func (m tuiModel) startSecretSelection(actionID string, filters map[string]string) (tea.Model, tea.Cmd) {
	m.mode = tuiModeSelect
	m.selectionAction = actionID
	m.selectionFilters = cloneStringMap(filters)
	m.selectionItems = nil
	m.selectionCursor = 0
	m.selectedSecret = secretOutputItem{}
	m.status = "[INFO] Загружаю список секретов..."
	return m, loadSecretSelectionCmd(filters)
}

func (m tuiModel) applySelectedSecret(item secretOutputItem) (tea.Model, tea.Cmd) {
	actionID := m.selectionAction
	m.selectedSecret = item
	m.clearSelectionState()

	switch actionID {
	case "update":
		m.mode = tuiModeForm
		m.currentAction = updateSelectedAction
		m.fieldIndex = 0
		m.fieldValues = map[string]string{
			"title": item.MetaOpen.Title,
			"tags":  strings.Join(item.MetaOpen.Tags, ","),
			"site":  item.MetaOpen.Site,
		}
		m.input = ""
		m.status = "[INFO] Выбран секрет: " + secretDisplayTitle(item)
		return m, nil
	case "delete":
		m.mode = tuiModeConfirmDelete
		m.clearFormState()
		m.status = "[INFO] Подтвердите удаление выбранного секрета"
		return m, nil
	default:
		m.mode = tuiModeMenu
		m.status = "[ERR] Неизвестный режим выбора секрета"
		return m, nil
	}
}

func (m *tuiModel) startAction(action tuiAction) tea.Cmd {
	if !m.authorized && !isAuthAction(action.ID) {
		m.status = "[INFO] Сначала войдите или зарегистрируйтесь"
		return nil
	}
	if m.authorized && isAuthAction(action.ID) {
		m.status = "[INFO] Вы уже авторизованы"
		return nil
	}

	if len(action.Fields) == 0 {
		m.status = "[INFO] Выполняю команду..."
		return runTUIActionCmd(action.ID, nil)
	}
	m.mode = tuiModeForm
	m.currentAction = action
	m.fieldIndex = 0
	m.fieldValues = make(map[string]string)
	m.input = ""
	m.status = "[INFO] Заполните форму (Enter - далее, Ctrl+G - выполнить, Esc - отмена)"
	return nil
}

func validateField(field tuiField, value string) error {
	if field.Required && value == "" {
		return fmt.Errorf("поле обязательно: %s", field.Label)
	}
	switch field.Key {
	case "since":
		if value != "" {
			if _, err := time.Parse(time.RFC3339, value); err != nil {
				return errors.New("время должно быть в RFC3339, например 2026-02-09T10:00:00Z")
			}
		}
	case fieldFindDate:
		if value != "" {
			if _, err := time.Parse("2006-01-02", value); err != nil {
				return errors.New("дата должна быть в формате ГГГГ-ММ-ДД, например 2026-02-09")
			}
		}
	}
	return nil
}

func (m tuiModel) submitCurrentAction() (tea.Model, tea.Cmd) {
	values := make(map[string]string, len(m.fieldValues))
	for k, v := range m.fieldValues {
		values[k] = v
	}
	actionID := m.currentAction.ID

	if requiresSecretSelection(actionID) {
		return m.startSecretSelection(actionID, values)
	}

	if actionID == actionUpdateSelected {
		values["id"] = m.selectedSecret.ID
		actionID = "update"
	}

	m.mode = tuiModeMenu
	m.clearFormState()
	m.clearSelectionState()
	m.status = "[INFO] Выполняю команду..."
	return m, runTUIActionCmd(actionID, values)
}
