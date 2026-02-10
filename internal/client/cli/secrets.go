package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
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

	trimmedSince := strings.TrimSpace(*since)
	trimmedServerURL := strings.TrimSpace(*serverURL)

	if trimmedSince != "" {
		if _, err := time.Parse(time.RFC3339, trimmedSince); err != nil {
			return errors.New("--since must be RFC3339")
		}
	}

	if trimmedSince != "" && !*once {
		sess, err := loadSession()
		if err != nil {
			return err
		}
		sess.LastSyncAt = trimmedSince
		if trimmedServerURL != "" {
			sess.ServerURL = trimmedServerURL
		}
		if err := saveSession(sess); err != nil {
			return err
		}
	}

	syncArgs := make([]string, 0, 4)
	if trimmedServerURL != "" {
		syncArgs = append(syncArgs, "--server", trimmedServerURL)
	}
	if *once && trimmedSince != "" {
		syncArgs = append(syncArgs, "--since", trimmedSince)
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
