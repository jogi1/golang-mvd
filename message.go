package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
	"strings"
)

type Message struct {
	size   uint
	data   []byte
	offset uint
	mvd    *Mvd
}

func (mvd *Mvd) MessageParse(message Message) {
	message.mvd = mvd
	for {
		if mvd.done == true {
			return
		}
		msg_type := SVC_TYPE(message.ReadByte())

		mvdPrint("handling: ", msg_type)
		m := reflect.ValueOf(&message).MethodByName(strings.Title(fmt.Sprintf("%s", msg_type)))

		if m.IsValid() == true {
			m.Call([]reflect.Value{reflect.ValueOf(mvd)})
		} else {
			fmt.Println(msg_type)
			mvd.Error.Fatalf("--> %#v", m)
		}
		if message.offset >= message.size {
			mvdPrint("message ended?")
			return
		}
		if mvd.done {
			return
		}
	}
	if message.offset != message.size {
		mvd.Error.Fatalln("did not read message fully ", message.offset, message.size)
	}
}

func (message *Message) Svc_serverdata(mvd *Mvd) {
	for {
		message.mvd.demo.protocol = PROTOCOL_VERSION(message.ReadLong())
		protocol := message.mvd.demo.protocol
		mvdPrint("protocol version: ", protocol)

		if protocol == protocol_fte2 {
			message.mvd.demo.fte_pext2 = FTE_PROTOCOL_EXTENSION(message.ReadLong())
			mvdPrint("fte protocol extensions: ", message.mvd.demo.fte_pext)
			continue
		}

		if protocol == protocol_fte {
			message.mvd.demo.fte_pext = FTE_PROTOCOL_EXTENSION(message.ReadLong())
			mvdPrint("fte protocol extensions: ", message.mvd.demo.fte_pext)
			continue
		}

		if protocol == protocol_mvd1 {
			message.mvd.demo.mvd_pext = MVD_PROTOCOL_EXTENSION(message.ReadLong())
			mvdPrint("mvd protocol extensions: ", message.mvd.demo.fte_pext)
			continue
		}
		if protocol == protocol_standard {
			break
		}
	}

	mvdPrint("server count: ", message.ReadLong())
	mvdPrint("gamedir: ", message.ReadString())
	mvdPrint("demotime: ", message.ReadFloat())
	mvd.state.Mapname = message.ReadString()
	for i := 0; i < 10; i++ {
		//fmt.Printf("movevar(%v): %v\n", i, message.ReadFloat())
		message.ReadFloat()
	}
}

/*
func (message *Message) Svc_bad(mvd *Mvd) {
}
*/

func (message *Message) Svc_cdtrack(mvd *Mvd) {
	message.ReadByte()
}

func (message *Message) Svc_stufftext(mvd *Mvd) {
	mvdPrint(message.ReadString())
}

func (message *Message) Svc_soundlist(mvd *Mvd) {
	message.ReadByte() // those are some indexes
	for {
		s := message.ReadString()
		message.mvd.demo.soundlist = append(message.mvd.demo.soundlist, s)
		if len(s) == 0 {
			break
		}
	}
	message.ReadByte() // some more indexes
}

func (message *Message) Svc_modellist(mvd *Mvd) {
	message.ReadByte() // those are some indexes
	for {
		s := message.ReadString()
		message.mvd.demo.modellist = append(message.mvd.demo.modellist, s)
		if len(s) == 0 {
			break
		}
	}
	message.ReadByte() // some more indexes
}

func (message *Message) Svc_spawnbaseline(mvd *Mvd) {
	mvdPrint("entity: ", message.ReadShort())

	mvdPrint("modelindex: ", message.ReadByte())
	mvdPrint("frame: ", message.ReadByte())
	mvdPrint("colormap: ", message.ReadByte())
	mvdPrint("skinnume: ", message.ReadByte())

	for i := 0; i < 3; i++ {
		mvdPrint("coord: ", message.ReadCoord())
		mvdPrint("angle: ", message.ReadAngle())
	}
}

func (message *Message) Svc_updatefrags(mvd *Mvd) {
	player := message.ReadByte()
	frags := message.ReadShort()
	mvd.state.Players[int(player)].Frags = int(frags)
	mvd.EmitEventPlayer(&mvd.state.Players[int(player)], player, PE_STATS)
}

