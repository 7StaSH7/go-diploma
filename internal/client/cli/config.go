package cli

import (
	"strings"

	appconfig "github.com/7StaSH7/practicum-diploma/internal/config"
)

func defaultServerURL() string {
	cfg, err := appconfig.Load()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(cfg.ServerURL)
}

func effectiveServerURL(overrideURL, sessionURL string) string {
	override := strings.TrimSpace(overrideURL)
	if override != "" {
		return override
	}
	fromConfig := defaultServerURL()
	if fromConfig != "" {
		return fromConfig
	}
	return strings.TrimSpace(sessionURL)
}
