package js_test

import (
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zsync"
	"github.com/sohaha/zlsgo/ztype"
	"github.com/zlsgo/js"
)

func TestJS(t *testing.T) {
	tt := zlsgo.NewTest(t)
	vm := js.New()
	res, err := vm.Run([]byte(`const m = 1;m`))
	tt.NoError(err)
	tt.Equal(int64(1), res)
}

func TestJSMethod(t *testing.T) {
	tt := zlsgo.NewTest(t)
	vm := js.New()
	res, err := vm.RunForMethod([]byte(`const m = 2;function run(i){return m*i}`), "run", 3)
	tt.NoError(err)
	tt.Equal(int64(6), res)
}

func TestTimeout(t *testing.T) {
	tt := zlsgo.NewTest(t)
	vm := js.New(func(option *js.Option) {
		option.Timeout = time.Second / 5
		option.Args = map[string]interface{}{
			"sleep": func(d int) {
				time.Sleep(time.Millisecond * time.Duration(d))
			},
		}
	})

	for d, b := range map[int]bool{1: false, 111: false, 222: true, 333: true, 444: true, 555: true} {
		_, err := vm.Run([]byte(`var m = 1;sleep(` + ztype.ToString(d) + `);m`))

		if b && tt.NoError(err) {
			tt.Fatal(d)
		}
	}
}

func TestModule(t *testing.T) {
	tt := zlsgo.NewTest(t)
	vm := js.New()
	code, err := vm.Transpile([]byte(`
	const rand = Math.random().toString(36).slice(-8)
	export default {rand}
	`))
	tt.NoError(err)

	res, err := vm.RunModule(code)
	tt.NoError(err)
	tt.Log(res)

	res, err = vm.RunModule(code)
	tt.NoError(err)
	tt.Log(res)

	var wg zsync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Go(func() {
			_, err := vm.RunModule(code)
			tt.NoError(err)
		})
	}
	_ = wg.Wait()
}
