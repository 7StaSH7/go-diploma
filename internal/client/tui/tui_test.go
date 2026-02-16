package tui

import (
	"strings"
	"testing"
)

func actionIDs(actions []tuiAction) []string {
	ids := make([]string, 0, len(actions))
	for _, action := range actions {
		ids = append(ids, action.ID)
	}
	return ids
}

func TestAppendOptionalFlag(t *testing.T) {
	args := appendOptionalFlag(nil, "--since", "2026-01-01T00:00:00Z")
	if len(args) != 2 || args[0] != "--since" || args[1] != "2026-01-01T00:00:00Z" {
		t.Fatalf("unexpected args: %#v", args)
	}

	args = appendOptionalFlag(args, "--since", "   ")
	if len(args) != 2 {
		t.Fatalf("blank value should not append, got: %#v", args)
	}
}

func TestEncodeSecretData(t *testing.T) {
	encoded := encodeSecretData("hello")
	if encoded != "aGVsbG8=" {
		t.Fatalf("unexpected encoded value: %s", encoded)
	}
}

func TestFormatSecretOutput(t *testing.T) {
	raw := `{"id":"sec-1","type":"note","ciphertext":"aGVsbG8="}`
	humanized := formatSecretOutput(raw)

	if !strings.Contains(humanized, "Данные: hello") {
		t.Fatalf("expected decoded data in human output: %s", humanized)
	}
	if strings.Contains(humanized, "ciphertext") {
		t.Fatalf("ciphertext should be hidden: %s", humanized)
	}
}

func TestFormatSecretOutputArray(t *testing.T) {
	raw := `[{"id":"a","ciphertext":"YQ=="},{"id":"b","ciphertext":"Yg=="}]`
	humanized := formatSecretOutput(raw)

	if !strings.Contains(humanized, "Данные: a") || !strings.Contains(humanized, "Данные: b") {
		t.Fatalf("expected decoded values in array output: %s", humanized)
	}
	if !strings.Contains(humanized, "Найдено: 2") {
		t.Fatalf("expected count in output: %s", humanized)
	}
	if strings.Contains(humanized, "ciphertext") {
		t.Fatalf("ciphertext should be hidden in array output: %s", humanized)
	}
}

func TestParseTagQuery(t *testing.T) {
	tags := parseTagQuery("  Work,work,  почта ,, ")
	if len(tags) != 2 {
		t.Fatalf("expected deduplicated tags, got: %#v", tags)
	}
	if tags[0] != "work" || tags[1] != "почта" {
		t.Fatalf("unexpected tags: %#v", tags)
	}
}

func TestResolveUpdateInput(t *testing.T) {
	if got := resolveUpdateInput("", "current"); got != "current" {
		t.Fatalf("expected current value, got: %q", got)
	}
	if got := resolveUpdateInput("-", "current"); got != "" {
		t.Fatalf("expected clear marker to reset value, got: %q", got)
	}
	if got := resolveUpdateInput("new", "current"); got != "new" {
		t.Fatalf("expected new value, got: %q", got)
	}
}

func TestValidateField(t *testing.T) {
	if err := validateField(tuiField{Label: "ID", Required: true}, ""); err == nil {
		t.Fatal("expected required validation error")
	}
	if err := validateField(tuiField{Key: "since"}, "bad-time"); err == nil {
		t.Fatal("expected RFC3339 validation error")
	}
	if err := validateField(tuiField{Key: "since"}, "2026-02-09T10:00:00Z"); err != nil {
		t.Fatalf("unexpected error for valid timestamp: %v", err)
	}

	mandatoryErr := validateField(tuiField{Label: "Логин", Required: true}, "")
	if !strings.Contains(mandatoryErr.Error(), "поле обязательно") {
		t.Fatalf("unexpected mandatory error: %v", mandatoryErr)
	}
}

func TestVisibleActionsForGuest(t *testing.T) {
	m := tuiModel{authorized: false}
	ids := actionIDs(m.visibleActions())

	if len(ids) != 2 {
		t.Fatalf("guest should see only signup/signin, got: %v", ids)
	}
	if ids[0] != "signup" || ids[1] != "signin" {
		t.Fatalf("unexpected guest actions: %v", ids)
	}
}

func TestVisibleActionsForAuthorizedUser(t *testing.T) {
	m := tuiModel{authorized: true}
	ids := actionIDs(m.visibleActions())

	for _, id := range ids {
		if id == "signup" || id == "signin" {
			t.Fatalf("authorized user should not see auth commands, got: %v", ids)
		}
	}
	if len(ids) == 0 {
		t.Fatal("authorized user should see non-auth actions")
	}
}

func TestIDFieldsAreHiddenInUserActions(t *testing.T) {
	for _, action := range tuiActions {
		if action.ID != "search" && action.ID != "update" && action.ID != "delete" {
			continue
		}
		for _, field := range action.Fields {
			if field.Key == "id" {
				t.Fatalf("action %s should not require id field", action.ID)
			}
		}
	}
}

func TestApplySecretSelectionFilters(t *testing.T) {
	items := []secretOutputItem{
		{
			ID:        "a",
			UpdatedAt: "2026-02-09T10:00:00Z",
			MetaOpen: secretOutputMeta{
				Title: "Рабочая почта",
				Tags:  []string{"work", "mail"},
			},
		},
		{
			ID:        "b",
			UpdatedAt: "2026-02-08T10:00:00Z",
			MetaOpen: secretOutputMeta{
				Title: "Домашний банк",
				Tags:  []string{"home", "finance"},
			},
		},
	}

	byTitle := applySecretSelectionFilters(items, map[string]string{fieldFindTitle: "почта"})
	if len(byTitle) != 1 || byTitle[0].ID != "a" {
		t.Fatalf("unexpected title filter result: %#v", byTitle)
	}

	byTags := applySecretSelectionFilters(items, map[string]string{fieldFindTags: "finance,other"})
	if len(byTags) != 1 || byTags[0].ID != "b" {
		t.Fatalf("unexpected tags filter result: %#v", byTags)
	}

	byDate := applySecretSelectionFilters(items, map[string]string{fieldFindDate: "2026-02-09"})
	if len(byDate) != 1 || byDate[0].ID != "a" {
		t.Fatalf("unexpected date filter result: %#v", byDate)
	}
}

func TestExtractDate(t *testing.T) {
	if got := extractDate("2026-02-09T10:00:00Z"); got != "2026-02-09" {
		t.Fatalf("unexpected extracted date: %s", got)
	}
	if got := extractDate("bad-value"); got != "" {
		t.Fatalf("expected empty date for invalid value, got: %s", got)
	}
}
