package commands

import "strings"

func IsNoActiveError(err error) bool {
	return strings.Contains(err.Error(), "No active device found")
}