func (message *Message) Svc_playerinfo(mvd *Mvd) {
	var pe_type PE_TYPE
	pnum := message.ReadByte()
	p := &mvd.state.Players[pnum]
	flags := DF_TYPE(message.ReadShort())
	mvdPrint("frame: ", message.ReadByte())
	for i := 0; i < 3; i++ {
		t := DF_ORIGIN << i
		if flags&t == t {
			pe_type |= PE_MOVEMENT
			flags -= t
			switch i {
			case 0:
				{
					p.Origin.X = message.ReadCoord()
				}
			case 1:
				{
					p.Origin.Y = message.ReadCoord()
				}
			case 2:
				{
					p.Origin.Z = message.ReadCoord()
				}
			}
		}
	}
	for i := 0; i < 3; i++ {
		t := DF_ANGLES << i
		if flags&t == t {
			pe_type |= PE_MOVEMENT
			flags -= t
			switch i {
			case 0:
				{
					p.Angle.X = message.ReadAngle16()
				}
			case 1:
				{
					p.Angle.Y = message.ReadAngle16()
				}
			case 2:
				{
					p.Angle.Z = message.ReadAngle16()
				}
			}
		}
	}

	mvdPrint(flags)

	if flags&DF_MODEL == DF_MODEL {
		pe_type |= PE_ANIMATION
		p.ModelIndex = message.ReadByte() // modelindex
	}

	if flags&DF_SKINNUM == DF_SKINNUM {
		pe_type |= PE_ANIMATION
		p.SkinNum = message.ReadByte() // skinnum
	}

	if flags&DF_EFFECTS == DF_EFFECTS {
		pe_type |= PE_ANIMATION
		p.Effects = message.ReadByte() // effects
	}

	if flags&DF_WEAPONFRAME == DF_WEAPONFRAME {
		pe_type |= PE_ANIMATION
		p.WeaponFrame = message.ReadByte() // weaponframe
	}

	mvd.EmitEventPlayer(p, pnum, pe_type)
}

func (message *Message) Svc_updateping(mvd *Mvd) {
	pnum := message.ReadByte() // num
	p := &mvd.state.Players[pnum]
	p.Ping = message.ReadShort() // ping
	mvd.EmitEventPlayer(p, pnum, PE_NETWORKINFO)
}

func (message *Message) Svc_updatepl(mvd *Mvd) {
	pnum := message.ReadByte() // num
	p := &mvd.state.Players[pnum]
	p.Pl = message.ReadByte() // pl
	mvd.EmitEventPlayer(p, pnum, PE_NETWORKINFO)
}

func (message *Message) Svc_updateentertime(mvd *Mvd) {
	pnum := message.ReadByte() // num
	p := &mvd.state.Players[pnum]
	p.Entertime = message.ReadFloat() // entertime
	mvd.EmitEventPlayer(p, pnum, PE_NETWORKINFO)
}

func (message *Message) Svc_updateuserinfo(mvd *Mvd) {
	pnum := message.ReadByte() // num
	uid := message.ReadLong()  // userid
	p := &mvd.state.Players[pnum]
	p.Userid = uid
	ui := message.ReadString()
	if len(ui) < 2 {
		return
	}
	ui = ui[1:]
	splits := strings.Split(ui, "\\")
	for i := 0; i < len(splits); i += 2 {
		v := splits[i+1]
		switch splits[i] {
		case "name":
			p.Name = v
		case "team":
			p.Team = v
		case "*spectator":
			if v == "1" {
				p.Spectator = true
			}
		}
	}
	mvd.EmitEventPlayer(p, pnum, PE_USERINFO)
}

func (message *Message) Svc_sound(mvd *Mvd) {
	var s Sound
	channel := SND_TYPE(message.ReadShort()) // channel
	s.Channel = channel
	if channel&SND_VOLUME == SND_VOLUME {
		mvdPrint("has volume")
		s.Volume = message.ReadByte()
	}

	if channel&SND_ATTENUATION == SND_ATTENUATION {
		mvdPrint("has attenuation")
		s.Attenuation = message.ReadByte()
	}
	s.Index = message.ReadByte()
	s.Origin.Set(message.ReadCoord(), message.ReadCoord(), message.ReadCoord())
	mvd.state.SoundsActive = append(mvd.state.SoundsActive, s)
	mvd.EmitEventSound(&s)
}

func (message *Message) Svc_spawnstaticsound(mvd *Mvd) {
	var s Sound
	s.Origin.Set(message.ReadCoord(), message.ReadCoord(), message.ReadCoord())
	s.Index = message.ReadByte()       // sound_num
	s.Volume = message.ReadByte()      // sound volume
	s.Attenuation = message.ReadByte() // sound attenuation
	mvd.state.SoundsStatic = append(mvd.state.SoundsStatic, s)
}

func (message *Message) Svc_setangle(mvd *Mvd) {
	message.ReadByte()  // something weird?
	message.ReadAngle() // x
	message.ReadAngle()
	message.ReadAngle()
}

