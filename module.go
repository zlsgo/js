package js

import (
	"errors"

	"github.com/dop251/goja"
	"github.com/sohaha/zlsgo/ztype"
)

func (vm *VM) RunModule(code []byte, rendered ...func(*goja.Runtime) (goja.Value, error)) (result ztype.Type, err error) {
	p, err := vm.getProgram(code, true)
	if err != nil {
		return ztype.New(nil), err
	}

	return vm.RunProgram(p, func(r *goja.Runtime) (result goja.Value, err error) {
		defer r.GlobalObject().Delete("exports")
		if len(rendered) == 0 {
			exports := r.GlobalObject().Get("exports")
			return exports, nil
		}

		for i := range rendered {
			result, err = rendered[i](r)
			if err != nil {
				return
			}
		}

		return
	})
}

func (vm *VM) RunModuleForMethod(code []byte, method string, args ...interface{}) (result interface{}, err error) {
	return vm.RunModule(code, func(r *goja.Runtime) (goja.Value, error) {
		fn, ok := goja.AssertFunction(r.GlobalObject().Get("exports").ToObject(r).Get(method))
		if !ok {
			return nil, errors.New("method " + method + " not found")
		}

		return vm.runMethod(r, fn, args)
	})
}

func (vm *VM) runMethod(r *goja.Runtime, method goja.Callable, args []interface{}) (goja.Value, error) {
	values := make([]goja.Value, 0, len(args))
	for i := range args {
		switch v := args[i].(type) {
		case ztype.Map:
			values = append(values, r.ToValue(map[string]interface{}(v)))
		default:
			values = append(values, r.ToValue(v))
		}
	}

	return vm.timeout(r, func() (goja.Value, error) {
		return method(goja.Undefined(), values...)
	})
}
