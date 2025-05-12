package filters

import (
	"logbirch/pkgs/dispatcher"
)

func filterAddressList(log *dispatcher.LogPart) bool {
	return log.Message == "address list entry added by betaidcapiuser" ||
		log.Message == "address list entry removed by betaidcapiuser"
}

func init() {
	dispatcher.RegisterFilter(filterAddressList)
}
