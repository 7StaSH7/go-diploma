package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m tuiModel) View() string {
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	descriptionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	okStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	panelStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))

	var b strings.Builder
	authState := "НЕ АВТОРИЗОВАН"
	if m.authorized {
		authState = "АВТОРИЗОВАН"
	}
	b.WriteString(panelStyle.Render("Сессия: " + authState))
	b.WriteString("\n\n")

	if m.mode == tuiModeMenu {
		b.WriteString(m.renderMenu(panelStyle, mutedStyle, descriptionStyle, hintStyle))
	} else if m.mode == tuiModeSelect {
		b.WriteString(m.renderSecretSelection(panelStyle, mutedStyle, descriptionStyle, hintStyle))
	} else if m.mode == tuiModeConfirmDelete {
		b.WriteString(m.renderDeleteConfirm(panelStyle, mutedStyle, descriptionStyle, hintStyle))
	} else {
		b.WriteString(m.renderForm(panelStyle, mutedStyle, descriptionStyle, hintStyle))
	}

	b.WriteString("\n")
	status := strings.TrimSpace(m.status)
	if status != "" {
		switch {
		case strings.HasPrefix(status, "[ERR]"):
			b.WriteString(errStyle.Render("Статус: " + status))
		case strings.HasPrefix(status, "[OK]"):
			b.WriteString(okStyle.Render("Статус: " + status))
		default:
			b.WriteString(mutedStyle.Render("Статус: " + status))
		}
		b.WriteString("\n")
	}

	if out := strings.TrimSpace(m.output); out != "" {
		b.WriteString("\n")
		b.WriteString(headerStyle.Render("Последний ответ"))
		b.WriteString("\n")
		b.WriteString(panelStyle.Render(strings.TrimRight(m.output, "\n")))
		b.WriteString("\n")
	}

	return b.String()
}

func (m tuiModel) renderMenu(panelStyle, mutedStyle, descriptionStyle, hintStyle lipgloss.Style) string {
	var b strings.Builder
	actions := m.visibleActions()
	b.WriteString(mutedStyle.Render("Меню действий"))
	b.WriteString("\n")
	for i, action := range actions {
		prefix := "  "
		if i == m.cursor {
			prefix = "> "
		}
		num := strconv.Itoa(i + 1)
		line := fmt.Sprintf("%s%s. %s", prefix, num, action.Title)
		desc := action.Description
		if i == m.cursor {
			line = lipgloss.NewStyle().Bold(true).Render(line)
		}
		b.WriteString(line)
		b.WriteString("\n")
		b.WriteString("   " + descriptionStyle.Render(desc))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	if !m.authorized {
		b.WriteString(hintStyle.Render("Чтобы увидеть остальные команды, сначала выполните вход или регистрацию"))
		b.WriteString("\n")
	}
	b.WriteString(hintStyle.Render("Совет: нажмите цифру 1-9 для быстрого выбора пункта"))
	b.WriteString("\n")
	b.WriteString(panelStyle.Render("Клавиши: Enter открыть | Up/Down перемещение | 1-9 быстрый выбор | q выход | Ctrl+C принудительный выход"))
	return b.String()
}

func (m tuiModel) renderForm(panelStyle, mutedStyle, descriptionStyle, hintStyle lipgloss.Style) string {
	var b strings.Builder
	total := len(m.currentAction.Fields)
	current := m.fieldIndex + 1
	b.WriteString(descriptionStyle.Render(m.currentAction.Description))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render(fmt.Sprintf("Шаг %d/%d", current, total)))
	b.WriteString("\n\n")
	for i, f := range m.currentAction.Fields {
		value := m.fieldValues[f.Key]
		if i == m.fieldIndex {
			value = m.input
		}
		if f.Secret {
			value = strings.Repeat("*", len(value))
		}
		req := "необязательно"
		if f.Required {
			req = "обязательно"
		}
		line := fmt.Sprintf("  %s (%s): %s", f.Label, req, value)
		if i == m.fieldIndex {
			line = "> " + strings.TrimSpace(line)
			line = lipgloss.NewStyle().Bold(true).Render(line)
		}
		b.WriteString(line)
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(panelStyle.Render("Клавиши: Enter/Tab далее | Ctrl+G выполнить | Shift+Tab назад | Ctrl+U очистить | Esc отмена"))
	b.WriteString("\n")
	currentField := m.currentAction.Fields[m.fieldIndex]
	if currentField.Hint != "" {
		b.WriteString(hintStyle.Render("Подсказка: " + currentField.Hint))
		b.WriteString("\n")
	}
	b.WriteString(hintStyle.Render("Можно нажать Ctrl+G на любом шаге, чтобы сразу выполнить команду"))
	return b.String()
}

func (m tuiModel) renderDeleteConfirm(panelStyle, mutedStyle, descriptionStyle, hintStyle lipgloss.Style) string {
	var b strings.Builder
	b.WriteString(descriptionStyle.Render("Подтверждение удаления"))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("Проверьте секрет перед удалением. Это действие необратимо."))
	b.WriteString("\n\n")
	b.WriteString(panelStyle.Render(
		"Секрет: " + secretDisplayTitle(m.selectedSecret) + "\n" +
			"ID: " + fallbackText(m.selectedSecret.ID) + "\n" +
			"Теги: " + fallbackText(strings.Join(m.selectedSecret.MetaOpen.Tags, ", ")),
	))
	b.WriteString("\n")
	b.WriteString(hintStyle.Render("Нажмите Enter или Y, чтобы удалить | N или Esc, чтобы отменить"))
	return b.String()
}

