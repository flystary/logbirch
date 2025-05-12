package server

import (
	"os"

	"logbirch/pkgs/dispatcher"
	"github.com/samber/do"
	"gopkg.in/mcuadros/go-syslog.v2"
)

type SysLogServer struct {
	logServer  *syslog.Server
	channel    syslog.LogPartsChannel
	dispatcher *dispatcher.Dispatcher
}

func BuildSysLogServer(i *do.Injector) (*SysLogServer, error) {
	channel := make(syslog.LogPartsChannel)
	handler := syslog.NewChannelHandler(channel)

	server := syslog.NewServer()
	server.SetFormat(syslog.RFC3164)
	server.SetHandler(handler)

	listenAddr := "0.0.0.0:514"
	if os.Getenv("LISTEN_ADDR") != "" {
		listenAddr = os.Getenv("LISTEN_ADDR")
	}

	err := server.ListenUDP(listenAddr)
	if err != nil {
		return nil, err
	}
	err = server.Boot()
	if err != nil {
		return nil, err
	}

	return &SysLogServer{
		logServer: server,
		channel:   channel,
        dispatcher: do.MustInvoke[*dispatcher.Dispatcher](i),
	}, nil
}

func (s *SysLogServer) Run(instanceId string) {
	go func(channel syslog.LogPartsChannel) {
		index := uint64(0)
		for logParts := range channel {
			part := parseContents(logParts, instanceId, index)
			s.dispatcher.Dispatch(&part)
			index++
		}
	}(s.channel)

	s.logServer.Wait()
}
