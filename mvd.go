package main

import (
	"bytes"
	"encoding/binary"
	//"strconv"
	"fmt"
	"strings"
)

func (mvd *Mvd) Parse(output_type string) {
	//mvd.Debug.Println("parsing ", mvd.filename)

	for {
		if mvd.done == true {
			return
		}
		mvd.Frame()
		if int(mvd.file_offset) > len(mvd.file) {
			//mvd.Debug.Println("parsing finished?")
			break
		}
	}
}

func (mvd *Mvd) Frame() {
	//mvd.Debug.Printf("Frame (%v)", mvd.frame)

	if mvd.ReadFrame() == false {
		mvd.Debug.Panic("somethings wrong")
		return
	}
	mvd.frame++
}

func (mvd *Mvd) ReadFrame() bool {
	for {
		mvd.demotime()
		cmd := DEM_TYPE(mvd.ReadByte() & 7)
		if cmd == dem_cmd {
			mvd.Error.Panic("this is an mvd parser")
		}

		//mvd.Debug.Println("handling cmd", DEM_TYPE(cmd))
		if cmd >= dem_multiple && cmd <= dem_all {
			switch cmd {
			case dem_multiple:
				{
					mvd.demo.last_to = uint(mvd.ReadUint())
					//mvd.Debug.Println("affected players: ", strconv.FormatInt(int64(mvd.demo.last_to), 2), mvd.demo.last_to)
					mvd.demo.last_type = dem_multiple
					break
				}
			case dem_single:
				{
					mvd.demo.last_to = uint(cmd >> 3)
					mvd.demo.last_type = dem_single
					break
				}
			case dem_all:
				{
					//mvd.Debug.Println("dem_all", mvd.file_offset)
					mvd.demo.last_to = 0
					mvd.demo.last_type = dem_all
					break
				}

			case dem_stats:
				{
					//mvd.Debug.Println("dem_stats", cmd, cmd&7, dem_stats, mvd.file_offset, "byte: ", mvd.file[mvd.file_offset])
					mvd.demo.last_to = uint(cmd >> 3)
					mvd.demo.last_type = dem_stats
					break
				}
			}
			cmd = dem_read
		}
		if cmd == dem_set {
			//mvd.Debug.Println("dem_set", mvd.file_offset)
			mvd.demo.outgoing_sequence = mvd.ReadUint()
			mvd.demo.incoming_sequence = mvd.ReadUint()
			//mvd.Debug.Printf("Squence in(%v) out(%v)", mvd.demo.incoming_sequence, mvd.demo.outgoing_sequence)
			continue
		}
		if cmd == dem_read {
			b := mvd.ReadIt(cmd)
			for b == true {
				//mvd.Debug.Println("did we loop?")
				b = mvd.ReadIt(cmd)
			}
			return true
		}
		//mvd.Debug.Println(cmd)
		return false
	}

}

func (mvd *Mvd) ReadNext() bool {
	mvd.demotime()
	return false
}

func (mvd *Mvd) ReadIt(cmd DEM_TYPE) bool {
	current_size := int(mvd.ReadUint())
	if current_size == 0 {
		//mvd.Debug.Println("ReadIt: current size 0 go to next Frame! <----------")
		return false
	}
	old_offset := mvd.file_offset
	mvd.file_offset += uint(current_size)
	//mvd.Debug.Printf("------------- moving ahead %v from (%v) to (%v) filesize: %v", current_size, old_offset, mvd.file_offset, len(mvd.file))
	mvd.MessageParse(Message{size: uint(current_size), data: mvd.file[old_offset:mvd.file_offset]})
	if mvd.demo.last_type == dem_multiple {
		//mvd.Debug.Println("looping")
		return true
	}
	//mvd.Debug.Println("ReadIt: go to next Frame! <----------")
	return false
}

func (mvd *Mvd) usercmd() {
	mvd.ReadBytes(userCmd_t_size)
}

func (mvd *Mvd) demotime() {
	b := mvd.ReadByte()
	mvd.demo.time += float64(b) * 0.0001
	//mvd.Debug.Printf("time (%v)", mvd.demo.time)
}

func (mvd *Mvd) ReadBytes(count uint) *bytes.Buffer {
	//mvd.Debug.Println("------------- READBYTES: ", mvd.GetInfo(count), count)
	b := bytes.NewBuffer(mvd.file[mvd.file_offset : mvd.file_offset+count])
	mvd.file_offset += count
	return b
}

func (mvd *Mvd) ReadByte() byte {
	//mvd.Debug.Println("------------- READBYTE: ", mvd.GetInfo(1))
	b := mvd.file[mvd.file_offset]
	mvd.file_offset += 1
	return b
}

func (mvd *Mvd) ReadInt() int32 {
	var i int32
	err := binary.Read(mvd.ReadBytes(4), binary.LittleEndian, &i)
	if err != nil {
		mvd.Error.Fatal(err)
	}
	return i
}

func (mvd *Mvd) ReadUint() uint32 {
	var i uint32
	err := binary.Read(mvd.ReadBytes(4), binary.LittleEndian, &i)
	if err != nil {
		mvd.Error.Fatal(err)
	}
	return i
}

func mvdPrint(a ...interface{}) {
}

func sanatize_name(name string) string {
	r := []byte(name)
	for i, ri := range r {
		if ri > 128 {
			r[i] = ri - 128
		}
	}
	return string(r)
}

func int_name(name string) string {
	var b strings.Builder
	r := []byte(name)
	for i, ri := range r {
		if i > 0 {
			fmt.Fprintf(&b, " %d", ri)
		} else {
			fmt.Fprintf(&b, "%d", ri)
		}
	}
	return b.String()
}

func sanatize_map_name(name string) string {
	var b strings.Builder
	r := []byte(name)
	for _, ri := range r {
		if ri == '\n' {
			fmt.Fprintf(&b, "\\n")
		} else {
			fmt.Fprintf(&b, "%s", string(ri))
		}
	}
	return b.String()
}
