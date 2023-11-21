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
	tt.Equal(int64(1), res.Int64())
}

func TestJSMethod(t *testing.T) {
	tt := zlsgo.NewTest(t)
	vm := js.New()
	res, err := vm.RunForMethod([]byte(`const m = 2;function run(i){return m*i}`), "run", 3)
	tt.NoError(err)
	tt.Equal(int64(6), res.Int64())
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

	for d, b := range map[int]bool{1: false, 111: false, 222: true, 333: true, 444: true, 555: true, 100: false} {
		_, err := vm.Run([]byte(`var m = 1;sleep(` + ztype.ToString(d) + `);m`))

		if b && err == nil {
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
	r1 := res.Get("default").Get("rand").String()
	tt.Log(r1)

	res, err = vm.RunModule(code)
	tt.NoError(err)
	r2 := res.Get("default").Get("rand").String()
	tt.Log(r2)
	tt.Log(res)

	tt.EqualTrue(len(r1) == 8)
	tt.EqualTrue(r1 != r2)

	var wg zsync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Go(func() {
			_, err := vm.RunModule(code)
			tt.NoError(err)
		})
	}
	_ = wg.Wait()
}
