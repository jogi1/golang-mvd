package main

import (
	"fmt"
	"github.com/robertkrimen/otto"
)

func (mvd *Mvd) InitVM(script []byte, name string) error {
	vm := otto.New()
	mvd.vm = vm

	vm.Set("sanatize", func(in string) string {
		return sanatize_name(in)
	})

	/*
		vm.Set("unicode", func(in string) string {
			return unicode_string(in)
		})
	*/

	vm.Set("sanatize_escapes", func(in string) string {
		return sanatize_map_name(in)
	})

	vm.Set("convert_int", func(in string) string {
		return int_name(in)
	})

	vm.Set("print", func(call otto.FunctionCall) otto.Value {
		m := ""
		for _, v := range call.ArgumentList {
			m = fmt.Sprintf("%s%s", m, v.String())
		}
		fmt.Printf("%s", m)
		return otto.Value{}
	})

	_, err := vm.Run(script)
	if err != nil {
		mvd.Error.Fatal("loading (", name, ") failed with: ", err)
		return err
	}
	mvd.vm_initialized = true
	return nil
}

func (mvd *Mvd) VmDemoFinished() {
	if mvd.vm_initialized == false {
		return
	}
	vm := mvd.vm
	err := vm.Set("demo", mvd.state)
	if err != nil {
		mvd.Error.Fatal(err)
	}
	if _, err := vm.Get("on_finish"); err == nil {
		_, err := vm.Run("on_finish()")
		if err != nil {
			mvd.Error.Fatal(err)
		}
	}
}
