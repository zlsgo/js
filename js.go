package js

import (
	"errors"
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/sohaha/zlsgo/zcache"
	"github.com/sohaha/zlsgo/zerror"
	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/ztype"
)

type VM struct {
	runtime  sync.Pool
	Program  *goja.Program
	Programs *zcache.FastCache
	timer    *time.Timer
	option   Option
}

func (vm *VM) GetRuntime() *goja.Runtime {
	return vm.runtime.Get().(*goja.Runtime)
}

func (vm *VM) PutRuntime(r *goja.Runtime) {
	vm.runtime.Put(r)
}

func (vm *VM) getProgram(code []byte, isExports bool) (p *goja.Program, err error) {
	codeStr := zstring.Bytes2String(code)
	name := codeStr
	if isExports {
		name = zstring.Md5("e::" + name)
	}

	data, ok := vm.Programs.ProvideGet(name, func() (interface{}, bool) {
		var p *goja.Program

		if isExports {
			codeStr = "var exports = {};(function (){" + codeStr + "})()"
		}
		p, err = goja.Compile(name, codeStr, true)
		if err != nil {
			return nil, false
		}
		return p, true
	})

	if !ok {
		if err != nil {
			return nil, err
		}
		return nil, errors.New("compile not existent")
	}

	return data.(*goja.Program), nil
}

func (vm *VM) Run(code []byte, rendered ...func(*goja.Runtime) (goja.Value, error)) (result ztype.Type, err error) {
	p, err := vm.getProgram(code, false)
	if err != nil {
		return ztype.Type{}, err
	}

	return vm.RunProgram(p, rendered...)
}

func (vm *VM) RunString(code string, rendered ...func(*goja.Runtime) (goja.Value, error)) (result ztype.Type, err error) {
	return vm.Run(zstring.String2Bytes(code), rendered...)
}

func (vm *VM) RunForMethod(code []byte, method string, args ...interface{}) (result ztype.Type, err error) {
	return vm.Run(code, func(r *goja.Runtime) (goja.Value, error) {
		fn, ok := goja.AssertFunction(r.Get(method))
		if !ok {
			return nil, errors.New("method " + method + " not found")
		}

		return vm.runMethod(r, fn, args)
	})
}

func (vm *VM) RunProgram(p *goja.Program, rendered ...func(*goja.Runtime) (goja.Value, error)) (ztype.Type, error) {
	if p == nil {
		return ztype.Type{}, errors.New("program is nil")
	}

	r := vm.GetRuntime()
	defer vm.PutRuntime(r)

	res, err := vm.timeout(r, func() (goja.Value, error) {
		value, err := r.RunProgram(p)
		if err == nil {
			for i := range rendered {
				value, err = rendered[i](r)
			}
		}
		return value, err
	})
	if err != nil {
		return ztype.New(nil), err
	}
	return ztype.New(res.Export()), nil
}

func (vm *VM) timeout(r *goja.Runtime, run func() (goja.Value, error)) (goja.Value, error) {
	ch := make(chan error)
	resCh := make(chan goja.Value)

	vm.timer.Reset(vm.option.Timeout)
	defer vm.timer.Stop()

	go func() {
		ch <- zerror.TryCatch(func() error {
			res, err := run()
			if err == nil {
				resCh <- res
			}
			return err
		})
	}()

	select {
	case res := <-resCh:
		return res, nil
	case err := <-ch:
		return nil, err
	case <-vm.timer.C:
		r.Interrupt("timeout")
		return nil, errors.New("timeout")
	}
}
