package tui

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type secretOutputMeta struct {
	Title string   `json:"title"`
	Tags  []string `json:"tags"`
	Site  string   `json:"site"`
}

type secretOutputItem struct {
	ID         string           `json:"id"`
	Type       string           `json:"type"`
	MetaOpen   secretOutputMeta `json:"meta_open"`
	Ciphertext string           `json:"ciphertext"`
	Version    int64            `json:"version"`
	UpdatedAt  string           `json:"updated_at"`
}

func formatSecretOutput(output string) string {
	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		return output
	}
	if strings.HasPrefix(trimmed, "[") {
		items, err := parseSecretListOutput(trimmed)
		if err != nil {
			return output
		}
		sortSecretsByRecent(items)
		return renderSecretList(items, "Список секретов")
	}
	if strings.HasPrefix(trimmed, "{") {
		item, err := parseSecretOutput(trimmed)
		if err != nil || strings.TrimSpace(item.ID) == "" {
			return output
		}
		return renderSecretList([]secretOutputItem{item}, "Карточка секрета")
	}
	return output
}

func parseSecretListOutput(output string) ([]secretOutputItem, error) {
	var items []secretOutputItem
	if err := json.Unmarshal([]byte(output), &items); err != nil {
		return nil, err
	}
	return items, nil
}

func parseSecretOutput(output string) (secretOutputItem, error) {
	var item secretOutputItem
	err := json.Unmarshal([]byte(output), &item)
	return item, err
}

func parseTagQuery(raw string) []string {
	parts := strings.Split(raw, ",")
	tags := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, part := range parts {
		tag := strings.ToLower(strings.TrimSpace(part))
		if tag == "" {
			continue
		}
		if _, exists := seen[tag]; exists {
			continue
		}
		seen[tag] = struct{}{}
		tags = append(tags, tag)
	}
	return tags
}

func cloneStringMap(source map[string]string) map[string]string {
	if source == nil {
		return map[string]string{}
	}
	copyMap := make(map[string]string, len(source))
	for key, value := range source {
		copyMap[key] = value
	}
	return copyMap
}

func loadSecretSelectionCmd(filters map[string]string) tea.Cmd {
	selectionFilters := cloneStringMap(filters)
	return func() tea.Msg {
		items, err := loadSecretsForSelection(selectionFilters)
		return secretSelectionLoadedMsg{Items: items, Err: err}
	}
}

func loadSecretsForSelection(filters map[string]string) ([]secretOutputItem, error) {
	rawOutput, err := executeCLI([]string{"secrets", "list"})
	if err != nil {
		return nil, err
	}
	items, parseErr := parseSecretListOutput(rawOutput)
	if parseErr != nil {
		return nil, errors.New("не удалось загрузить список секретов")
	}
	filtered := applySecretSelectionFilters(items, filters)
	sortSecretsByRecent(filtered)
	return filtered, nil
}

func sortSecretsByRecent(items []secretOutputItem) {
	sort.Slice(items, func(i, j int) bool {
		return compareSecretDate(items[i].UpdatedAt, items[j].UpdatedAt)
	})
}

func compareSecretDate(left, right string) bool {
	leftTime, leftErr := time.Parse(time.RFC3339, strings.TrimSpace(left))
	rightTime, rightErr := time.Parse(time.RFC3339, strings.TrimSpace(right))
	if leftErr != nil && rightErr != nil {
		return strings.TrimSpace(left) > strings.TrimSpace(right)
	}
	if leftErr != nil {
		return false
	}
	if rightErr != nil {
		return true
	}
	return leftTime.After(rightTime)
}

func applySecretSelectionFilters(items []secretOutputItem, filters map[string]string) []secretOutputItem {
	titleQuery := strings.ToLower(strings.TrimSpace(filters[fieldFindTitle]))
	tagQuery := parseTagQuery(filters[fieldFindTags])
	dateQuery := strings.TrimSpace(filters[fieldFindDate])

	matched := make([]secretOutputItem, 0, len(items))
	for _, item := range items {
		if titleQuery != "" {
			title := strings.ToLower(strings.TrimSpace(item.MetaOpen.Title))
			if !strings.Contains(title, titleQuery) {
				continue
			}
		}
		if len(tagQuery) > 0 && !secretMatchesTags(item, tagQuery) {
			continue
		}
		if !isOnOrAfterDate(item.UpdatedAt, dateQuery) {
			continue
		}
		matched = append(matched, item)
	}
	return matched
}

func isOnOrAfterDate(updatedAt, dateQuery string) bool {
	query := strings.TrimSpace(dateQuery)
	if query == "" {
		return true
	}
	fromDate, err := time.Parse("2006-01-02", query)
	if err != nil {
		return false
	}
	itemDate := extractDate(updatedAt)
	if itemDate == "" {
		return false
	}
	parsedItemDate, err := time.Parse("2006-01-02", itemDate)
	if err != nil {
		return false
	}
	return !parsedItemDate.Before(fromDate)
}

func extractSecretIDFromOutput(output string) string {
	item, err := parseSecretOutput(strings.TrimSpace(output))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(item.ID)
}

func appendRevalidationOutput(mainOutput, revalidation string) string {
	base := strings.TrimSpace(mainOutput)
	recheck := strings.TrimSpace(revalidation)
	if base == "" {
		return recheck
	}
	if recheck == "" {
		return base + "\n"
	}
	return base + "\n\n" + recheck + "\n"
}

