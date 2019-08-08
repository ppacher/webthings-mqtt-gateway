package lua

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/ppacher/webthings-mqtt-gateway/pkg/payload"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
)

var (
	//refTypeLStatePtr  = reflect.TypeOf((*LState)(nil))
	refTypeLuaLValue  = reflect.TypeOf((*lua.LValue)(nil)).Elem()
	refTypeInt        = reflect.TypeOf(int(0))
	refTypeEmptyIface = reflect.TypeOf((*interface{})(nil)).Elem()
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
		fmt.Println(code)
		if !ok {
			return nil, fmt.Errorf("no converter code specified")
		}
	}

	vm := lua.NewState(lua.Options{
		IncludeGoStackTrace: true,
		MinimizeStackMemory: true,
		//SkipOpenLibs:        true,
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

	vm.SetGlobal("map", luar.New(vm, func(m map[string]interface{}) interface{} {
		fmt.Println(m)
		ud := vm.NewUserData()
		ud.Value = m
		return ud
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
		/*
			case lua.LTTable:
				var x map[string]interface{}
				val, err := lValueToReflect(vm, value, reflect.TypeOf(x), nil)
				return val.Interface(), err
		*/
	}

	return nil, fmt.Errorf("invalid result")
}

//
// the following code is take from Gopher-Luar
//

type conversionError struct {
	Lua  lua.LValue
	Hint reflect.Type
}

func (c conversionError) Error() string {
	if _, isNil := c.Lua.(*lua.LNilType); isNil {
		return fmt.Sprintf("cannot use nil as type %s", c.Hint)
	}

	var val interface{}

	if userData, ok := c.Lua.(*lua.LUserData); ok {
		val = userData.Value
	} else {
		val = c.Lua
	}

	return fmt.Sprintf("cannot use %v (type %T) as type %s", val, val, c.Hint)
}

type structFieldError struct {
	Field string
	Type  reflect.Type
}

func (s structFieldError) Error() string {
	return `type ` + s.Type.String() + ` has no field ` + s.Field
}

func lValueToReflect(L *lua.LState, v lua.LValue, hint reflect.Type, tryConvertPtr *bool) (reflect.Value, error) {
	visited := make(map[*lua.LTable]reflect.Value)
	return lValueToReflectInner(L, v, hint, visited, tryConvertPtr)
}

func lValueToReflectInner(L *lua.LState, v lua.LValue, hint reflect.Type, visited map[*lua.LTable]reflect.Value, tryConvertPtr *bool) (reflect.Value, error) {
	if hint.Implements(refTypeLuaLValue) {
		return reflect.ValueOf(v), nil
	}

	switch converted := v.(type) {
	case lua.LBool:
		val := reflect.ValueOf(bool(converted))
		if !val.Type().ConvertibleTo(hint) {
			return reflect.Value{}, conversionError{
				Lua:  v,
				Hint: hint,
			}
		}
		return val.Convert(hint), nil
	case lua.LChannel:
		val := reflect.ValueOf(converted)
		if !val.Type().ConvertibleTo(hint) {
			return reflect.Value{}, conversionError{
				Lua:  v,
				Hint: hint,
			}
		}
		return val.Convert(hint), nil
	case lua.LNumber:
		val := reflect.ValueOf(float64(converted))
		if !val.Type().ConvertibleTo(hint) {
			return reflect.Value{}, conversionError{
				Lua:  v,
				Hint: hint,
			}
		}
		return val.Convert(hint), nil
	case *lua.LFunction:
		panic("function not supported")

	case *lua.LNilType:
		switch hint.Kind() {
		case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice, reflect.UnsafePointer, reflect.Uintptr:
			return reflect.Zero(hint), nil
		}

		return reflect.Value{}, conversionError{
			Lua:  v,
			Hint: hint,
		}

	case *lua.LState:
		val := reflect.ValueOf(converted)
		if !val.Type().ConvertibleTo(hint) {
			return reflect.Value{}, conversionError{
				Lua:  v,
				Hint: hint,
			}
		}
		return val.Convert(hint), nil

	case lua.LString:
		val := reflect.ValueOf(string(converted))
		if !val.Type().ConvertibleTo(hint) {
			return reflect.Value{}, conversionError{
				Lua:  v,
				Hint: hint,
			}
		}
		return val.Convert(hint), nil

	case *lua.LTable:
		if existing := visited[converted]; existing.IsValid() {
			return existing, nil
		}

		if hint == refTypeEmptyIface {
			hint = reflect.MapOf(refTypeEmptyIface, refTypeEmptyIface)
		}

		switch {
		case hint.Kind() == reflect.Array:
			elemType := hint.Elem()
			length := converted.Len()
			if length != hint.Len() {
				return reflect.Value{}, conversionError{
					Lua:  v,
					Hint: hint,
				}
			}
			s := reflect.New(hint).Elem()
			visited[converted] = s

			for i := 0; i < length; i++ {
				value := converted.RawGetInt(i + 1)
				elemValue, err := lValueToReflectInner(L, value, elemType, visited, nil)
				if err != nil {
					return reflect.Value{}, err
				}
				s.Index(i).Set(elemValue)
			}

			return s, nil

		case hint.Kind() == reflect.Slice:
			elemType := hint.Elem()
			length := converted.Len()
			s := reflect.MakeSlice(hint, length, length)
			visited[converted] = s

			for i := 0; i < length; i++ {
				value := converted.RawGetInt(i + 1)
				elemValue, err := lValueToReflectInner(L, value, elemType, visited, nil)
				if err != nil {
					return reflect.Value{}, err
				}
				s.Index(i).Set(elemValue)
			}

			return s, nil

		case hint.Kind() == reflect.Map:
			keyType := hint.Key()
			elemType := hint.Elem()
			s := reflect.MakeMap(hint)
			visited[converted] = s

			for key := lua.LNil; ; {
				var value lua.LValue
				key, value = converted.Next(key)
				if key == lua.LNil {
					break
				}

				lKey, err := lValueToReflectInner(L, key, keyType, visited, nil)
				if err != nil {
					return reflect.Value{}, err
				}
				lValue, err := lValueToReflectInner(L, value, elemType, visited, nil)
				if err != nil {
					return reflect.Value{}, err
				}
				s.SetMapIndex(lKey, lValue)
			}

			return s, nil

		case hint.Kind() == reflect.Ptr && hint.Elem().Kind() == reflect.Struct:
			hint = hint.Elem()
			fallthrough
		case hint.Kind() == reflect.Struct:
			panic("struct not supported")
		}

		return reflect.Value{}, conversionError{
			Lua:  v,
			Hint: hint,
		}

	case *lua.LUserData:
		val := reflect.ValueOf(converted.Value)
		if tryConvertPtr != nil && val.Kind() != reflect.Ptr && hint.Kind() == reflect.Ptr && val.Type() == hint.Elem() {
			newVal := reflect.New(hint.Elem())
			newVal.Elem().Set(val)
			val = newVal
			*tryConvertPtr = true
		} else {
			if !val.Type().ConvertibleTo(hint) {
				return reflect.Value{}, conversionError{
					Lua:  converted,
					Hint: hint,
				}
			}
			val = val.Convert(hint)
			if tryConvertPtr != nil {
				*tryConvertPtr = false
			}
		}
		return val, nil
	}

	panic("never reaches")
}

func init() {
	payload.MustRegisterType("lua", Handler{})
}
