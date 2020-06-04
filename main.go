package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"github.com/jogi1/mvdreader"
	"github.com/robertkrimen/otto"
	"io"
	"io/ioutil"
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
}

func main() {
	var err error
	var parser Parser

	/*
		mvd.Debug = log.New(os.Stdout,
			"DEBUG: ",
			log.Ldate|log.Ltime|log.Lshortfile)
	*/

	if len(os.Args) < 2 {
		fmt.Println("no demo supplied")
		os.Exit(1)
	}
	filename := os.Args[1]
	parser.filename = filename

	r, err := zip.OpenReader(filename)
	defer r.Close()
	if err == nil {
		f := r.File[0]
		rc, err := f.Open()
		if err != nil {
			panic(err)
		}
		buf := bytes.NewBuffer(nil)
		io.Copy(buf, rc)

		err, parser.mvd = mvdreader.Load(buf.Bytes())
		if err != nil {
			panic(err)
		}
	} else {
		read_file, err := ioutil.ReadFile(os.Args[1])
		if err != nil {
			panic(err)
		}
		err, parser.mvd = mvdreader.Load(read_file)
		if err != nil {
			panic(err)
		}

	}

	script, err := ioutil.ReadFile("runme.js")
	if err != nil {
		s, err := Asset("data/default.js")
		if err != nil {
			panic(err)
		}
		err = parser.InitVM(s, "data/default.js")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	} else {
		err = parser.InitVM(script, "runme.js")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	parser.Ascii_Init("data/ascii.table")

	for {
		err, done := parser.mvd.ParseFrame()

		parser.handlePlayerEvents()

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = parser.VmDemoFrame()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if done {
			break
		}
		parser.clearPlayerEvents()
	}

	err = parser.VmDemoFinished()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}