func (m tuiModel) renderSecretSelection(panelStyle, mutedStyle, descriptionStyle, hintStyle lipgloss.Style) string {
	var b strings.Builder
	b.WriteString(descriptionStyle.Render(selectionActionLabel(m.selectionAction)))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render(selectionFilterSummary(m.selectionFilters)))
	b.WriteString("\n\n")

	if len(m.selectionItems) == 0 {
		b.WriteString(panelStyle.Render("Подбираю подходящие секреты..."))
		return b.String()
	}

	m.ensureSelectionCursor()
	start := 0
	if m.selectionCursor > 5 {
		start = m.selectionCursor - 5
	}
	end := minInt(len(m.selectionItems), start+10)
	if end-start < 10 {
		start = maxInt(0, end-10)
	}

	for i := start; i < end; i++ {
		item := m.selectionItems[i]
		prefix := "  "
		if i == m.selectionCursor {
			prefix = "> "
		}
		line := fmt.Sprintf("%s%d. %s", prefix, i+1, secretDisplayTitle(item))
		if i == m.selectionCursor {
			line = lipgloss.NewStyle().Bold(true).Render(line)
		}
		b.WriteString(line)
		b.WriteString("\n")
		b.WriteString("   ")
		b.WriteString(descriptionStyle.Render(secretShortLine(item)))
		b.WriteString("\n")
	}

	if len(m.selectionItems) > end-start {
		b.WriteString("\n")
		b.WriteString(mutedStyle.Render(fmt.Sprintf("Показаны %d-%d из %d", start+1, end, len(m.selectionItems))))
	}

	selected := m.selectionItems[m.selectionCursor]
	b.WriteString("\n\n")
	b.WriteString(panelStyle.Render("Выбран: " + secretDisplayTitle(selected) + "\nДанные: " + decodeSecretData(selected.Ciphertext)))
	b.WriteString("\n")
	b.WriteString(hintStyle.Render("Клавиши: Enter выбрать | Up/Down перемещение | 1-9 быстрый выбор | Esc отмена"))
	return b.String()
}

func selectionActionLabel(actionID string) string {
	switch actionID {
	case "update":
		return "Выбор секрета для обновления"
	case "delete":
		return "Выбор секрета для удаления"
	default:
		return "Выбор секрета"
	}
}

func selectionFilterSummary(filters map[string]string) string {
	title := strings.TrimSpace(filters[fieldFindTitle])
	tags := strings.TrimSpace(filters[fieldFindTags])
	date := strings.TrimSpace(filters[fieldFindDate])
	parts := make([]string, 0, 3)
	if title != "" {
		parts = append(parts, "название: "+title)
	}
	if tags != "" {
		parts = append(parts, "теги: "+tags)
	}
	if date != "" {
		parts = append(parts, "с даты: "+date)
	}
	if len(parts) == 0 {
		return "Фильтры: без фильтрации"
	}
	return "Фильтры: " + strings.Join(parts, " | ")
}

func secretShortLine(item secretOutputItem) string {
	tags := "-"
	if len(item.MetaOpen.Tags) > 0 {
		tags = strings.Join(item.MetaOpen.Tags, ", ")
	}
	date := extractDate(item.UpdatedAt)
	if date == "" {
		date = "-"
	}
	return fmt.Sprintf("Теги: %s | Дата: %s", tags, date)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
