package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jogi1/mvdreader"
	"github.com/robertkrimen/otto"
)

func (parser *Parser) OutputFileSet(name string) error {
    if parser.output_file != nil{
        parser.output_file.Close()
        parser.output_file = nil
    }

    f, err := os.OpenFile(name,
        os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
    if err != nil {
        return err
    }
    parser.output_file = f
    return nil
}

func (parser *Parser) InitVM(script []byte, name string) error {
	vm := otto.New()
	parser.vm = vm

    vm.Set("__SetOutputFile", func (in string) error {
        return parser.OutputFileSet(in)
    })

	vm.Set("__StringSanatize", func(in string) string {
		return parser.SanatizeName(in)
	})

	vm.Set("__StringSanatizeEscapes", func(in string) string {
		return sanatize_map_name(in)
	})

	vm.Set("__StringConvertInt", func(in string) string {
		return int_name(in)
	})

	vm.Set("__print", func(call otto.FunctionCall) otto.Value {
		m := ""
		for _, v := range call.ArgumentList {
			m = fmt.Sprintf("%s%s", m, v.String())
		}
		if parser.output_file == nil {
			fmt.Printf("%s", m)
		} else {
			b := []byte(m)
			parser.output_file.Write(b)
		}
		return otto.Value{}
	})

	vm.Set("__println", func(call otto.FunctionCall) otto.Value {
		m := ""
		for _, v := range call.ArgumentList {
			m = fmt.Sprintf("%s%s", m, v.String())
		}
		if parser.output_file == nil {
			fmt.Printf("%s\n", m)
		} else {
			b := []byte(m)
			parser.output_file.Write(b)
			parser.output_file.Write([]byte{10})
		}
		return otto.Value{}
	})
	name_num := "stats = { "
	num_name := "stats_name = { "
	for i := mvdreader.STAT_HEALTH; i < mvdreader.STAT_ACTIVEWEAPON; i++ {
		name := strings.TrimPrefix(strings.ToLower((mvdreader.STAT_TYPE(i)).String()), "stat_")
		num := strconv.Itoa(int(i))
		name_num = name_num + name + ": " + num + ","
		num_name = num_name + num + ": \"" + name + "\","
	}
	name_num = name_num + "};"
	num_name = num_name + "}"
	s := name_num + num_name
    _, err := vm.Run(s)
	if err != nil {
		return err
	}

	name_num = "items = { "
	num_name = "items_name = { "
	for i := mvdreader.IT_SHOTGUN; i <= mvdreader.IT_SIGIL4; i = i << 1 {
		name := strings.TrimPrefix(strings.ToLower((mvdreader.IT_TYPE(i)).String()), "it_")
		num := strconv.Itoa(int(i))
		name_num = name_num + name + ": " + num + ","
		num_name = num_name + num + ": \"" + name + "\","
	}
	name_num = name_num + "};"
	num_name = num_name + "}"
	s = name_num + num_name
	_, err = vm.Run(s)
	if err != nil {
		return err
	}

	name_num = "event_types = { "
	num_name = "event_types_name = { "
	for i := EPT_Spawn; i <= EPT_Drop; i++ {
		name := strings.ToLower((Event_Type(i)).String())[4:]
		num := strconv.Itoa(int(i))
		name_num = name_num + name + ": " + num + ","
		num_name = num_name + num + ": \"" + name + "\","
	}
	name_num = name_num + "};"
	num_name = num_name + "}"
	s = name_num + num_name
	_, err = vm.Run(s)
	if err != nil {
		return err
	}

	_, err = vm.Run(script)
	if err != nil {
		return err
	}

	frame_function, err := vm.Get("on_frame")
	if err == nil {
		if frame_function != otto.UndefinedValue() {
			parser.vm_frame_function = &frame_function
		}
	}
	finish_function, err := vm.Get("on_finish")
	if err == nil {
		if finish_function != otto.UndefinedValue() {
			parser.vm_finish_function = &finish_function
		}
	}
	init_function, err := vm.Get("on_init")
	if err == nil {
		if init_function != otto.UndefinedValue() {
			parser.vm_init_function = &init_function
		}
	}
	return nil
}

func (parser *Parser) VmDemoInit() error {
	if parser.vm_init_function == nil {
		return nil
	}

	_, err := parser.vm_init_function.Call(
		*parser.vm_init_function,
		parser.filename,
	)

	if err != nil {
		return err
	}
	return nil
}

func (parser *Parser) VmDemoFrame() error {
	if parser.vm_frame_function == nil {
		return nil
	}
	_, err := parser.vm_frame_function.Call(
		*parser.vm_frame_function,
		parser.mvd.Server,
		parser.mvd.State,
		parser.mvd.State_last_frame,
		parser.events,
		parser.stats,
		parser.fragmessagesFrame,
		parser.PlayersFrameCurrent,
	)
	if err != nil {
		return err
	}
	return nil
}

func (parser *Parser) VmDemoFinished() error {
	if parser.vm_finish_function == nil {
		return nil
	}
	_, err := parser.vm_finish_function.Call(
		*parser.vm_finish_function,
		parser.mvd.Server,
		parser.mvd.State,
		parser.stats,
		parser.fragmessages,
		parser.PlayersFrameCurrent,
		parser.mod_parser.State,
	)
	if err != nil {
		return err
	}
	return nil
}