func (message *Message) Svc_lightstyle(mvd *Mvd) {
	b := message.ReadByte() // lightstyle num
	mvdPrint(b)
	mvdPrint(message.ReadString())
}

func (message *Message) Svc_updatestatlong(mvd *Mvd) {
	stat := STAT_TYPE(message.ReadByte())
	value := message.ReadLong()
	p := &mvd.state.Players[mvd.demo.last_to]
	s := fmt.Sprintf("%s", STAT_TYPE(stat))
	s = strings.TrimPrefix(s, "STAT_")
	s = strings.ToLower(s)
	s = strings.Title(s)
	ps := reflect.ValueOf(p)
	st := ps.Elem()
	f := st.FieldByName(s)
	if f.IsValid() {
		if f.CanSet() {
			if f.Kind() == reflect.Int {
				f.SetInt(int64(value))
			}
		}
	}
	mvd.EmitEventPlayer(p, byte(mvd.demo.last_to), PE_STATS)
}

func (message *Message) Svc_updatestat(mvd *Mvd) {
	stat := STAT_TYPE(message.ReadByte())
	value := message.ReadByte()
	p := &mvd.state.Players[mvd.demo.last_to]
	s := fmt.Sprintf("%s", STAT_TYPE(stat))
	s = strings.TrimPrefix(s, "STAT_")
	s = strings.ToLower(s)
	s = strings.Title(s)
	ps := reflect.ValueOf(p)
	st := ps.Elem()
	f := st.FieldByName(s)
	if f.IsValid() {
		if f.CanSet() {
			if f.Kind() == reflect.Int {
				f.SetInt(int64(value))
			}
		}
	}
	mvd.EmitEventPlayer(p, byte(mvd.demo.last_to), PE_STATS)
}

func (message *Message) Svc_deltapacketentities(mvd *Mvd) {
	from := message.ReadByte()
	mvdPrint(from)
	for {
		w := message.ReadShort()
		if w == 0 {
			break
		}

		w &= ^511
		bits := w

		if bits&U_MOREBITS == U_MOREBITS {
			i := message.ReadByte()
			bits |= int(i)
		}

		if bits&U_MODEL == U_MODEL {
			message.ReadByte()
		}
		if bits&U_FRAME == U_FRAME {
			message.ReadByte()
		}
		if bits&U_COLORMAP == U_COLORMAP {
			message.ReadByte()
		}
		if bits&U_SKIN == U_SKIN {
			message.ReadByte()
		}
		if bits&U_EFFECTS == U_EFFECTS {
			message.ReadByte()
		}
		if bits&U_ORIGIN1 == U_ORIGIN1 {
			message.ReadCoord()
		}
		if bits&U_ANGLE1 == U_ANGLE1 {
			message.ReadAngle()
		}
		if bits&U_ORIGIN2 == U_ORIGIN2 {
			message.ReadCoord()
		}
		if bits&U_ANGLE2 == U_ANGLE2 {
			message.ReadAngle()
		}
		if bits&U_ORIGIN3 == U_ORIGIN3 {
			message.ReadCoord()
		}
		if bits&U_ANGLE3 == U_ANGLE3 {
			message.ReadAngle()
		}
	}
}

func (message *Message) Svc_packetentities(mvd *Mvd) {
	for {
		w := message.ReadShort()
		if w == 0 {
			break
		}

		w &= ^511
		bits := w

		if bits&U_MOREBITS == U_MOREBITS {
			i := message.ReadByte()
			bits |= int(i)
		}

		if bits&U_MODEL == U_MODEL {
			message.ReadByte()
		}
		if bits&U_FRAME == U_FRAME {
			message.ReadByte()
		}
		if bits&U_COLORMAP == U_COLORMAP {
			message.ReadByte()
		}
		if bits&U_SKIN == U_SKIN {
			message.ReadByte()
		}
		if bits&U_EFFECTS == U_EFFECTS {
			message.ReadByte()
		}
		if bits&U_ORIGIN1 == U_ORIGIN1 {
			message.ReadCoord()
		}
		if bits&U_ANGLE1 == U_ANGLE1 {
			message.ReadAngle()
		}
		if bits&U_ORIGIN2 == U_ORIGIN2 {
			message.ReadCoord()
		}
		if bits&U_ANGLE2 == U_ANGLE2 {
			message.ReadAngle()
		}
		if bits&U_ORIGIN3 == U_ORIGIN3 {
			message.ReadCoord()
		}
		if bits&U_ANGLE3 == U_ANGLE3 {
			message.ReadAngle()
		}
	}
}

