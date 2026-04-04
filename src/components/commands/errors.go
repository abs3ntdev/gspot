package commands

import "strings"

func IsNoActiveError(err error) bool {
	return strings.Contains(err.Error(), "No active device found")
}

func IsRestrictionError(err error) bool {
	return strings.Contains(err.Error(), "Restriction violated")
}
