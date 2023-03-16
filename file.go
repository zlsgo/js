package js

func (vm *VM) RunFile(path string) (result interface{}, err error) {
	return vm.RunModuleFile(path)
}

func (vm *VM) RunFileForMethod(path, method string, args ...interface{}) (result interface{}, err error) {
	code, err := vm.TranspileFile(path)
	if err != nil {
		return nil, err
	}
	return vm.RunForMethod(code, method, args...)
}

func (vm *VM) RunModuleFile(path string) (result interface{}, err error) {
	code, err := vm.TranspileFile(path)
	if err != nil {
		return nil, err
	}

	return vm.RunModule(code)
}
