//go:build !(openwrt || nixos)

package identity

import "os"

func GetIdentity() string {
	env := os.Getenv("LOG_IDENTITY")
	if len(env) != 0 {
		return env
	}

	return "unknown"
}
