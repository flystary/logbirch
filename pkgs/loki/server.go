package loki

import (
	"fmt"
	"net/http"
	"sort"
	"sync/atomic"

	"logbirch/pkgs/dispatcher"
	"github.com/gin-gonic/gin"
	"github.com/grafana/dskit/tenant"
	"github.com/grafana/loki/v3/pkg/loghttp/push"
	"github.com/grafana/loki/v3/pkg/logproto"
	promql_parser "github.com/prometheus/prometheus/promql/parser"
	"github.com/samber/do"
)

type Loki struct {
	e *gin.Engine
}

type emptyLogger struct{}

func (l *emptyLogger) Log(keyvals ...interface{}) error {
	return nil
}

const PUSH_ENDPOINT = "/loki/api/v1/push"

func BuildLoki(i *do.Injector) (*Loki, error) {
	d := do.MustInvoke[*dispatcher.Dispatcher](i)
	instanceId := do.MustInvokeNamed[string](i, "instanceId")
	gin.SetMode(gin.ReleaseMode)
	g := gin.New()
	g.Use(gin.Recovery())

	var index atomic.Uint64
	g.POST(PUSH_ENDPOINT, func(c *gin.Context) {
		var err error
		var pushRequest *logproto.PushRequest
		userID, _ := tenant.TenantID(c.Request.Context())
		if pushRequest, err = push.ParseRequest(&emptyLogger{}, userID, c.Request, nil, push.EmptyLimits{}, push.ParseLokiRequest, nil); err != nil {
			fmt.Printf("parse request error: %v\n", err)
			c.JSON(400, gin.H{
				"err": err.Error(),
			})
			return
		}

		for _, stream := range pushRequest.Streams {
			ls, err := promql_parser.ParseMetric(stream.Labels)
			if err != nil {
				fmt.Printf("loki parse metric error: %v\n", err)
				continue
			}
			sort.Sort(ls)
			labels := ls.Map()

			for _, entry := range stream.Entries {
				d.Dispatch(&dispatcher.LogPart{
					Labels:     labels,
					Message:    entry.Line,
					Timestamp:  entry.Timestamp,
					InstanceId: instanceId,
					Index:      index.Add(1),
					Tags:       []string{},
				})
			}
		}

		c.Status(http.StatusNoContent)
	})

	return &Loki{
		e: g,
	}, nil
}

func (e *Loki) Run() {
	fmt.Println("loki server running on :3100")
	err := e.e.Run("127.0.0.1:3100")

	if err != nil {
		fmt.Printf("loki stopped err: %v\n", err)
	}
}
