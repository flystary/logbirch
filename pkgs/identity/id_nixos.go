//go:build nixos

package identity

import "os"

func GetIdentity() string {
	name, _ := os.Hostname()
	return name
}