func (message *Message) Svc_temp_entity(mvd *Mvd) {

	t := message.ReadByte()

	if t == TE_GUNSHOT || t == TE_BLOOD {
		message.ReadByte()
	}

	if t == TE_LIGHTNING1 || t == TE_LIGHTNING2 || t == TE_LIGHTNING3 {
		message.ReadShort()
		message.ReadCoord()
		message.ReadCoord()
		message.ReadCoord()
	}

	message.ReadCoord()
	message.ReadCoord()
	message.ReadCoord()
	return
}

func (message *Message) Svc_print(mvd *Mvd) {
	from := message.ReadByte()
	s := message.ReadString()
	mvdPrint(from, s)
}

func (message *Message) Svc_serverinfo(mvd *Mvd) {
	key := message.ReadString()
	value := message.ReadString()
	if key == "hostname" {
		mvd.state.Hostname = value
	}
	mvdPrint(key, value)
}

func (message *Message) Svc_centerprint(mvd *Mvd) {
	s := message.ReadString()
	mvdPrint(s)
}

func (message *Message) Svc_setinfo(mvd *Mvd) {
	message.ReadByte()   // num
	message.ReadString() // key
	message.ReadString() // value
}

func (message *Message) Svc_damage(mvd *Mvd) {
	message.ReadByte() // armor
	message.ReadByte() // blood
	message.ReadCoord()
	message.ReadCoord()
	message.ReadCoord()
}

func (message *Message) Svc_chokecount(mvd *Mvd) {
	message.ReadByte()
}

func (message *Message) Svc_spawnstatic(mvd *Mvd) {
	message.ReadByte()
	message.ReadByte()
	message.ReadByte()
	message.ReadByte()
	message.ReadCoord()
	message.ReadAngle()
	message.ReadCoord()
	message.ReadAngle()
	message.ReadCoord()
	message.ReadAngle()
}

func (message *Message) Nq_svc_cutscene(mvd *Mvd) {
	message.Svc_smallkick(mvd)
}

func (message *Message) Svc_smallkick(mvd *Mvd) {
}

func (message *Message) Svc_bigkick(mvd *Mvd) {
}

func (message *Message) Svc_muzzleflash(mvd *Mvd) {
	message.ReadShort()
}

func (message *Message) Svc_intermission(mvd *Mvd) {
	message.ReadCoord()
	message.ReadCoord()
	message.ReadCoord()
	message.ReadAngle()
	message.ReadAngle()
	message.ReadAngle()
}

func (message *Message) Svc_disconnect(mvd *Mvd) {
	mvd.done = true
}

func (message *Message) ReadBytes(count uint) *bytes.Buffer {
	b := bytes.NewBuffer(message.data[message.offset : message.offset+count])
	//mvdPrint("offset: ", message.offset, " - len: ", b.Len())
	message.offset += count
	return b
}

func (message *Message) ReadByte() byte {
	var b byte
	err := binary.Read(message.ReadBytes(1), binary.BigEndian, &b)
	if err != nil {
		message.mvd.Error.Fatal(err)
	}
	return b
}

func (message *Message) ReadLong() int {
	var i int32
	err := binary.Read(message.ReadBytes(4), binary.LittleEndian, &i)
	if err != nil {
		message.mvd.Error.Fatal(err)
	}
	return int(i)
}

func (message *Message) ReadFloat() float32 {
	var i float32
	err := binary.Read(message.ReadBytes(4), binary.LittleEndian, &i)
	if err != nil {
		message.mvd.Error.Fatal(err)
	}
	return float32(i)
}

func (message *Message) ReadString() string {
	b := make([]byte, 0)
	for {
		c := message.ReadByte()
		if c == 255 {
			continue
		}
		if c == 0 {
			break
		}
		b = append(b, c)
	}
	return string(b)
}

func (message *Message) ReadCoord() float32 {
	if message.mvd.demo.fte_pext&FTE_PEXT_FLOATCOORDS == FTE_PEXT_FLOATCOORDS {
		return message.ReadFloat()
	}
	b := message.ReadShort()
	return float32(b) * (1.0 / 8)
}

func (message *Message) ReadAngle() float32 {
	if message.mvd.demo.fte_pext&FTE_PEXT_FLOATCOORDS == FTE_PEXT_FLOATCOORDS {
		return message.ReadAngle16()
	}
	b := message.ReadByte()
	return float32(b) * (360.0 / 256.0)
}

func (message *Message) ReadAngle16() float32 {
	b := message.ReadShort()
	return float32(b) * (360.0 / 65536)
}

func (message *Message) ReadShort() int {
	var i int16
	err := binary.Read(message.ReadBytes(2), binary.LittleEndian, &i)
	if err != nil {
		message.mvd.Error.Fatal(err)
	}
	return int(i)
}
