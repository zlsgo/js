package js

import (
	"sync"
	"time"

	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/ztype"

	"github.com/dop251/goja"
	"github.com/dop251/goja/parser"
	"github.com/dop251/goja_nodejs/require"
	"github.com/sohaha/zlsgo/zcache"
	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zlog"
)

type Option struct {
	Args             map[string]interface{}
	Modules          map[string]require.ModuleLoader
	CustomVm         func() *goja.Runtime
	CompilerOptions  ztype.Map
	Dir              string
	Inject           []byte
	Timeout          time.Duration
	MaxPrograms      uint
	DisabledConsole  bool
	ParserSourceMaps bool
}

var log = zlog.New("[JS]")

func init() {
	log.ResetFlags(zlog.BitLevel | zlog.BitTime)
}

func New(opt ...func(*Option)) *VM {
	o := Option{
		Dir:         zfile.RealPath("."),
		Timeout:     time.Minute * 1,
		MaxPrograms: 1 << 10,
	}
	for _, f := range opt {
		f(&o)
	}

	for name := range o.Modules {
		require.RegisterNativeModule(name, o.Modules[name])
	}

	timer := time.NewTimer(o.Timeout)
	timer.Stop()

	vm := &VM{
		timer:  timer,
		option: o,
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

				var vm *goja.Runtime
				if o.CustomVm != nil {
					vm = o.CustomVm()
				} else {
					vm = goja.New()
				}

				var parserOpts []parser.Option
				if !o.ParserSourceMaps {
					parserOpts = append(parserOpts, parser.WithDisableSourceMaps)
				}

				if len(parserOpts) > 0 {
					vm.SetParserOptions(parserOpts...)
				}

				r := require.NewRegistry(opts...)
				r.Enable(vm)

				self := vm.GlobalObject()
				vm.Set("self", self)

				vm.Set("atob", func(code string) string {
					raw, err := zstring.Base64DecodeString(code)
					if err != nil {
						panic(err)
					}
					return raw
				})
				vm.Set("btoa", func(code string) string {
					return zstring.Base64EncodeString(code)
				})

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

				if o.Inject != nil {
					zlog.Debug(vm.RunString(zstring.Bytes2String(o.Inject)))
				}

				return vm
			},
		},
	}

	return vm
}
