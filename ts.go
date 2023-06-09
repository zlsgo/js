package js

import (
	_ "embed"

	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zjson"
	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/ztype"
)

// ts v4.9.3.js

//go:embed ts.js
var tsTranspile []byte

var tsVm *VM

func init() {
	tsVm = New(func(o *Option) {
		o.Inject = tsTranspile
	})
}

func Transpile(code []byte, compilerOptions ztype.Map) ([]byte, error) {
	compiler := `{"strict":false,"target":"ES5","module":"CommonJS"}`
	for k := range compilerOptions {
		zjson.Set(compiler, k, compilerOptions[k])
	}
	s := `ts.transpileModule(atob("` + zstring.Base64EncodeString(zstring.Bytes2String(code)) + `"), {
		"compilerOptions": ` + compiler + `,
		})`

	r := tsVm.GetRuntime()
	defer tsVm.PutRuntime(r)

	res, err := r.RunString(s)
	if err != nil {
		return nil, err
	}

	return zstring.String2Bytes(ztype.ToString(res.ToObject(r).Get("outputText").Export()) + ""), nil
}

func (vm *VM) TranspileFile(file string) ([]byte, error) {
	code, err := zfile.ReadFile(file)
	if err != nil {
		return nil, err
	}

	return vm.Transpile(code)
}

func (vm *VM) Transpile(code []byte) ([]byte, error) {
	return Transpile(code, vm.option.CompilerOptions)
}
