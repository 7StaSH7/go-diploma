package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m tuiModel) handleOperationResult(msg operationResultMsg) (tea.Model, tea.Cmd) {
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
}

func (m tuiModel) handleSelectionLoaded(msg secretSelectionLoadedMsg) (tea.Model, tea.Cmd) {
	if msg.Err != nil {
		m.mode = tuiModeMenu
		m.clearSelectionState()
		m.status = "[ERR] " + msg.Err.Error()
		return m, nil
	}

	if len(msg.Items) == 0 {
		m.mode = tuiModeMenu
		m.clearSelectionState()
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
}

func (m tuiModel) handleSyncTick() (tea.Model, tea.Cmd) {
	if !m.autoSync {
		return m, nil
	}
	if m.syncInFlight {
		return m, syncTickCmd()
	}
	m.syncInFlight = true
	return m, tea.Batch(syncTickCmd(), runTUIActionCmd("auto_sync", map[string]string{}))
}
