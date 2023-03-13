package js

import (
	"errors"
	"time"

	"github.com/dop251/goja"
	"github.com/sohaha/zlsgo/zerror"
	"github.com/sohaha/zlsgo/ztype"
)

func (vm *VM) RunCommonjs(code []byte, rendered ...func(*goja.Runtime) (interface{}, error)) (result interface{}, err error) {
	p, err := vm.getProgram(code, true)
	if err != nil {
		return nil, err
	}

	return vm.RunProgram(p, func(r *goja.Runtime) (result interface{}, err error) {
		defer r.GlobalObject().Delete("exports")
		if len(rendered) == 0 {
			exports := r.GlobalObject().Get("exports")
			return exports.Export(), nil
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

func (vm *VM) RunCommonjsForMethod(code []byte, method string, args ...interface{}) (result interface{}, err error) {
	return vm.RunCommonjs(code, func(r *goja.Runtime) (interface{}, error) {
		fn, ok := goja.AssertFunction(r.GlobalObject().Get("exports").ToObject(r).Get(method))
		if !ok {
			return nil, errors.New("method " + method + " not found")
		}

		values := make([]goja.Value, 0, len(args))
		for i := range args {
			switch v := args[i].(type) {
			case ztype.Map:
				values = append(values, r.ToValue(map[string]interface{}(v)))
			default:
				values = append(values, r.ToValue(v))
			}
		}
		ch := make(chan struct{})

		var (
			res goja.Value
			err error
		)

		go func() {
			err = zerror.TryCatch(func() error {
				res, err = fn(goja.Undefined(), values...)
				return err
			})
			ch <- struct{}{}
		}()

		select {
		case <-ch:
			if err != nil {
				return nil, err
			}

			return res.Export(), nil
		case <-time.After(vm.timeout):
			r.Interrupt("timeout")
			return nil, errors.New("timeout: " + vm.timeout.String())
		}
	})
}
