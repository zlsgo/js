package js

import (
	"os"
	"path/filepath"

	"github.com/dop251/goja_nodejs/require"
	"github.com/sohaha/zlsgo/zfile"
)

func sourceLoader(dir string) func(string) ([]byte, error) {
	return func(filename string) ([]byte, error) {
		data, err := os.ReadFile(zfile.RealPath(filename))
		if err != nil && !filepath.IsAbs(filename) {
			data, err = os.ReadFile(zfile.RealPath(dir, true) + filename)
		}

		if err != nil {
			return nil, require.ModuleFileDoesNotExistError
		}
		return data, err
	}
}
