package cli

import (
	"errors"
	"os"
	"strings"

	"github.com/spf13/viper"
)

func defaultServerURL() string {
	v := viper.New()
	v.SetDefault("SERVER_URL", "http://localhost:8080")
	v.SetConfigFile(".env")
	if err := v.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) && !os.IsNotExist(err) {
			return ""
		}
	}
	v.AutomaticEnv()
	return strings.TrimSpace(v.GetString("SERVER_URL"))
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
