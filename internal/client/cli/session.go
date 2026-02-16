package cli

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

func sessionPath() (string, error) {
	if override := os.Getenv("PKEEPER_SESSION_PATH"); override != "" {
		return override, nil
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "pkeeper", "session.json"), nil
}

func loadSession() (session, error) {
	path, err := sessionPath()
	if err != nil {
		return session{}, err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return session{}, errNoSession
		}
		return session{}, err
	}
	var sess session
	if err := json.Unmarshal(raw, &sess); err != nil {
		return session{}, err
	}
	return sess, nil
}

func saveSession(sess session) error {
	path, err := sessionPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	encoded, err := json.MarshalIndent(sess, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, encoded, 0o600)
}

func AuthorizedSession() (bool, error) {
	sess, err := loadSession()
	if err != nil {
		return false, err
	}
	return HasAuthorizedTokens(sess.AccessToken, sess.RefreshToken), nil
}

func HasAuthorizedTokens(accessToken, refreshToken string) bool {
	return strings.TrimSpace(accessToken) != "" && strings.TrimSpace(refreshToken) != ""
}

func HasAuthorizedSession() bool {
	authorized, err := AuthorizedSession()
	if err != nil {
		return false
	}
	return authorized
}
