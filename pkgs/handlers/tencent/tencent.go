package tencent

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"logbirch/pkgs/dispatcher"
	"logbirch/pkgs/identity"
	"github.com/go-resty/resty/v2"
	"github.com/samber/do"
)

var region = ""
var topicId = ""

const maxCacheLines = 100000

func init() {
	if len(os.Getenv("TENCENT_TOPIC_ID")) > 0 {
		topicId = os.Getenv("TENCENT_TOPIC_ID")
	}

	if len(os.Getenv("TENCENT_REGION")) > 0 {
		topicId = os.Getenv("TENCENT_REGION")
	}
}

type LogEntity struct {
	Log  map[string]any `json:"contents"`
	Time uint64         `json:"time"`
}

type LogFlatted struct {
	Tags       string `json:"tags"`
	Message    string `json:"message"`
	Identity   string `json:"identity"`
	InstanceId string `json:"instanceId"`
	Index      string `json:"index"`
}

func LogToFlatted(log *dispatcher.LogPart) map[string]any {
	flatted := &LogFlatted{
		Tags:       strings.Join(log.Tags, ","),
		Message:    log.Message,
		Identity:   identity.GetIdentity(),
		InstanceId: log.InstanceId,
		Index:      fmt.Sprintf("%v", log.Index),
	}
	data, err := json.Marshal(flatted)
	if err != nil {
		return nil
	}

	var inbound map[string]any
	if err = json.Unmarshal(data, &inbound); err != nil {
		return nil
	}

	if log.Labels != nil {
		for k, v := range log.Labels {
			inbound[k] = v
		}
	}

	return inbound
}

// 腾讯云, 日志服务, 匿名写入
type TencentAnymouseHandler struct {
	lock   sync.RWMutex
	caches []LogEntity
}

func BuildTencentAnymouseHandler(i *do.Injector) (*TencentAnymouseHandler, error) {
	server := do.MustInvoke[*dispatcher.Dispatcher](i)

	tencent := &TencentAnymouseHandler{}
	server.RegisterHandler("tencent", tencent.onHandlerMessage)

	t := time.NewTicker(time.Second * 5)

	go func() {
		for {
			<-t.C
			tencent.doReport()
		}
	}()

	return tencent, nil
}

func (t *TencentAnymouseHandler) onHandlerMessage(log *dispatcher.LogContext) {
	// Dont upload log if filtered
	if log.Filtered {
		return
	}

	t.appendLogs([]LogEntity{
		{
			Log:  LogToFlatted(log.Part),
			Time: uint64(log.Part.Timestamp.Unix()),
		},
	})
}

func (t *TencentAnymouseHandler) doReport() {
	currentCaches := t.getCurrentCache()

	if len(currentCaches) == 0 {
		return
	}

	resp, err := resty.New().
		SetTimeout(time.Second*10).
		R().
		SetQueryParam("topic_id", topicId).
		SetBody(map[string]any{
			"logs":   currentCaches,
			"source": "127.0.0.1",
		}).
		Post(fmt.Sprintf("https://%s.cls.tencentcs.com/tracklog", region))

	if err != nil {
		fmt.Printf("report log error: %s\n", err.Error())
		return
	}

	data, _ := io.ReadAll(resp.RawBody())
	if resp.StatusCode() != http.StatusOK {
		fmt.Printf("post log status code not ok status=%v resp=%v\n", resp.StatusCode(), data)
	}

	t.flushCache()
}

func (t *TencentAnymouseHandler) appendLogs(logs []LogEntity) {
	t.lock.Lock()
	t.caches = append(t.caches, logs...)
	if len(t.caches) > maxCacheLines {
		t.caches = t.caches[len(t.caches)-maxCacheLines:]
	}
	t.lock.Unlock()
}

func (t *TencentAnymouseHandler) flushCache() {
	t.lock.Lock()
	t.caches = []LogEntity{}
	t.lock.Unlock()
}

func (t *TencentAnymouseHandler) getCurrentCache() []LogEntity {
	t.lock.RLock()
	logs := make([]LogEntity, len(t.caches))
	copy(logs, t.caches)
	t.lock.RUnlock()
	return logs
}

func (t *TencentAnymouseHandler) ReportNow() {
	t.doReport()
}
