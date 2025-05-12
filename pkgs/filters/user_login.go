package filters

import (
	"logbirch/pkgs/dispatcher"
	"github.com/samber/lo"
)

func ignoreUserLogin(input *dispatcher.LogPart) bool {
	if lo.IndexOf(input.Tags, "account") != -1 {
		return true
	}

	return false
}

func init() {
	dispatcher.RegisterFilter(ignoreUserLogin)
}
