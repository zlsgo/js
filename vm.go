package js

import (
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"github.com/sohaha/zlsgo/zcache"
	"github.com/sohaha/zlsgo/zlog"
)

type Option struct {
	Dir             string
	DisabledConsole bool
	Timeout         time.Duration
	Args            map[string]interface{}
	Modules         map[string]require.ModuleLoader
	CustomVm        func() *goja.Runtime
	MaxPrograms     uint
}

var log = zlog.New("[JS]")

func init() {
	log.ResetFlags(zlog.BitLevel | zlog.BitTime)
}

func New(opt ...func(*Option)) *VM {
	o := Option{
		Timeout:     time.Minute * 1,
		MaxPrograms: 1 << 10,
	}
	for _, f := range opt {
		f(&o)
	}

	for name := range o.Modules {
		require.RegisterNativeModule(name, o.Modules[name])
	}

	vm := &VM{
		timeout: o.Timeout,
		Programs: zcache.NewFast(func(co *zcache.Option) {
			m := o.MaxPrograms / 10
			if m > 1 {
				co.Bucket = uint16(m)
				co.Cap = 10
				co.LRU2Cap = co.Bucket / 2
			}
		}),
		runtime: sync.Pool{
			New: func() interface{} {
				var opts []require.Option
				if o.Dir != "" {
					opts = append(opts, require.WithLoader(sourceLoader(o.Dir)))
				}

				if o.CustomVm != nil {
					return o.CustomVm()
				}

				vm := goja.New()
				r := require.NewRegistry(opts...)
				r.Enable(vm)

				if !o.DisabledConsole {
					clog := vm.NewObject()
					clog.Set("log", log.Tips)
					clog.Set("warn", log.Warn)
					clog.Set("error", log.Error)
					vm.Set("console", clog)
				}

				for k, v := range o.Args {
					vm.Set(k, v)
				}

				return vm
			},
		},
	}

	return vm
}