func revalidateSecretsAfterMutation(actionID, affectedID string, shouldExist bool) string {
	output, err := executeCLI([]string{"secrets", "list"})
	if err != nil {
		return "Ревалидация: не удалось получить актуальный список секретов."
	}
	items, parseErr := parseSecretListOutput(output)
	if parseErr != nil {
		return "Ревалидация: список получен, но не удалось разобрать ответ сервера."
	}

	headline := "Ревалидация после операции"
	if actionID != "" {
		headline = "Ревалидация после " + actionID
	}

	statusLine := "Состояние: данные на сервере актуальны"
	trimmedID := strings.TrimSpace(affectedID)
	if trimmedID != "" {
		exists := secretIDExists(items, trimmedID)
		switch {
		case shouldExist && exists:
			statusLine = "Состояние: изменения подтверждены сервером"
		case shouldExist && !exists:
			statusLine = "Состояние: предупреждение, секрет пока не найден в списке"
		case !shouldExist && !exists:
			statusLine = "Состояние: удаление подтверждено сервером"
		case !shouldExist && exists:
			statusLine = "Состояние: предупреждение, секрет все еще в списке"
		}
	}

	ordered := append([]secretOutputItem(nil), items...)
	sortSecretsByRecent(ordered)

	var b strings.Builder
	b.WriteString(headline)
	b.WriteString("\n")
	b.WriteString(statusLine)
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Всего секретов: %d", len(ordered)))

	if len(ordered) > 0 {
		b.WriteString("\n")
		b.WriteString("Последние изменения:")
		limit := minInt(5, len(ordered))
		for i := 0; i < limit; i++ {
			item := ordered[i]
			b.WriteString("\n")
			b.WriteString(fmt.Sprintf("- %s | %s", secretDisplayTitle(item), fallbackText(item.UpdatedAt)))
		}
		if len(ordered) > limit {
			b.WriteString("\n")
			b.WriteString(fmt.Sprintf("И еще %d секрет(ов).", len(ordered)-limit))
		}
	}

	return b.String()
}

func secretIDExists(items []secretOutputItem, id string) bool {
	for _, item := range items {
		if strings.TrimSpace(item.ID) == id {
			return true
		}
	}
	return false
}

func extractDate(timestamp string) string {
	trimmed := strings.TrimSpace(timestamp)
	if trimmed == "" {
		return ""
	}
	if parsed, err := time.Parse(time.RFC3339, trimmed); err == nil {
		return parsed.UTC().Format("2006-01-02")
	}
	if len(trimmed) >= 10 {
		return trimmed[:10]
	}
	return ""
}

func secretMatchesTags(item secretOutputItem, searchTags []string) bool {
	if len(searchTags) == 0 {
		return true
	}
	if len(item.MetaOpen.Tags) == 0 {
		return false
	}
	available := make(map[string]struct{}, len(item.MetaOpen.Tags))
	for _, tag := range item.MetaOpen.Tags {
		available[strings.ToLower(strings.TrimSpace(tag))] = struct{}{}
	}
	for _, tag := range searchTags {
		if _, exists := available[tag]; exists {
			return true
		}
	}
	return false
}

func renderSecretList(items []secretOutputItem, headline string) string {
	var b strings.Builder
	b.WriteString(headline)
	b.WriteString("\n")
	b.WriteString(strings.Repeat("-", len([]rune(headline))))
	b.WriteString("\n")

	if len(items) == 0 {
		b.WriteString("Секреты не найдены.\n")
		return b.String()
	}

	b.WriteString(fmt.Sprintf("Найдено: %d\n", len(items)))
	visibleCount := minInt(5, len(items))
	for i := 0; i < visibleCount; i++ {
		item := items[i]
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("%d) %s\n", i+1, secretDisplayTitle(item)))
		b.WriteString(fmt.Sprintf("   ID: %s\n", fallbackText(item.ID)))
		if len(item.MetaOpen.Tags) > 0 {
			b.WriteString(fmt.Sprintf("   Теги: %s\n", strings.Join(item.MetaOpen.Tags, ", ")))
		} else {
			b.WriteString("   Теги: -\n")
		}
		b.WriteString(fmt.Sprintf("   Сайт: %s\n", fallbackText(item.MetaOpen.Site)))
		b.WriteString(fmt.Sprintf("   Обновлен: %s\n", fallbackText(item.UpdatedAt)))
		appendSecretData(&b, decodeSecretData(item.Ciphertext))
	}
	if len(items) > visibleCount {
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("И еще %d секрет(ов). Уточните фильтры поиска, чтобы сузить список.\n", len(items)-visibleCount))
	}
	b.WriteString("\n")
	return b.String()
}

func secretDisplayTitle(item secretOutputItem) string {
	title := strings.TrimSpace(item.MetaOpen.Title)
	if title == "" {
		title = "Без названия"
	}
	typeName := strings.TrimSpace(item.Type)
	if typeName == "" {
		return title
	}
	return title + " [" + typeName + "]"
}

func fallbackText(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "-"
	}
	return trimmed
}

func decodeSecretData(ciphertext string) string {
	raw := strings.TrimSpace(ciphertext)
	if raw == "" {
		return "(пусто)"
	}
	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return "(данные недоступны)"
	}
	text := strings.TrimSpace(string(decoded))
	if text == "" {
		return "(пусто)"
	}
	runes := []rune(text)
	if len(runes) > 280 {
		return string(runes[:280]) + "..."
	}
	return text
}

func appendSecretData(b *strings.Builder, data string) {
	if !strings.Contains(data, "\n") {
		b.WriteString(fmt.Sprintf("   Данные: %s\n", data))
		return
	}
	b.WriteString("   Данные:\n")
	for _, line := range strings.Split(data, "\n") {
		b.WriteString("     ")
		b.WriteString(line)
		b.WriteString("\n")
	}
}
