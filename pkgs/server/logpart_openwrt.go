//go:build openwrt

package server

import (
	"time"

	"logbirch/pkgs/dispatcher"
	"gopkg.in/mcuadros/go-syslog.v2/format"
)

func parseContents(logParts format.LogParts, instanceId string, index uint64) dispatcher.LogPart {
	var part dispatcher.LogPart

	data := logParts["content"].(string)
	part.Message = data
	part.Timestamp = time.Now()

	part.Tags = []string{
		"openwrt",
	}

	tag, ok := logParts["tag"]
	if ok {
		part.Tags = append(part.Tags, tag.(string))
	}

	part.Labels = map[string]string{}
	part.InstanceId = instanceId
	part.Index = index
	return part
}
