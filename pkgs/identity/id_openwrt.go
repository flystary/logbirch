//go:build openwrt

package identity

import (
	"os"
	"strings"
	"sync"
	"time"
)

var (
	identityTag = "identity_unknown"
	lock        sync.RWMutex
)

func init() {
	go func() {
		for {
			err := doUpdateIdentity()
			if err == nil {
				break
			}
			time.Sleep(time.Millisecond * 500)
		}
	}()
}

func doUpdateIdentity() error {
	identity, err := os.ReadFile("/etc/device_identity")
	if err != nil {
		return err
	}

	lock.Lock()
	identityTag = strings.TrimSpace(string(identity))
	lock.Unlock()
	return nil
}

func GetIdentity() string {
	lock.RLock()
	defer lock.RUnlock()

	return identityTag
}
