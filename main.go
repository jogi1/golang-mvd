package main

import (
	"archive/zip"
	"bytes"
	"encoding/hex"
	"fmt"
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
	name      string
	team      string
	frags     int
	userid    int
	spectator bool
	deaths    int
}

type Mvd struct {
	Trace *log.Logger
	Error *log.Logger
	Debug *log.Logger

	file        []byte
	file_offset uint
	filename    string
	frame       uint
	done        bool
	demo        demo
	players     [32]Player
	mapname     string
	hostname    string
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
		fmt.Println("loading ", f.Name, " from zip")
	} else {
		mvd.file, err = ioutil.ReadFile(os.Args[1])
		if err != nil {
			panic(err)
		}
	}

	mvd.Parse("")

	fmt.Printf("{\n")
	fmt.Printf("\t\"map_name\": \"%s\",\n", sanatize_map_name(mvd.mapname))
	fmt.Printf("\t\"map_file\": \"%s\",\n", mvd.demo.modellist[0])
	fmt.Printf("\t\"hostname\": \"%s\",\n", mvd.hostname)
	fmt.Printf("\t\"players\": [\n")
	first := true
	for _, p := range mvd.players {
		if len(p.name) == 0 || p.spectator == true {
			continue
		}
		if first == false {
			fmt.Printf(",\n")
		}
		if first == true {
			first = false
		}
		fmt.Printf("\t\t{\n")
		fmt.Printf("\t\t\"name_sanatized\": \"%s\",\n", sanatize_name(p.name))
		fmt.Printf("\t\t\"name_int\": \"%s\",\n", int_name(p.name))
		fmt.Printf("\t\t\"team_sanatized\": \"%s\",\n", sanatize_name(p.team))
		fmt.Printf("\t\t\"team_int\": \"%s\",\n", int_name(p.team))
		fmt.Printf("\t\t\"frags\": \"%d\",\n", p.frags)
		fmt.Printf("\t\t\"deaths\": \"%d\"\n", p.deaths)
		fmt.Printf("\t\t}")
	}
	fmt.Println("\n")
	fmt.Printf("\t]\n")
	fmt.Printf("}\n")

	os.Exit(1)
}

func (mvd *Mvd) GetInfo(length uint) string {
	return fmt.Sprintf("offset(%v - %x) byte(%v)", mvd.file_offset, mvd.file_offset, hex.EncodeToString(mvd.file[mvd.file_offset:mvd.file_offset+length]))
}
