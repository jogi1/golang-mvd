package main

import (
	"archive/zip"
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/robertkrimen/otto"
	"io"
	"io/ioutil"
	"log"
	"os"
)

type demo struct {
	time                                 float64
	last_to                              uint
	last_type                            DEM_TYPE
	outgoing_sequence, incoming_sequence uint32
	soundlist                            []string
	modellist                            []string
	protocol                             PROTOCOL_VERSION
	fte_pext                             FTE_PROTOCOL_EXTENSION
	fte_pext2                            FTE_PROTOCOL_EXTENSION
	mvd_pext                             MVD_PROTOCOL_EXTENSION
}

type Player struct {
	Name      string
	Team      string
	Frags     int
	Userid    int
	Spectator bool
	Deaths    int
}

type mvd_state struct {
	Players  [32]Player
	Mapname  string
	Mapfile  string
	Hostname string
}

type Mvd struct {
	Trace *log.Logger
	Error *log.Logger
	Debug *log.Logger

	file           []byte
	file_offset    uint
	filename       string
	frame          uint
	done           bool
	demo           demo
	vm             *otto.Otto
	vm_initialized bool

	state mvd_state
}

func main() {
	var mvd Mvd
	var err error

	mvd.Trace = log.New(os.Stderr,
		"TRACE: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	mvd.Error = log.New(os.Stderr,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	mvd.Debug = log.New(os.Stdout,
		"DEBUG: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	if len(os.Args) < 2 {
		mvd.Error.Print("no demo supplied")
		os.Exit(1)
	}
	mvd.filename = os.Args[1]

	r, err := zip.OpenReader(os.Args[1])
	defer r.Close()
	if err == nil {
		f := r.File[0]
		rc, err := f.Open()
		if err != nil {
			panic(err)
		}
		buf := bytes.NewBuffer(nil)
		io.Copy(buf, rc)

		mvd.file = buf.Bytes()
		//fmt.Println("loading ", f.Name, " from zip")
	} else {
		mvd.file, err = ioutil.ReadFile(os.Args[1])
		if err != nil {
			panic(err)
		}
	}

	script, err := ioutil.ReadFile("runme.js")
	if err != nil {
		s, err := Asset("data/default.js")
		if err != nil {
			mvd.Error.Fatal(err)
		}
		mvd.InitVM(s, "default.js")
	} else {
		mvd.InitVM(script, "runme.js")
	}

	mvd.Ascii_Init()

	mvd.Parse("")
	mvd.state.Mapfile = mvd.demo.modellist[0]

	mvd.VmDemoFinished()
	os.Exit(0)
}

func (mvd *Mvd) GetInfo(length uint) string {
	return fmt.Sprintf("offset(%v - %x) byte(%v)", mvd.file_offset, mvd.file_offset, hex.EncodeToString(mvd.file[mvd.file_offset:mvd.file_offset+length]))
}
