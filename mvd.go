package main

import (
	"bytes"
	"encoding/binary"
	"strconv"
)

func (mvd *Mvd) Parse(output_type string) {
	if mvd.debug {
		mvd.Debug.Println("parsing ", mvd.filename)
	}
	for {
		if mvd.done == true {
			return
		}
		mvd.Frame()
		if int(mvd.file_offset) > len(mvd.file) {
			if mvd.debug {
				mvd.Debug.Println("parsing finished?")
			}
			break
		}
		mvd.VmDemoFrame()
	}
}

func (mvd *Mvd) Frame() {
	if mvd.debug {
		mvd.Debug.Printf("Frame (%v)", mvd.frame)
	}
	mvd.state_last_frame = mvd_state(mvd.state)
	mvd.state_last_frame.Events = mvd.state.Events
	mvd.state.Events = make([]interface{}, 0)

	if mvd.ReadFrame() == false {
		if mvd.debug {
			mvd.Debug.Panic("somethings wrong")
		}
		return
	}
	mvd.HandlePlayerEvents()
	mvd.frame++
}

func (mvd *Mvd) ReadFrame() bool {
	for {
		mvd.demotime()
		mvd.state.Time = mvd.demo.time
		cmd := mvd.ReadByte()
		msg_type := DEM_TYPE(cmd & 7)
		if msg_type == dem_cmd {
			mvd.Error.Panic("this is an mvd parser")
		}

		if mvd.debug {
			mvd.Debug.Println("handling cmd", DEM_TYPE(cmd))
		}
		if msg_type >= dem_multiple && msg_type <= dem_all {
			switch msg_type {
			case dem_multiple:
				{
					i := mvd.ReadInt()
					//mvd.demo.last_to = uint(mvd.ReadUint())
					mvd.demo.last_to = uint(i)
					if mvd.debug {
						mvd.Debug.Println("affected players: ", strconv.FormatInt(int64(mvd.demo.last_to), 2), mvd.demo.last_to)
					}
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
					if mvd.debug {
						mvd.Debug.Println("dem_all", mvd.file_offset)
					}
					mvd.demo.last_to = 0
					mvd.demo.last_type = dem_all
					break
				}

			case dem_stats:
				{
					if mvd.debug {
						mvd.Debug.Println("dem_stats", cmd, cmd&7, dem_stats, mvd.file_offset, "byte: ", mvd.file[mvd.file_offset])
					}
					mvd.demo.last_to = uint(cmd >> 3)
					mvd.demo.last_type = dem_stats
					break
				}
			}
			msg_type = dem_read
		}
		if msg_type == dem_set {
			if mvd.debug {
				mvd.Debug.Println("dem_set", mvd.file_offset)
			}
			mvd.demo.outgoing_sequence = mvd.ReadUint()
			mvd.demo.incoming_sequence = mvd.ReadUint()
			if mvd.debug {
				mvd.Debug.Printf("Squence in(%v) out(%v)", mvd.demo.incoming_sequence, mvd.demo.outgoing_sequence)
			}
			continue
		}
		if msg_type == dem_read {
			b := mvd.ReadIt(msg_type)
			for b == true {
				if mvd.debug {
					mvd.Debug.Println("did we loop?")
				}
				b = mvd.ReadIt(msg_type)
			}
			return true
		}
		if mvd.debug {
			mvd.Debug.Println(cmd)
		}
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
		if mvd.debug {
			mvd.Debug.Println("ReadIt: current size 0 go to next Frame! <----------")
		}
		return false
	}
	old_offset := mvd.file_offset
	mvd.file_offset += uint(current_size)
	if mvd.debug {
		mvd.Debug.Printf("------------- moving ahead %v from (%v) to (%v) filesize: %v", current_size, old_offset, mvd.file_offset, len(mvd.file))
	}
	mvd.MessageParse(Message{size: uint(current_size), data: mvd.file[old_offset:mvd.file_offset]})
	if mvd.demo.last_type == dem_multiple {
		if mvd.debug {
			mvd.Debug.Println("looping")
		}
		return true
	}
	if mvd.debug {
		mvd.Debug.Println("ReadIt: go to next Frame! <----------")
	}
	return false
}

func (mvd *Mvd) usercmd() {
	mvd.ReadBytes(userCmd_t_size)
}

func (mvd *Mvd) demotime() {
	b := mvd.ReadByte()
	mvd.demo.time += float64(b) * 0.0001
	if mvd.debug {
		mvd.Debug.Printf("time (%v)", mvd.demo.time)
	}
}

func (mvd *Mvd) ReadBytes(count uint) *bytes.Buffer {
	if mvd.debug {
		mvd.Debug.Println("------------- READBYTES: ", mvd.GetInfo(count), count)
	}
	b := bytes.NewBuffer(mvd.file[mvd.file_offset : mvd.file_offset+count])
	mvd.file_offset += count
	return b
}

func (mvd *Mvd) ReadByte() byte {
	if mvd.debug {
		mvd.Debug.Println("------------- READBYTE: ", mvd.GetInfo(1))
	}
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
