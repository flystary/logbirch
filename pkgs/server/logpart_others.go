//go:build !openwrt

package server

import (
	"strings"
	"time"

	"logbirch/pkgs/dispatcher"
	"gopkg.in/mcuadros/go-syslog.v2/format"
)

func parseContents(logParts format.LogParts, instanceId string, index uint64) dispatcher.LogPart {
	data := logParts["content"].(string)

	var part dispatcher.LogPart
	part.Timestamp = time.Now()

	before, after, found := strings.Cut(data, " ")
	if found {
		part.Tags = strings.Split(before, ",")
		part.Message = strings.TrimSpace(after)
	} else {
		part.Tags = []string{}
		part.Message = strings.TrimSpace(before)
	}

	part.Labels = map[string]string{}
	part.InstanceId = instanceId
	part.Index = index
	return part
}
