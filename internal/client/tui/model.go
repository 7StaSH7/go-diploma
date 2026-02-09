package tui

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/7StaSH7/practicum-diploma/internal/client/cli"
	tea "github.com/charmbracelet/bubbletea"
)

func newTUIModel() tuiModel {
	authorized := cli.HasAuthorizedSession()
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
		m.syncInFlight = false
		previousAutoSync := m.autoSync
		isAutoSync := msg.ActionID == "auto_sync"
		if msg.Err != nil {
			if !isAutoSync {
				m.status = "[ERR] " + msg.Err.Error()
			}
		} else {
			if !isAutoSync && strings.TrimSpace(msg.Output) != "" {
				m.output = msg.Output
			}
			if !isAutoSync {
				m.status = "[OK] Готово"
			}
		}
		m.refreshAuthorizationState()
		if !previousAutoSync && m.autoSync {
			m.status = "[OK] Вход выполнен"
			return m, syncTickCmd()
		}
		return m, nil
	case secretSelectionLoadedMsg:
		if msg.Err != nil {
			m.mode = tuiModeMenu
			m.selectionItems = nil
			m.selectionCursor = 0
			m.status = "[ERR] " + msg.Err.Error()
			return m, nil
		}
		if len(msg.Items) == 0 {
			m.mode = tuiModeMenu
			m.selectionItems = nil
			m.selectionCursor = 0
			m.status = "[INFO] По заданным фильтрам ничего не найдено"
			return m, nil
		}
		m.selectionItems = msg.Items
		m.selectionCursor = 0
		m.status = fmt.Sprintf("[INFO] Найдено %d секрет(ов). Выберите нужный", len(msg.Items))
		if len(msg.Items) == 1 {
			return m.applySelectedSecret(msg.Items[0])
		}
		return m, nil
	case syncTickMsg:
		if !m.autoSync {
			return m, nil
		}
		if m.syncInFlight {
			return m, syncTickCmd()
		}
		m.syncInFlight = true
		return m, tea.Batch(syncTickCmd(), runTUIActionCmd("auto_sync", map[string]string{}))
	}
	return m, nil
}

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
	m.authorized = cli.HasAuthorizedSession()
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
	m.selectionItems = nil
	m.selectionCursor = 0
	m.selectionAction = ""
	m.selectionFilters = nil

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
		m.currentAction = tuiAction{}
		m.fieldIndex = 0
		m.fieldValues = map[string]string{}
		m.input = ""
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

func (m tuiModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	}

	if m.mode == tuiModeMenu {
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

	if m.mode == tuiModeSelect {
		m.ensureSelectionCursor()
		switch msg.String() {
		case "esc":
			m.mode = tuiModeMenu
			m.selectionItems = nil
			m.selectionCursor = 0
			m.selectionAction = ""
			m.selectionFilters = nil
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

	if m.mode == tuiModeConfirmDelete {
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
			m.currentAction = tuiAction{}
			m.fieldValues = make(map[string]string)
			m.input = ""
			m.selectionAction = ""
			m.selectionFilters = nil
			m.status = "[INFO] Выполняю удаление..."
			return m, runTUIActionCmd("delete", map[string]string{"id": secretID})
		}
		return m, nil
	}

	switch msg.String() {
	case "esc":
		m.mode = tuiModeMenu
		m.currentAction = tuiAction{}
		m.selectionAction = ""
		m.selectionFilters = nil
		m.fieldIndex = 0
		m.input = ""
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
		return m.handleKey(tea.KeyMsg{Type: tea.KeyTab})
	default:
		if len(msg.Runes) > 0 {
			m.input += string(msg.Runes)
		}
		return m, nil
	}
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
	m.currentAction = tuiAction{}
	m.fieldIndex = 0
	m.fieldValues = make(map[string]string)
	m.input = ""
	m.selectionAction = ""
	m.selectionFilters = nil
	m.status = "[INFO] Выполняю команду..."
	return m, runTUIActionCmd(actionID, values)
}
