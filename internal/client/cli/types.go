package cli

import (
	"errors"
	"net/http"
	"time"
)

const defaultHTTPTimeout = 10 * time.Second
const SyncInterval = 10 * time.Second
const syncInterval = SyncInterval

var errNoSession = errors.New("session not found")
var apiHTTPClientFactory = func() *http.Client {
	return &http.Client{Timeout: defaultHTTPTimeout}
}

type session struct {
	ServerURL    string `json:"server_url"`
	UserID       string `json:"user_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	KDFSalt      string `json:"kdf_salt"`
	LastSyncAt   string `json:"last_sync_at,omitempty"`
}
