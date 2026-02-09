package cli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	dtosecret "github.com/7StaSH7/practicum-diploma/internal/dto/secret"
	"github.com/7StaSH7/practicum-diploma/internal/models"
)

func runSecrets(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return errors.New("usage: secrets <list|get|create|update|delete|sync>")
	}
	switch args[0] {
	case "list":
		return runSecretsList(args[1:], stdout, false)
	case "sync":
		return runSecretsSync(args[1:], stdout)
	case "get":
		return runSecretsGet(args[1:], stdout)
	case "create":
		return runSecretsCreate(args[1:], stdout)
	case "update":
		return runSecretsUpdate(args[1:], stdout)
	case "delete":
		return runSecretsDelete(args[1:], stdout)
	default:
		return fmt.Errorf("unknown secrets command: %s", args[0])
	}
}

func runSecretsSync(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("secrets sync", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	serverURL := fs.String("server", "", "Server base URL")
	since := fs.String("since", "", "RFC3339 timestamp")
	once := fs.Bool("once", false, "Run one sync iteration")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if strings.TrimSpace(*since) != "" {
		if _, err := time.Parse(time.RFC3339, strings.TrimSpace(*since)); err != nil {
			return errors.New("--since must be RFC3339")
		}
	}

	if strings.TrimSpace(*since) != "" && !*once {
		sess, err := loadSession()
		if err != nil {
			return err
		}
		sess.LastSyncAt = strings.TrimSpace(*since)
		if strings.TrimSpace(*serverURL) != "" {
			sess.ServerURL = strings.TrimSpace(*serverURL)
		}
		if err := saveSession(sess); err != nil {
			return err
		}
	}

	syncArgs := make([]string, 0, 4)
	if strings.TrimSpace(*serverURL) != "" {
		syncArgs = append(syncArgs, "--server", strings.TrimSpace(*serverURL))
	}
	if *once && strings.TrimSpace(*since) != "" {
		syncArgs = append(syncArgs, "--since", strings.TrimSpace(*since))
	}

	if err := runSecretsList(syncArgs, stdout, true); err != nil {
		return err
	}
	if *once {
		return nil
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ticker := time.NewTicker(syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := runSecretsList(syncArgs, stdout, true); err != nil {
				return err
			}
		}
	}
}

func runSecretsList(args []string, stdout io.Writer, updateLastSync bool) error {
	fs := flag.NewFlagSet("secrets list", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	serverURL := fs.String("server", "", "Server base URL")
	since := fs.String("since", "", "RFC3339 timestamp")
	if err := fs.Parse(args); err != nil {
		return err
	}

	sess, client, err := loadSessionAndClient(*serverURL)
	if err != nil {
		return err
	}
	effectiveSince := *since
	if effectiveSince == "" && updateLastSync {
		effectiveSince = sess.LastSyncAt
	}

	var result []dtosecret.SecretResponse
	sess, err = withAutoRefresh(sess, client, func(accessToken string) error {
		secrets, requestErr := client.ListSecrets(context.Background(), accessToken, effectiveSince)
		if requestErr != nil {
			return requestErr
		}
		result = secrets
		return nil
	})
	if err != nil {
		return err
	}
	if err := saveSession(sess); err != nil {
		return err
	}

	if updateLastSync && len(result) > 0 {
		if latest, ok := findLatestUpdatedAt(result); ok {
			sess.LastSyncAt = latest
		}
		if err := saveSession(sess); err != nil {
			return err
		}
	}
	return printJSON(stdout, result)
}

func runSecretsGet(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("secrets get", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	serverURL := fs.String("server", "", "Server base URL")
	secretID := fs.String("id", "", "Secret ID")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *secretID == "" {
		return errors.New("--id is required")
	}

	sess, client, err := loadSessionAndClient(*serverURL)
	if err != nil {
		return err
	}
	var result dtosecret.SecretResponse
	sess, err = withAutoRefresh(sess, client, func(accessToken string) error {
		secret, requestErr := client.GetSecret(context.Background(), accessToken, *secretID)
		if requestErr != nil {
			return requestErr
		}
		result = secret
		return nil
	})
	if err != nil {
		return err
	}
	if err := saveSession(sess); err != nil {
		return err
	}
	return printJSON(stdout, result)
}

func runSecretsCreate(args []string, stdout io.Writer) error {
	payload, serverURL, _, err := parseSecretWriteFlags("secrets create", args, false)
	if err != nil {
		return err
	}

	sess, client, err := loadSessionAndClient(serverURL)
	if err != nil {
		return err
	}
	var result dtosecret.SecretResponse
	sess, err = withAutoRefresh(sess, client, func(accessToken string) error {
		secret, requestErr := client.CreateSecret(context.Background(), accessToken, payload)
		if requestErr != nil {
			return requestErr
		}
		result = secret
		return nil
	})
	if err != nil {
		return err
	}
	if err := saveSession(sess); err != nil {
		return err
	}
	return printJSON(stdout, result)
}

func runSecretsUpdate(args []string, stdout io.Writer) error {
	payload, serverURL, secretID, err := parseSecretWriteFlags("secrets update", args, true)
	if err != nil {
		return err
	}

	sess, client, err := loadSessionAndClient(serverURL)
	if err != nil {
		return err
	}
	var result dtosecret.SecretResponse
	sess, err = withAutoRefresh(sess, client, func(accessToken string) error {
		secret, requestErr := client.UpdateSecret(context.Background(), accessToken, secretID, payload)
		if requestErr != nil {
			return requestErr
		}
		result = secret
		return nil
	})
	if err != nil {
		return err
	}
	if err := saveSession(sess); err != nil {
		return err
	}
	return printJSON(stdout, result)
}

func runSecretsDelete(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("secrets delete", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	serverURL := fs.String("server", "", "Server base URL")
	secretID := fs.String("id", "", "Secret ID")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *secretID == "" {
		return errors.New("--id is required")
	}

	sess, client, err := loadSessionAndClient(*serverURL)
	if err != nil {
		return err
	}
	sess, err = withAutoRefresh(sess, client, func(accessToken string) error {
		return client.DeleteSecret(context.Background(), accessToken, *secretID)
	})
	if err != nil {
		return err
	}
	if err := saveSession(sess); err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "secret %s deleted\n", *secretID)
	return err
}

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
