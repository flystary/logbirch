package stdio

import (
	"strings"

	"logbirch/pkgs/dispatcher"
	"github.com/samber/do"
)

type StdIOHandler struct {
}

func BuildStdIOHandler(i *do.Injector) (*StdIOHandler, error) {
	server := do.MustInvoke[*dispatcher.Dispatcher](i)

	std := &StdIOHandler{}
	server.RegisterHandler("stdio", std.onHandlerMessage)

	return std, nil
}

func (std *StdIOHandler) onHandlerMessage(log *dispatcher.LogContext) {
	sb := strings.Builder{}

	sb.WriteString("routeros log: ")

	for _, tag := range log.Part.Tags {
		sb.WriteByte('[')
		sb.WriteString(tag)
		sb.WriteByte(']')
	}

	for key, value := range log.Part.Labels {
		sb.WriteByte('{')
		sb.WriteString(key)
		sb.WriteByte('=')
		sb.WriteString(value)
		sb.WriteByte('}')
	}

	sb.WriteByte(' ')
	sb.WriteString(log.Part.Message)

	println(sb.String())
}
