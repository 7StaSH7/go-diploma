package version

import "fmt"

var (
	Version   = "dev"
	BuildDate = "unknown"
)

func Info() string {
	return fmt.Sprintf("version: %s\nbuild_date: %s\n", Version, BuildDate)
}
