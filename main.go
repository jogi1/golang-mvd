package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/jogi1/golang-fragfile"
	"github.com/jogi1/mvdreader"
	"github.com/robertkrimen/otto"
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
	Flags               ParserFlags
	debug               bool
	scriptname          string
	mvd                 mvdreader.Mvd
	vm                  *otto.Otto
	vm_finish_function  *otto.Value
	vm_frame_function   *otto.Value
	vm_init_function    *otto.Value
	events              []interface{}
	ascii_table         []rune
	stats               map[int]*Stats
	filename            string
	output_file         *os.File
	fragfile            *fragfile.Fragfile
	fragmessagesFrame   []*fragfile.FragMessage
	fragmessages        []*fragfile.FragMessage
	players             map[int]mvdreader.Player
	mod_parser          *Mod
	logger              *log.Logger
	PlayersFrameCurrent []*ParserPlayer
	PlayersFrameLast    []*ParserPlayer
}

type JsonDump struct {
	Mvd                *mvdreader.Mvd
	Stats              map[int]*Stats
	Filename           string
	Fragmessages       []*fragfile.FragMessage
	Players            map[int]mvdreader.Player
	ModParserInfo      interface{}
	PlayersAccumulated []*ParserPlayer
}

func (parser *Parser) init() {
	parser.players = make(map[int]mvdreader.Player)
	parser.stats = make(map[int]*Stats)
}

func (parser *Parser) clear() {
	parser.events = nil
	parser.stats = make(map[int]*Stats)
	parser.players = make(map[int]mvdreader.Player)
}

type ParserFlags struct {
	AggregatePlayerInfo  bool    // aggregate all possible info into players
	Debug                bool    // debug prints
	DebugFile            *string // output file (filename|stdout|stderr)
	PlayerEventsDisabled bool    // disable player event generation
	ModParserDisabled    bool    // disable mod stat parser
	Script               *string // script to be run in the vm
	ScriptBuffer         *[]byte // script to be run in the vm
	ScriptBufferName     *string // script to be run in the vm
	AsciiTable           *string // ascii table to be used
	AsciiTableFile       *string // ascii table file to be used
	FragFile             *string // fragfile
	FragFileBuffer       *[]byte // fragfile
	JsonDump             bool
	OutputFile           *string // if not set defaults to stdout
	Logger               *log.Logger
	RetainFrames         bool // retain parser frames
	WpsParserEnabled     bool // retain parser frames
}

