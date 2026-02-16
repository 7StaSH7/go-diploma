package cli

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
	"time"

	dtosecret "github.com/7StaSH7/practicum-diploma/internal/dto/secret"
	"github.com/7StaSH7/practicum-diploma/internal/models"
)

func parseSecretWriteFlags(name string, args []string, includeID bool) (dtosecret.SecretPayload, string, string, error) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	serverURL := fs.String("server", "", "Server base URL")
	secretType := fs.String("type", "", "Secret type")
	ciphertext := fs.String("ciphertext", "", "Base64 ciphertext")
	title := fs.String("title", "", "Meta title")
	tags := fs.String("tags", "", "Comma-separated tags")
	site := fs.String("site", "", "Meta site")
	var id *string
	if includeID {
		id = fs.String("id", "", "Secret ID")
	}
	if err := fs.Parse(args); err != nil {
		return dtosecret.SecretPayload{}, "", "", err
	}
	if strings.TrimSpace(*secretType) == "" {
		return dtosecret.SecretPayload{}, "", "", errors.New("--type is required")
	}
	if strings.TrimSpace(*ciphertext) == "" {
		return dtosecret.SecretPayload{}, "", "", errors.New("--ciphertext is required")
	}
	payload := dtosecret.SecretPayload{
		Type:       strings.TrimSpace(*secretType),
		Ciphertext: strings.TrimSpace(*ciphertext),
		MetaOpen: models.MetaOpen{
			Title: strings.TrimSpace(*title),
			Site:  strings.TrimSpace(*site),
			Tags:  parseCSV(*tags),
		},
	}
	secretID := ""
	if includeID {
		secretID = strings.TrimSpace(*id)
		if secretID == "" {
			return dtosecret.SecretPayload{}, "", "", errors.New("--id is required")
		}
	}
	return payload, strings.TrimSpace(*serverURL), secretID, nil
}

func parseCSV(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	tags := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value != "" {
			tags = append(tags, value)
		}
	}
	if len(tags) == 0 {
		return nil
	}
	return tags
}

func findLatestUpdatedAt(secrets []dtosecret.SecretResponse) (string, bool) {
	var latest time.Time
	var found bool
	for _, secret := range secrets {
		parsed, err := time.Parse(time.RFC3339, secret.UpdatedAt)
		if err != nil {
			continue
		}
		if !found || parsed.After(latest) {
			found = true
			latest = parsed
		}
	}
	if !found {
		return "", false
	}
	return latest.UTC().Format(time.RFC3339), true
}

func printJSON(w io.Writer, value any) error {
	encoded, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "%s\n", encoded)
	return err
}
