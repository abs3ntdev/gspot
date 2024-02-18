package commands

import "strings"

func isNoActiveError(err error) bool {
	return strings.Contains(err.Error(), "No active device found")
}
