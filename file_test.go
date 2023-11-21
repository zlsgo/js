package js_test

import (
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/zlsgo/js"
)

func TestFile(t *testing.T) {
	tt := zlsgo.NewTest(t)

	vm := js.New(func(o *js.Option) {
		o.Dir = "./testdata"
	})

	res, err := vm.RunFile("./testdata/test.js")
	tt.NoError(err)
	tt.Equal(666, res.Get("default.ok").Int())
}

func TestFileForMethod(t *testing.T) {
	tt := zlsgo.NewTest(t)

	vm := js.New()

	res, err := vm.RunFileForMethod("./testdata/method.ts", "run", "ts")
	tt.NoError(err)
	tt.Equal("hello ts", res.String())
}
