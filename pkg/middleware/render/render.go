package render

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/ppacher/mqtt-home/controller/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/macaron.v1"
)

const Unspecified = -1

type OriginalHandler macaron.ReturnHandler

// List of mime-types that we interpret as YAML since there's no
// offical mime-type definition by IANA
var yamlMimeTypes = []string{
	"text/yaml",
	"application/yaml",
	"application/x-yaml",
	"text/vnd.yaml",
	"text/x-yaml",
}

func macaronFallback(ctx *macaron.Context, values []reflect.Value) {
	logrus.Errorf("unsupport API response")
	handler := ctx.GetVal(reflect.TypeOf(OriginalHandler(nil))).Interface()
	handler.(OriginalHandler)(ctx, values)
}

func clientWantsYAML(ctx *macaron.Context) bool {
	accept := ctx.Req.Header.Get("Accept")
	for _, mime := range yamlMimeTypes {
		if strings.Contains(accept, mime) {
			return true
		}
	}

	return false
}

func render(ctx *macaron.Context, code int, payload interface{}) {
	if clientWantsYAML(ctx) {
		blob, err := json.Marshal(payload)
		if err != nil {
			ctx.Error(500, err.Error())
			return
		}

		yamlBlob, err := yaml.JSONToYAML(blob)
		if err != nil {
			ctx.Error(500, err.Error())
			return
		}

		// Copy the accept header
		ctx.Resp.Header().Set("Accept", ctx.Req.Header.Get("Accept"))
		ctx.RawData(code, yamlBlob)
		return
	}

	ctx.JSON(code, payload)
}

func JSONAPI(ctx *macaron.Context, values []reflect.Value) {
	if len(values) == 0 {
		return
	}

	code := 200

	var value reflect.Value

	if len(values) == 1 {
		value = values[0]
	} else if len(values) == 2 {
		if c, ok := values[0].Interface().(int); ok {
			if c != Unspecified {
				code = c
			}
		} else {
			macaronFallback(ctx, values)
			return
		}

		value = values[1]
	} else {
		macaronFallback(ctx, values)
		return
	}

	if err, ok := value.Interface().(errors.HTTPError); ok {
		render(ctx, err.StatusCode(), err)
		return
	}

	if err, ok := value.Interface().(error); ok {
		render(ctx, 500, errors.WrapWithStatus(500, err))
		return
	}

	if code, ok := value.Interface().(int); ok {
		ctx.Status(code)
		return
	}

	render(ctx, code, value.Interface())
}

func Bind(m *macaron.Macaron) {
	// Remap the default ReturnHandler to OriginalHandler
	// so we can fallback
	original := m.GetVal(reflect.TypeOf(macaron.ReturnHandler(nil))).Interface().(macaron.ReturnHandler)
	m.Map(OriginalHandler(original))

	m.Map(macaron.ReturnHandler(JSONAPI))
}
