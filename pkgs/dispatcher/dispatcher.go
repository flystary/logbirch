package dispatcher

import (
	"time"

	"github.com/samber/do"
)

type LogPart struct {
	Tags       []string          `json:"tags"`
	Labels     map[string]string `json:"label"`
	Message    string            `json:"message"`
	Timestamp  time.Time         `json:"timestamp"`
	InstanceId string            `json:"instanceId"`
	Index      uint64            `json:"index"`
}

type LogContext struct {
	Part     *LogPart
	Filtered bool
}

type Handler func(msg *LogContext)
type FilterHandler = func(msg *LogPart) bool

var handlers []FilterHandler

func RegisterFilter(handler FilterHandler) {
	handlers = append(handlers, handler)
}

type Dispatcher struct {
	handlers map[string]Handler
}

func (s *Dispatcher) RegisterHandler(name string, handler Handler) {
	s.handlers[name] = handler
}

func (s *Dispatcher) Dispatch(msg *LogPart) {
	// Build context
	ctx := &LogContext{
		Part:     msg,
		Filtered: s.filter(msg),
	}

	for _, v := range s.handlers {
		v(ctx)
	}
}

func (s *Dispatcher) filter(input *LogPart) bool {
	for _, it := range handlers {
		if it(input) {
			return true
		}
	}

	return false
}

func BuildDispatcher(i *do.Injector) (*Dispatcher, error) {
	return &Dispatcher{
		handlers: make(map[string]Handler),
	}, nil
}
