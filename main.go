package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/jogi1/mvdreader"
	"github.com/robertkrimen/otto"
	"io"
	"io/ioutil"
	"log"
	"os"
)

type Weapon_Stat struct {
	Pickup, Drop, Damage int
}

type Armor_Stat struct {
	Pickup, Damage_Absorbed int
}

type Item_Stat struct {
	Pickup, Drop int
}

type Stats struct {
	Axe, Shotgun, SuperShotgun, NailGun, SuperNailGun, GrenadeLauncher, RocketLauncher, LightningGun Weapon_Stat
	GreenArmor, YellowArmor, RedArmor                                                                Armor_Stat
	MegaHealth, Quad, Pentagram, Ring                                                                Item_Stat
	Kills, Deaths, Suicides, Teamkills                                                               int
}

//go:generate stringer -type=Event_Type
type Event_Type uint

const (
	EPT_Spawn Event_Type = iota
	EPT_Death
	EPT_Suicide
	EPT_Kill
	EPT_Teamkill
	EPT_Pickup
	EPT_Drop
)

// EPT_Spawn, EPT_Death, EPT_Suicide
type Event_Player struct {
	Type          Event_Type
	Player_Number int
}

// EPT_Kill, EPT_Teamkill
type Event_Player_Kill struct {
	Type          Event_Type
	Player_Number int
}

type Event_Player_Item struct {
	Type          Event_Type
	Player_Number int
	Item_Type     mvdreader.IT_TYPE
}

type Event_Player_Stat struct {
	Type          Event_Type
	Player_Number int
	Stat          mvdreader.STAT_TYPE
	Amount        int
}

type Parser struct {
	scriptname         string
	mvd                mvdreader.Mvd
	vm                 *otto.Otto
	vm_finish_function *otto.Value
	vm_frame_function  *otto.Value
	events             []interface{}
	ascii_table        []rune
	stats              [32]Stats
	filename           string
	output_file        *os.File
}

type JsonDump struct {
	Mvd      *mvdreader.Mvd
	Stats    [32]Stats
	Filename string
}

func (parser *Parser) clear() {
	parser.events = nil
	parser.stats = [32]Stats{}
}

func main() {
	var parser Parser
	var logger *log.Logger

	debug_file := flag.String("debug_file", "stdout", "debug output target")
	debug := flag.Bool("debug", false, "debug output enabled")
	output_script := flag.String("output_script", "data/default.js", "script to run")
	ascii_table_file := flag.String("ascii_table", "data/ascii.table", "ascii translation table file")
	json_dump := flag.Bool("json_dump", false, "do not run a script, just dump all info as json")
	output_file := flag.String("output_file", "stdout", "output target")

	flag.Parse()

	if *output_file != "stdout" {
		f, err := os.OpenFile(*output_file,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		parser.output_file = f
		defer f.Close()
	}

	if *debug {
		if *debug_file != "stdout" {
			f, err := os.OpenFile(*debug_file,
				os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			defer f.Close()

			logger = log.New(f, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
		} else {
			logger = log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
		}
	}

	if len(flag.Args()) < 1 {
		fmt.Println("no demos supplied")
		os.Exit(1)
	}

	for _, filename := range flag.Args() {
		parser.filename = filename
		parser.clear()

		r, err := zip.OpenReader(filename)
		defer r.Close()
		if err == nil {
			f := r.File[0]
			rc, err := f.Open()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			buf := bytes.NewBuffer(nil)
			io.Copy(buf, rc)

			err, parser.mvd = mvdreader.Load(buf.Bytes(), logger)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		} else {
			read_file, err := ioutil.ReadFile(filename)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			err, parser.mvd = mvdreader.Load(read_file, logger)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

		}

		if *json_dump == false {
			if *output_script == "data/default.js" {
				s, err := Asset("data/default.js")
				err = parser.InitVM(s, "data/default.js")
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			} else {
				script, err := ioutil.ReadFile(*output_script)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				err = parser.InitVM(script, *output_script)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			}
		}

		err = parser.Ascii_Init(*ascii_table_file)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		for {
			err, done := parser.mvd.ParseFrame()
			parser.handlePlayerEvents()

			if err != nil {
				fmt.Println(filename, " - ", err)
				os.Exit(1)
			}
			if *json_dump == false {
				err = parser.VmDemoFrame()
				if err != nil {
					fmt.Println(filename, " - ", err)
					os.Exit(1)
				}
			}
			if done {
				break
			}
			parser.clearPlayerEvents()
		}

		if *json_dump == false {
			err = parser.VmDemoFinished()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		} else {
			var jsonS JsonDump
			jsonS.Filename = parser.filename
			jsonS.Mvd = &parser.mvd
			jsonS.Stats = parser.stats
			js, err := json.MarshalIndent(jsonS, "", "")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if parser.output_file != nil {
				parser.output_file.Write(js)
			} else {
				fmt.Println(string(js))
			}
		}
	}
	os.Exit(0)
}
