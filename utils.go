package js

import (
	"os"
	"path/filepath"

	"github.com/dop251/goja_nodejs/require"
	"github.com/sohaha/zlsgo/zfile"
)

func sourceLoader(dir string) func(string) ([]byte, error) {
	return func(filename string) ([]byte, error) {
		ext := filepath.Ext(filename)
		data, err := os.ReadFile(zfile.RealPath(filename))
		if err != nil {
			fullPath := zfile.RealPath(dir, true) + filename
			for _, v := range []string{fullPath, fullPath + ".ts"} {
				data, err = os.ReadFile(v)
				if err == nil {
					break
				}
			}

		}

		if err != nil {
			return nil, require.ModuleFileDoesNotExistError
		}

		if ext == ".ts" || ext == "" {
			data, err = Transpile(data, nil)
		}

		return data, err
	}
}
