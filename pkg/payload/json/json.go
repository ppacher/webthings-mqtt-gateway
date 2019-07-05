package json

import (
	"encoding/json"
	"fmt"

	"github.com/ppacher/mqtt-home/controller/pkg/payload"
	"github.com/yalp/jsonpath"
)

// Handler is a `payload.Handler`
type Handler struct {
	pathOverwrite string
}

// Parse implements the `Parse()` method of `payload.Handler`
func (h *Handler) Parse(payload []byte, cfg payload.HandlerSpec) (interface{}, error) {
	var x interface{}
	err := json.Unmarshal(payload, &x)
	if err != nil {
		return nil, err
	}

	path := "$"
	if h.pathOverwrite != "" {
		path = h.pathOverwrite

		if _, ok := cfg["path"]; ok {
			return nil, fmt.Errorf("`path` argument not supported")
		}
	} else {
		p, ok := cfg["path"]
		if ok {
			ps, oks := p.(string)
			if !oks {
				return nil, fmt.Errorf("`path` argument must be a string")
			}

			path = ps
		}
	}

	val, err := jsonpath.Read(x, path)
	return val, err
}

func init() {
	payload.MustRegisterType("json", &Handler{})
	payload.MustRegisterType("json-extended", &Handler{"$.val"})
}
