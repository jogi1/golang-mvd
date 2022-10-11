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
    "github.com/bodgit/sevenzip"
)

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
	compressed_filename string
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
    Mvd                *mvdreader.Mvd `json:"mvd"`
    Stats              map[int]*Stats `json:"stats"`
    Filename           string `json:"filename"`
    CompressedFilename string`json:"compressed_filename"`
    FragMessages       []*fragfile.FragMessage `json:"frag_messages"`
    Players            map[int]mvdreader.Player `json:"frag_players"`
    ModParserInfo      interface{} `json:"mod_parser_information"`
    PlayersAccumulated []*ParserPlayer `json:"players_accumulated"`

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
	WpsParserEnabled     bool // ktx wps parsing
    StatsEnabled         bool // player stats tracking
    StatsStrengthEnabled bool // player strength calculation
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
	p.filename = filename
	read_file, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return p.LoadByte(read_file, filename, "")
}

func (p *Parser) LoadByte(demo []byte, filename string, compressed_filename string) error {
	p.filename = filename
	p.compressed_filename = compressed_filename
	mvd, err := mvdreader.Load(demo, p.logger, nil)
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

		done, err := p.mvd.ParseFrame()
		if err != nil && !done {
			return err
		}
		if !p.Flags.PlayerEventsDisabled {
			p.handlePlayerEvents()
			p.handlePlayerDisconnects()
		}

		if p.Flags.WpsParserEnabled {
			for _, message := range p.mvd.State.StuffText {
				p.WpsParserParse(message.String)
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
				fm, err := p.fragfile.ParseMessage(message.Message.String)
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
	aggregatePlayerInfo := flag.Bool("aggregate_player_info", false, "aggregate all possible info sources in a player")
	wpsParserEnabled := flag.Bool("wps_parser_enabled", false, "enable /wps parsing, very expensive")

	statsEnabled := flag.Bool("stats_enabled", false, "enable player stat tracking")
	statsStrengthEnabled := flag.Bool("stats_strength_enabled", false, "enable player strength calculation")

	flag.Parse()

	var parserFlags ParserFlags
	parserFlags.StatsEnabled = *statsEnabled
	parserFlags.StatsStrengthEnabled = *statsStrengthEnabled
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
		zip_reader, err := zip.OpenReader(filename)
		if err == nil {
			defer zip_reader.Close()
            for _, f := range zip_reader.File {
                rc, err := f.Open()
                if err != nil {
                    fmt.Println(err)
                    os.Exit(1)
                }
                buf := bytes.NewBuffer(nil)
                io.Copy(buf, rc)

                err = parse_demo(parser, buf.Bytes(), f.FileInfo().Name(), filename)
                if err != nil {
                    fmt.Println(err)
                    os.Exit(1)
                }
            }
            os.Exit(0)
        }

        sz_reader, err := sevenzip.OpenReader(filename)
        if err == nil {
            defer sz_reader.Close()
            for _, f := range sz_reader.File {
                file_reader, err := f.Open()
                if err != nil {
                    fmt.Println(err)
                    os.Exit(1)
                }
                buf := bytes.NewBuffer(nil)
                io.Copy(buf, file_reader)

                err = parse_demo(parser, buf.Bytes(), f.FileInfo().Name(), filename)
                if err != nil {
                    fmt.Println(err)
                    os.Exit(1)
                }
            }
            os.Exit(0)
        }

        read_file, err := ioutil.ReadFile(filename)
        if err != nil {
            fmt.Println(err)
            os.Exit(1)
        }
        err = parse_demo(parser, read_file, filename, "")
        if err != nil {
            fmt.Println(err)
            os.Exit(1)
        }

    }
    os.Exit(0)
}

func parse_demo(parser *Parser , demo []byte, file_name string, compressed_filename string) error {
    err := parser.LoadByte(demo, file_name, compressed_filename)
    if err != nil {
        return err;
    }

    err = parser.Parse()
    if err != nil {
        return err;
    }

    if parser.vm == nil || parser.Flags.JsonDump {
        var jsonS JsonDump
        jsonS.Filename = parser.filename
        jsonS.CompressedFilename = parser.compressed_filename
        jsonS.Mvd = &parser.mvd
        jsonS.Stats = parser.stats
        jsonS.FragMessages = parser.fragmessages
        jsonS.Players = parser.players
        jsonS.PlayersAccumulated = parser.PlayersFrameCurrent

        if parser.mod_parser != nil {
            jsonS.ModParserInfo = parser.mod_parser.State
        }
        js, err := json.MarshalIndent(jsonS, "", "\t")
        if err != nil {
            return err
        }
        if parser.output_file != nil {
            parser.output_file.Write(js)
            parser.output_file.Close()
        } else {
            fmt.Println(string(js))
        }
    }
    return nil
}
