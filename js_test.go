package js

import (
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zsync"
)

func TestJS(t *testing.T) {
	tt := zlsgo.NewTest(t)
	js := New()
	res, err := js.Run([]byte(`const m = 1;m`))
	tt.NoError(err)
	tt.Equal(int64(1), res)
}

func TestCommonjs(t *testing.T) {
	tt := zlsgo.NewTest(t)
	js := New()
	code, err := js.Transpile([]byte(`
	const rand = Math.random().toString(36).slice(-8)
	export default {rand}
	`), nil)
	tt.NoError(err)

	res, err := js.RunCommonjs(code)
	tt.NoError(err)
	tt.Log(res)

	res, err = js.RunCommonjs(code)
	tt.NoError(err)
	tt.Log(res)

	var wg zsync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Go(func() {
			_, err := js.RunCommonjs(code)
			tt.NoError(err)
		})
	}
	wg.Wait()
}
