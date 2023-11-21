package js

import (
	"github.com/sohaha/zlsgo/ztype"
)

func (vm *VM) RunFile(path string) (result ztype.Type, err error) {
	return vm.RunModuleFile(path)
}

func (vm *VM) RunFileForMethod(path, method string, args ...interface{}) (result ztype.Type, err error) {
	code, err := vm.TranspileFile(path)
	if err != nil {
		return ztype.New(nil), err
	}
	return vm.RunForMethod(code, method, args...)
}

func (vm *VM) RunModuleFile(path string) (result ztype.Type, err error) {
	code, err := vm.TranspileFile(path)
	if err != nil {
		return ztype.New(nil), err
	}

	return vm.RunModule(code)
}
