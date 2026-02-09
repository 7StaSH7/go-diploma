package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	dtoauth "github.com/7StaSH7/practicum-diploma/internal/dto/auth"
	dtosecret "github.com/7StaSH7/practicum-diploma/internal/dto/secret"
)

func TestSyncIntervalIsTenSeconds(t *testing.T) {
	if syncInterval != 10*time.Second {
		t.Fatalf("unexpected sync interval: %v", syncInterval)
	}
}

func TestHasAuthorizedSession(t *testing.T) {
	t.Setenv("PKEEPER_SESSION_PATH", filepath.Join(t.TempDir(), "session.json"))

	if HasAuthorizedSession() {
		t.Fatal("expected false when session file does not exist")
	}

	if err := saveSession(session{AccessToken: "access", RefreshToken: "refresh"}); err != nil {
		t.Fatalf("save session: %v", err)
	}
	if !HasAuthorizedSession() {
		t.Fatal("expected true with both tokens")
	}
}

func TestHasAuthorizedSessionRequiresBothTokens(t *testing.T) {
	t.Setenv("PKEEPER_SESSION_PATH", filepath.Join(t.TempDir(), "session.json"))

	if err := saveSession(session{AccessToken: "access", RefreshToken: ""}); err != nil {
		t.Fatalf("save session: %v", err)
	}
	if HasAuthorizedSession() {
		t.Fatal("expected false when refresh token is missing")
	}
}

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func installMockHTTPClient(t *testing.T, transport roundTripFunc) {
	t.Helper()
	prev := apiHTTPClientFactory
	apiHTTPClientFactory = func() *http.Client {
		return &http.Client{Transport: transport}
	}
	t.Cleanup(func() {
		apiHTTPClientFactory = prev
	})
}

func jsonResponse(statusCode int, value any) *http.Response {
	body := []byte{}
	if value != nil {
		encoded, _ := json.Marshal(value)
		body = encoded
	}
	return &http.Response{
		StatusCode: statusCode,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(body)),
	}
}

func TestSignupStoresSession(t *testing.T) {
	t.Setenv("PKEEPER_SESSION_PATH", filepath.Join(t.TempDir(), "session.json"))
	installMockHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodPost || req.URL.Path != "/auth/signup" {
			t.Fatalf("unexpected request: %s %s", req.Method, req.URL.Path)
		}
		var payload dtoauth.AuthRequest
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if payload.Login != "alice" || payload.Password != "secret" {
			t.Fatalf("unexpected payload: %+v", payload)
		}
		return jsonResponse(http.StatusOK, dtoauth.AuthResponse{
			UserID:       "u-1",
			AccessToken:  "access-1",
			RefreshToken: "refresh-1",
			KDFSalt:      "salt-1",
		}), nil
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{
		"signup",
		"--server", "http://example.test",
		"--login", "alice",
		"--password", "secret",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code=%d stderr=%s", code, stderr.String())
	}

	sess, err := loadSession()
	if err != nil {
		t.Fatalf("load session: %v", err)
	}
	if sess.ServerURL != "http://example.test" || sess.AccessToken != "access-1" || sess.RefreshToken != "refresh-1" {
		t.Fatalf("unexpected session: %+v", sess)
	}
}

func TestSecretsListAutoRefreshOnUnauthorized(t *testing.T) {
	t.Setenv("PKEEPER_SESSION_PATH", filepath.Join(t.TempDir(), "session.json"))
	if err := saveSession(session{
		ServerURL:    "http://example.test",
		UserID:       "u-1",
		AccessToken:  "old-access",
		RefreshToken: "old-refresh",
		KDFSalt:      "salt-1",
	}); err != nil {
		t.Fatalf("save session: %v", err)
	}

	var secretsCalls int
	var refreshCalls int
	installMockHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		switch req.URL.Path {
		case "/secrets":
			secretsCalls++
			auth := req.Header.Get("Authorization")
			if auth == "Bearer old-access" {
				return jsonResponse(http.StatusUnauthorized, nil), nil
			}
			if auth != "Bearer new-access" {
				t.Fatalf("unexpected auth header: %s", auth)
			}
			return jsonResponse(http.StatusOK, []dtosecret.SecretResponse{
				{ID: "sec-1", UpdatedAt: "2026-02-06T09:00:00Z"},
			}), nil
		case "/auth/refresh":
			refreshCalls++
			var payload dtoauth.RefreshRequest
			if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
				t.Fatalf("decode refresh request: %v", err)
			}
			if payload.RefreshToken != "old-refresh" {
				t.Fatalf("unexpected refresh token: %s", payload.RefreshToken)
			}
			return jsonResponse(http.StatusOK, dtoauth.AuthResponse{
				UserID:       "u-1",
				AccessToken:  "new-access",
				RefreshToken: "new-refresh",
				KDFSalt:      "salt-2",
			}), nil
		default:
			t.Fatalf("unexpected path: %s", req.URL.Path)
		}
		return nil, nil
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"secrets", "list"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code=%d stderr=%s", code, stderr.String())
	}
	if secretsCalls != 2 {
		t.Fatalf("unexpected secrets calls: %d", secretsCalls)
	}
	if refreshCalls != 1 {
		t.Fatalf("unexpected refresh calls: %d", refreshCalls)
	}
	if !strings.Contains(stdout.String(), "sec-1") {
		t.Fatalf("unexpected output: %s", stdout.String())
	}

	sess, err := loadSession()
	if err != nil {
		t.Fatalf("load session: %v", err)
	}
	if sess.AccessToken != "new-access" || sess.RefreshToken != "new-refresh" {
		t.Fatalf("session not updated: %+v", sess)
	}
}

func TestSecretsSyncUpdatesLastSyncAt(t *testing.T) {
	t.Setenv("PKEEPER_SESSION_PATH", filepath.Join(t.TempDir(), "session.json"))
	if err := saveSession(session{
		ServerURL:    "http://example.test",
		UserID:       "u-1",
		AccessToken:  "access",
		RefreshToken: "refresh",
		LastSyncAt:   "2026-02-06T08:00:00Z",
	}); err != nil {
		t.Fatalf("save session: %v", err)
	}

	var capturedSince string
	installMockHTTPClient(t, func(req *http.Request) (*http.Response, error) {
		if req.URL.Path != "/secrets" {
			t.Fatalf("unexpected path: %s", req.URL.Path)
		}
		capturedSince = req.URL.Query().Get("since")
		return jsonResponse(http.StatusOK, []dtosecret.SecretResponse{
			{ID: "a", UpdatedAt: "2026-02-06T09:00:00Z"},
			{ID: "b", UpdatedAt: "2026-02-06T10:00:00Z"},
		}), nil
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"secrets", "sync", "--once"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code=%d stderr=%s", code, stderr.String())
	}
	if capturedSince != "2026-02-06T08:00:00Z" {
		t.Fatalf("unexpected since query: %s", capturedSince)
	}

	sess, err := loadSession()
	if err != nil {
		t.Fatalf("load session: %v", err)
	}
	if sess.LastSyncAt != "2026-02-06T10:00:00Z" {
		t.Fatalf("unexpected last sync: %s", sess.LastSyncAt)
	}
}
