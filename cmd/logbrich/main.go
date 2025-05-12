package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/google/uuid"

	"logbirch/pkgs/dispatcher"
	_ "logbirch/pkgs/filters"
	"logbirch/pkgs/handlers/stdio"
	"logbirch/pkgs/handlers/tencent"
	"logbirch/pkgs/loki"
	"logbirch/pkgs/server"
	"github.com/samber/do"
)

func RegisterBuild[T any](i *do.Injector, f do.Provider[T]) T {
	do.Provide[T](i, f)
	return do.MustInvoke[T](i)
}

func main() {
	debug.SetTraceback("crash")

	di := do.New()

	do.Provide(di, server.BuildSysLogServer)
	do.Provide(di, loki.BuildLoki)
	do.Provide(di, dispatcher.BuildDispatcher)

	if os.Getenv("ENABLE_STDIO") == "true" {
		RegisterBuild(di, stdio.BuildStdIOHandler)
	}
	instanceId, err := uuid.NewUUID()
	if err != nil {
		panic(err)
	}
	do.ProvideNamedValue(di, "instanceId", instanceId.String())

	reportHandler := RegisterBuild(di, tencent.BuildTencentAnymouseHandler)

	server := do.MustInvoke[*server.SysLogServer](di)

	loki := do.MustInvoke[*loki.Loki](di)
	go loki.Run()

	fmt.Printf("logbirch server serve at :514, instanceId=%v\n", instanceId.String())
	go server.Run(instanceId.String())

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

	<-ch
	reportHandler.ReportNow()
}
