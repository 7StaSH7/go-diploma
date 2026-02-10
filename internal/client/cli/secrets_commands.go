package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/7StaSH7/practicum-diploma/internal/api"
	dtosecret "github.com/7StaSH7/practicum-diploma/internal/dto/secret"
)

func runSecretsList(args []string, stdout io.Writer, updateLastSync bool) error {
	fs := flag.NewFlagSet("secrets list", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	serverURL := fs.String("server", "", "Server base URL")
	since := fs.String("since", "", "RFC3339 timestamp")
	if err := fs.Parse(args); err != nil {
		return err
	}

	effectiveSince := strings.TrimSpace(*since)
	sess, result, err := runAuthorizedRequest(strings.TrimSpace(*serverURL), func(ctx context.Context, client *api.API, accessToken string, sess session) ([]dtosecret.SecretResponse, error) {
		if effectiveSince == "" && updateLastSync {
			effectiveSince = sess.LastSyncAt
		}
		secrets, requestErr := client.ListSecrets(ctx, accessToken, effectiveSince)
		return secrets, requestErr
	})
	if err != nil {
		return err
	}

	if updateLastSync && len(result) > 0 {
		if latest, ok := findLatestUpdatedAt(result); ok {
			sess.LastSyncAt = latest
		}
	}
	if err := saveSession(sess); err != nil {
		return err
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
	if strings.TrimSpace(*secretID) == "" {
		return errors.New("--id is required")
	}

	sess, result, err := runAuthorizedRequest(strings.TrimSpace(*serverURL), func(ctx context.Context, client *api.API, accessToken string, sess session) (dtosecret.SecretResponse, error) {
		secret, requestErr := client.GetSecret(ctx, accessToken, strings.TrimSpace(*secretID))
		return secret, requestErr
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

	sess, result, err := runAuthorizedRequest(serverURL, func(ctx context.Context, client *api.API, accessToken string, sess session) (dtosecret.SecretResponse, error) {
		secret, requestErr := client.CreateSecret(ctx, accessToken, payload)
		return secret, requestErr
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

	sess, result, err := runAuthorizedRequest(serverURL, func(ctx context.Context, client *api.API, accessToken string, sess session) (dtosecret.SecretResponse, error) {
		secret, requestErr := client.UpdateSecret(ctx, accessToken, secretID, payload)
		return secret, requestErr
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
	trimmedID := strings.TrimSpace(*secretID)
	if trimmedID == "" {
		return errors.New("--id is required")
	}

	sess, _, err := runAuthorizedRequest(strings.TrimSpace(*serverURL), func(ctx context.Context, client *api.API, accessToken string, sess session) (struct{}, error) {
		requestErr := client.DeleteSecret(ctx, accessToken, trimmedID)
		return struct{}{}, requestErr
	})
	if err != nil {
		return err
	}
	if err := saveSession(sess); err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "secret %s deleted\n", trimmedID)
	return err
}

func runAuthorizedRequest[T any](overrideURL string, request func(ctx context.Context, client *api.API, accessToken string, sess session) (T, error)) (session, T, error) {
	var zero T
	sess, client, err := loadSessionAndClient(strings.TrimSpace(overrideURL))
	if err != nil {
		return session{}, zero, err
	}

	ctx := context.Background()
	var out T
	sess, err = withAutoRefresh(sess, client, func(accessToken string) error {
		result, requestErr := request(ctx, client, accessToken, sess)
		if requestErr != nil {
			return requestErr
		}
		out = result
		return nil
	})
	if err != nil {
		return session{}, zero, err
	}
	return sess, out, nil
}
