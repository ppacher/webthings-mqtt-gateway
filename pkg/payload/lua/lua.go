package lua

import (
	"encoding/json"
	"fmt"

	"github.com/own-home/central/pkg/payload"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
)

// TODO(ppacher) cache lua.LState based on payload.HandlerSpec
// create a go-routine to cleanup unused ones

// Handler implements payload.Handler
type Handler struct{}

// Parse parses the given payload and returns the extracted value. It implements
// the `Parse()` method of `payload.Handler`
func (h Handler) Parse(body []byte, cfg payload.HandlerSpec) (interface{}, error) {
	content, ok := cfg["content"].(string)
	code, ok := cfg["return"].(string)
	if ok {
		code = "return " + code
	}

	if !ok {
		code, ok = cfg["code"].(string)
		if !ok {
			return nil, fmt.Errorf("no converter code specified")
		}
	}

	vm := lua.NewState(lua.Options{
		IncludeGoStackTrace: true,
		MinimizeStackMemory: true,
		SkipOpenLibs:        true,
	})

	if content == "" {
		vm.SetGlobal("value", lua.LString(body))
	} else if content == "json" {
		var x interface{}
		if err := json.Unmarshal(body, &x); err != nil {
			return nil, err
		}

		vm.SetGlobal("value", luar.New(vm, x))
	} else {
		return nil, fmt.Errorf("invalid content type")
	}

	vm.SetGlobal("int", vm.NewFunction(func(L *lua.LState) int {
		n := L.ToInt(1)
		ud := L.NewUserData()
		ud.Value = n
		return 1
	}))

	vm.SetGlobal("bool", vm.NewFunction(func(L *lua.LState) int {
		n := L.ToBool(1)
		ud := L.NewUserData()
		ud.Value = n
		return 1
	}))

	vm.SetGlobal("json", luar.New(vm, func(payload string) interface{} {
		var x interface{}

		if err := json.Unmarshal([]byte(payload), &x); err != nil {
			panic(err)
		}

		return x
	}))

	if err := vm.DoString(code); err != nil {
		return nil, err
	}

	value := vm.Get(1)

	switch value.Type() {
	case lua.LTBool:
		return value == lua.LTrue, nil
	case lua.LTString:
		return value.(lua.LString).String(), nil
	case lua.LTNumber:
		return float64(value.(lua.LNumber)), nil
	case lua.LTUserData:
		return value.(*lua.LUserData).Value, nil
	}
	return nil, fmt.Errorf("invalid result")
}

func init() {
	payload.MustRegisterType("lua", Handler{})
}
