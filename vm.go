package main

import (
	"fmt"
	"github.com/robertkrimen/otto"
	"strconv"
	"strings"
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

	s := "stats = { "
	for i := STAT_HEALTH; i < STAT_ACTIVEWEAPON; i++ {
		name := strings.TrimPrefix(strings.ToLower((STAT_TYPE(i)).String()), "stat_")
		num := strconv.Itoa(int(i))
		s = s + name + ": " + num + ","
	}
	s = s + "}"
	_, err := vm.Run(s)
	if err != nil {
		mvd.Error.Fatal("setting stats failed with: ", err)
		return err
	}

	s = "items = { "
	for i := IT_SHOTGUN; i <= IT_SIGIL4; i = i << 1 {
		name := strings.TrimPrefix(strings.ToLower((IT_TYPE(i)).String()), "it_")
		num := strconv.Itoa(int(i))
		s = s + name + ": " + num + ","
	}
	s = s + "}"
	_, err = vm.Run(s)
	if err != nil {
		mvd.Error.Fatal("setting items failed with: ", err)
		return err
	}

	s = "event_types = { "
	for i := EPT_Spawn; i <= EPT_Drop; i++ {
		name := strings.ToLower((Event_Type(i)).String())[4:]
		num := strconv.Itoa(int(i))
		s = s + name + ": " + num + ","
	}
	s = s + "}"
	_, err = vm.Run(s)
	if err != nil {
		mvd.Error.Fatal("setting items failed with: ", err)
		return err
	}

	_, err = vm.Run(script)
	if err != nil {
		mvd.Error.Fatal("loading (", name, ") failed with: ", err)
		return err
	}
	frame_function, err := vm.Get("on_frame")
	if err == nil {
		mvd.vm_frame_function = &frame_function
	}
	finish_function, err := vm.Get("on_finish")
	if err == nil {
		mvd.vm_finish_function = &finish_function
	}
	mvd.vm_initialized = true
	return nil
}

func (mvd *Mvd) VmDemoFrame() {
	if mvd.vm_initialized == false {
		return
	}
	if mvd.vm_frame_function == nil {
		return
	}
	//fmt.Println(len(mvd.state.Events))
	_, err := mvd.vm_frame_function.Call(*mvd.vm_frame_function, mvd.state, mvd.state_last_frame, mvd.state.Events)
	if err != nil {
		mvd.Error.Fatal(err)
	}
}

func (mvd *Mvd) VmDemoFinished() {
	if mvd.vm_initialized == false {
		return
	}
	if mvd.vm_finish_function == nil {
		return
	}
	vm := mvd.vm
	err := vm.Set("demo", mvd.state)
	if err != nil {
		mvd.Error.Fatal(err)
	}
	_, err = mvd.vm_finish_function.Call(*mvd.vm_finish_function)
	if err != nil {
		mvd.Error.Fatal(err)
	}
}
