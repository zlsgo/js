package js

import (
	_ "embed"

	"github.com/dop251/goja"
	"github.com/sohaha/zlsgo/zjson"
	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/ztype"
)

// ts v4.9.3.js

//go:embed ts.js
var tsTranspile string

var tsVm = New(func(o *Option) {
	o.CustomVm = func() *goja.Runtime {
		vm := goja.New()
		vm.RunString(tsTranspile)
		vm.Set("DecodeCode", func(code string) (string, error) {
			return zstring.Base64DecodeString(code)
		})
		return vm
	}
})

func (vm *VM) Transpile(code []byte, compilerOptions ztype.Map) ([]byte, error) {
	compiler := `{"strict":false,"target":"ES5","module":"CommonJS"}`
	for k := range compilerOptions {
		zjson.Set(compiler, k, compilerOptions[k])
	}

	// moduleName := "m_" + zstring.Md5Byte(code)
	s := `ts.transpileModule(DecodeCode("` + zstring.Base64EncodeString(zstring.Bytes2String(code)) + `"), {
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
