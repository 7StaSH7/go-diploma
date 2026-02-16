package tui

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/7StaSH7/practicum-diploma/internal/client/cli"
	"github.com/7StaSH7/practicum-diploma/internal/version"
	tea "github.com/charmbracelet/bubbletea"
)

func syncTickCmd() tea.Cmd {
	return tea.Tick(cli.SyncInterval, func(time.Time) tea.Msg {
		return syncTickMsg{}
	})
}

func runTUIActionCmd(actionID string, values map[string]string) tea.Cmd {
	return func() tea.Msg {
		output, err := runTUIAction(actionID, values)
		return operationResultMsg{
			ActionID: actionID,
			Output:   output,
			Err:      err,
		}
	}
}

func runTUIAction(actionID string, values map[string]string) (string, error) {
	if values == nil {
		values = map[string]string{}
	}
	switch actionID {
	case "signup":
		output, err := executeCLI([]string{
			"signup",
			"--login", values["login"],
			"--password", values["password"],
		})
		return output, err
	case "signin":
		output, err := executeCLI([]string{
			"signin",
			"--login", values["login"],
			"--password", values["password"],
		})
		return output, err
	case "search":
		output, err := executeCLI([]string{"secrets", "list"})
		if err != nil {
			return "", err
		}
		items, parseErr := parseSecretListOutput(output)
		if parseErr != nil {
			return "", errors.New("не удалось прочитать список секретов")
		}
		filtered := applySecretSelectionFilters(items, values)
		sortSecretsByRecent(filtered)
		return renderSecretList(filtered, "Результаты поиска"), nil
	case "create":
		ciphertext := encodeSecretData(values["data"])
		args := []string{}
		args = append(args, "--type", defaultSecretType, "--ciphertext", ciphertext)
		args = appendOptionalFlag(args, "--title", values["title"])
		args = appendOptionalFlag(args, "--tags", values["tags"])
		args = appendOptionalFlag(args, "--site", values["site"])
		output, err := executeCLI(append([]string{"secrets", "create"}, args...))
		if err != nil {
			return "", err
		}
		formatted := formatSecretOutput(output)
		revalidated := revalidateSecretsAfterMutation("create", extractSecretIDFromOutput(output), true)
		return appendRevalidationOutput(formatted, revalidated), nil
	case "update":
		snapshot, err := loadSecretSnapshot(values["id"])
		if err != nil {
			return "", fmt.Errorf("не удалось загрузить секрет: %w", err)
		}

		ciphertext := snapshot.Ciphertext
		if data := strings.TrimSpace(values["data"]); data != "" {
			ciphertext = encodeSecretData(data)
		}
		secretType := snapshot.Type
		if strings.TrimSpace(secretType) == "" {
			secretType = defaultSecretType
		}
		title := resolveUpdateInput(values["title"], snapshot.Title)
		tags := resolveUpdateInput(values["tags"], strings.Join(snapshot.Tags, ","))
		site := resolveUpdateInput(values["site"], snapshot.Site)
		args := []string{}
		args = append(args, "--id", values["id"], "--type", secretType, "--ciphertext", ciphertext)
		args = append(args, "--title", title, "--tags", tags, "--site", site)
		output, err := executeCLI(append([]string{"secrets", "update"}, args...))
		if err != nil {
			return "", err
		}
		formatted := formatSecretOutput(output)
		revalidated := revalidateSecretsAfterMutation("update", strings.TrimSpace(values["id"]), true)
		return appendRevalidationOutput(formatted, revalidated), nil
	case "delete":
		args := []string{}
		args = append(args, "--id", values["id"])
		output, err := executeCLI(append([]string{"secrets", "delete"}, args...))
		if err != nil {
			return "", err
		}
		revalidated := revalidateSecretsAfterMutation("delete", strings.TrimSpace(values["id"]), false)
		return appendRevalidationOutput(output, revalidated), nil
	case "auto_sync":
		_, err := executeCLI([]string{"secrets", "sync", "--once"})
		return "", err
	case "version":
		return version.Info(), nil
	default:
		return "", fmt.Errorf("неизвестное действие: %s", actionID)
	}
}

func encodeSecretData(data string) string {
	return base64.StdEncoding.EncodeToString([]byte(data))
}

type secretSnapshot struct {
	Type       string
	Ciphertext string
	Title      string
	Tags       []string
	Site       string
}

func loadSecretSnapshot(secretID string) (secretSnapshot, error) {
	trimmedID := strings.TrimSpace(secretID)
	if trimmedID == "" {
		return secretSnapshot{}, errors.New("пустой ID секрета")
	}
	output, err := executeCLI([]string{"secrets", "get", "--id", trimmedID})
	if err != nil {
		return secretSnapshot{}, err
	}

	var payload struct {
		Type       string `json:"type"`
		Ciphertext string `json:"ciphertext"`
		MetaOpen   struct {
			Title string   `json:"title"`
			Tags  []string `json:"tags"`
			Site  string   `json:"site"`
		} `json:"meta_open"`
	}
	if err := json.Unmarshal([]byte(output), &payload); err != nil {
		return secretSnapshot{}, err
	}

	return secretSnapshot{
		Type:       strings.TrimSpace(payload.Type),
		Ciphertext: strings.TrimSpace(payload.Ciphertext),
		Title:      strings.TrimSpace(payload.MetaOpen.Title),
		Tags:       payload.MetaOpen.Tags,
		Site:       strings.TrimSpace(payload.MetaOpen.Site),
	}, nil
}

func resolveUpdateInput(input, current string) string {
	value := strings.TrimSpace(input)
	if value == "-" {
		return ""
	}
	if value == "" {
		return strings.TrimSpace(current)
	}
	return value
}

func executeCLI(args []string) (string, error) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	code := cli.Execute(args, &out, &errOut)
	if code == 0 {
		return out.String(), nil
	}
	trimmedErr := strings.TrimSpace(errOut.String())
	if trimmedErr == "" {
		trimmedErr = fmt.Sprintf("команда завершилась с ошибкой: %d", code)
	}
	return out.String(), errors.New(trimmedErr)
}

func appendOptionalFlag(args []string, flagName, value string) []string {
	if strings.TrimSpace(value) == "" {
		return args
	}
	return append(args, flagName, strings.TrimSpace(value))
}
