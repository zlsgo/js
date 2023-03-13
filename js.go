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
	timeout  time.Duration
}

func (vm *VM) GetRuntime() *goja.Runtime {
	return vm.runtime.Get().(*goja.Runtime)
}

func (vm *VM) PutRuntime(r *goja.Runtime) {
	vm.runtime.Put(r)
}

func (vm *VM) getProgram(code []byte, isExports bool) (p *goja.Program, err error) {
	name := zstring.Md5Byte(code)
	if isExports {
		name = "e_" + name
	}
	data, ok := vm.Programs.ProvideGet(name, func() (interface{}, bool) {
		var p *goja.Program
		codeStr := zstring.Bytes2String(code)
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
		return nil, errors.New("compile not existent")
	}

	return data.(*goja.Program), nil
}

func (vm *VM) Run(code []byte) (result interface{}, err error) {
	p, err := vm.getProgram(code, false)
	if err != nil {
		return nil, err
	}

	return vm.RunProgram(p)
}

func (vm *VM) RunProgram(p *goja.Program, rendered ...func(*goja.Runtime) (interface{}, error)) (result interface{}, err error) {
	if p == nil {
		return nil, errors.New("program is nil")
	}
	r := vm.GetRuntime()
	defer vm.PutRuntime(r)

	ch := make(chan struct{})

	var res goja.Value

	go func() {
		err = zerror.TryCatch(func() error {
			res, err = r.RunProgram(p)
			return err
		})
		ch <- struct{}{}
	}()

	select {
	case <-ch:
		if err != nil {
			return
		}

		if len(rendered) == 0 {
			result = res.Export()
			return
		}

		for i := range rendered {
			result, err = rendered[i](r)
			if err != nil {
				return
			}
		}

		return
	case <-time.After(vm.timeout):
		r.Interrupt("timeout")
		r.ClearInterrupt()
		return nil, errors.New("timeout: " + vm.timeout.String())
	}
}
