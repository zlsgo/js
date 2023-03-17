package js

import (
	"errors"
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/sohaha/zlsgo/zcache"
	"github.com/sohaha/zlsgo/zerror"
	"github.com/sohaha/zlsgo/zstring"
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
		name = "e::" + name
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

func (vm *VM) Run(code []byte, rendered ...func(*goja.Runtime) (goja.Value, error)) (result interface{}, err error) {
	p, err := vm.getProgram(code, false)
	if err != nil {
		return nil, err
	}

	return vm.RunProgram(p, rendered...)
}

func (vm *VM) RunForMethod(code []byte, method string, args ...interface{}) (result interface{}, err error) {
	return vm.Run(code, func(r *goja.Runtime) (goja.Value, error) {
		fn, ok := goja.AssertFunction(r.Get(method))
		if !ok {
			return nil, errors.New("method " + method + " not found")
		}

		return vm.runMethod(r, fn, args)
	})
}

func (vm *VM) RunProgram(p *goja.Program, rendered ...func(*goja.Runtime) (goja.Value, error)) (interface{}, error) {
	if p == nil {
		return nil, errors.New("program is nil")
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
		return nil, err
	}
	return res.Export(), nil
}

func (vm *VM) timeout(r *goja.Runtime, run func() (goja.Value, error)) (goja.Value, error) {

	ch := make(chan error)
	resCh := make(chan goja.Value)

	vm.timer.Reset(vm.option.Timeout)
	go func() {
		ch <- zerror.TryCatch(func() error {
			res, err := run()
			if err == nil {
				resCh <- res
			}
			return err
		})
	}()

	for {
		select {
		case res := <-resCh:
			return res, nil
		case err := <-ch:
			return nil, err
		case <-vm.timer.C:
			r.Interrupt("timeout")
		}
	}
}