func ParserNew(flags ParserFlags) (*Parser, error) {
	p := new(Parser)
	p.init()
	p.Flags = flags

	if p.Flags.OutputFile != nil {
		if *p.Flags.OutputFile != "stdout" {
			f, err := os.OpenFile(*p.Flags.OutputFile,
				os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
			if err != nil {
				return nil, err
			}
			p.output_file = f
		}
	}

	// buffers take presidence
	if p.Flags.FragFileBuffer != nil {
		fragfilep, err := fragfile.FragfileLoadByte(*p.Flags.FragFileBuffer)
		if err != nil {
			return nil, err
		}
		p.fragfile = fragfilep
	} else if p.Flags.FragFile != nil {
		fragfilep, err := fragfile.FragfileLoadFile(*p.Flags.FragFile)
		if err != nil {
			return nil, err
		}
		p.fragfile = fragfilep
	}

	if p.Flags.Logger != nil {
		p.logger = p.Flags.Logger
	} else if p.Flags.Debug {
		if *p.Flags.DebugFile != "stdout" {
			f, err := os.OpenFile(*p.Flags.DebugFile,
				os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			p.logger = log.New(f, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
		} else {
			p.logger = log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
		}
	}

	if p.Flags.AsciiTable != nil {
		err := p.asciiInitString(*p.Flags.AsciiTable)
		if err != nil {
			return nil, err
		}
	} else if p.Flags.AsciiTableFile != nil {
		err := p.asciiInitFile(*p.Flags.AsciiTableFile)
		if err != nil {
			return nil, err
		}
	}

	if !p.Flags.ModParserDisabled {
		p.ModInfoParserInit()
	}

	if p.Flags.ScriptBuffer != nil {
		name := "__loaded_from_buffer__"
		if p.Flags.ScriptBufferName != nil {
			name = *p.Flags.ScriptBufferName
		}
		err := p.InitVM(*p.Flags.ScriptBuffer, name)
		if err != nil {
			return nil, err
		}
	} else if p.Flags.Script != nil {
		script, err := ioutil.ReadFile(*p.Flags.Script)
		if err != nil {
			return nil, err
		}
		err = p.InitVM(script, *p.Flags.Script)
		if err != nil {
			return nil, err
		}
	}
	return p, nil
}

func (p *Parser) LoadFile(filename string) error {
	read_file, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return p.LoadByte(read_file, filename)
}

func (p *Parser) LoadByte(demo []byte, filename string) error {
	err, mvd := mvdreader.Load(demo, p.logger)
	if err != nil {
		return err
	}
	p.mvd = mvd
	return err
}

func (p *Parser) Parse() error {
	if p.vm != nil {
		err := p.VmDemoInit()
		if err != nil {
			return err
		}
	}

	for {
		err := p.PlayersNewFrame()
		if err != nil {
			return err
		}

		err, done := p.mvd.ParseFrame()
		if err != nil {
			return err
		}
		if !p.Flags.PlayerEventsDisabled {
			p.handlePlayerEvents()
			p.handlePlayerDisconnects()
		}

		if p.Flags.WpsParserEnabled {
			for _, message := range p.mvd.State.StuffText {
				p.WpsParserParse(message)
			}
		}

		if !p.Flags.ModParserDisabled {
			if p.mod_parser == nil {
				p.ModInfoParserFind()
			} else {
				err := p.mod_parser.Frame(p)
				if err != nil {
					return fmt.Errorf("mod parser frame: %s - %s error: %s", p.mod_parser.Name, p.mod_parser.Version, err)
				}
			}
		}

		if p.fragfile != nil {
			for _, message := range p.mvd.State.Messages {
				fm, err := p.fragfile.ParseMessage(message.Message)
				if err != nil {
					return err
				}
				if fm != nil {
					p.fragmessages = append(p.fragmessages, fm)
					p.fragmessagesFrame = append(p.fragmessagesFrame, fm)
				}
			}
		}
		if p.vm != nil {
			err = p.VmDemoFrame()
			if err != nil {
				return err
			}
		}

		p.clearPlayerEvents()
		p.fragmessagesFrame = nil

		if done {
			break
		}
	}

	if p.mod_parser != nil {
		err := p.mod_parser.End(p)
		if err != nil {
			return fmt.Errorf(
				"mod parser end: %s - %s error: %s",
				p.mod_parser.Name,
				p.mod_parser.Version,
				err,
			)
		}
	}

	if p.vm != nil {
		err := p.VmDemoFinished()
		return err
	}
	return nil
}

func main() {
	debug_file := flag.String("debug_file", "stdout", "debug output target")
	debug := flag.Bool("debug", false, "debug output enabled")
	player_events_disabled := flag.Bool("player_events_disabled", false, "disable player events")
	script := flag.String("script", "", "script to run")
	ascii_table_file := flag.String("ascii_table", "data/ascii.table", "ascii translation table file")
	fragfile := flag.String("fragfile", "", "fragfile to use for parsing frag messages")
	output_file := flag.String("output_file", "stdout", "output target")
	retain_frames := flag.Bool("retain_frames", false, "retain parser frames")
	aggregatePlayerInfo := flag.Bool("aggregate_player_info", true, "aggregate all possible info sources in a player")
	wpsParserEnabled := flag.Bool("wps_parser_enabled", true, "enable /wps parsing, very expensive")

	flag.Parse()

	var parserFlags ParserFlags
	parserFlags.WpsParserEnabled = *wpsParserEnabled
	parserFlags.AggregatePlayerInfo = *aggregatePlayerInfo
	parserFlags.Debug = *debug
	parserFlags.DebugFile = debug_file
	parserFlags.RetainFrames = *retain_frames
	parserFlags.PlayerEventsDisabled = *player_events_disabled
	if len(*script) > 0 {
		parserFlags.Script = script
	}
	parserFlags.AsciiTableFile = ascii_table_file
	if len(*fragfile) > 0 {
		parserFlags.FragFile = fragfile
	}
	parserFlags.OutputFile = output_file

	parser, err := ParserNew(parserFlags)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if len(flag.Args()) < 1 {
		fmt.Println("no demos supplied")
		os.Exit(1)
	}

	for _, filename := range flag.Args() {
		// read zip files
		r, err := zip.OpenReader(filename)
		if err == nil {
			defer r.Close()
			f := r.File[0]
			rc, err := f.Open()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			buf := bytes.NewBuffer(nil)
			io.Copy(buf, rc)

			err = parser.LoadByte(buf.Bytes(), f.FileInfo().Name())
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
			err = parser.LoadByte(read_file, filename)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

		}

		err = parser.Parse()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if parser.vm == nil || parser.Flags.JsonDump {
			var jsonS JsonDump
			jsonS.Filename = parser.filename
			jsonS.Mvd = &parser.mvd
			jsonS.Stats = parser.stats
			jsonS.Fragmessages = parser.fragmessages
			jsonS.Players = parser.players
			jsonS.PlayersAccumulated = parser.PlayersFrameCurrent

			if parser.mod_parser != nil {
				jsonS.ModParserInfo = parser.mod_parser.State
			}
			js, err := json.MarshalIndent(jsonS, "", "\t")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if parser.output_file != nil {
				parser.output_file.Write(js)
				parser.output_file.Close()
			} else {
				fmt.Println(string(js))
			}
		}
	}
	os.Exit(0)
}
