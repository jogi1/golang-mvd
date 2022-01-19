package main

type (
	ModCheckServerinfoFunction func(map[string]string) bool
	ModFrameFunction           func(*Parser) error
	ModEndFunction             func(*Parser) error
)

type Mod struct {
	Name    string
	Version string
	Check   ModCheckServerinfoFunction
	Frame   ModFrameFunction
	End     ModEndFunction
	State   interface{}
}

type ModParserString struct {
	p      *Parser
	String string
	Byte   []byte
}

func (p *Parser) ModParserStringNew(b []byte) *ModParserString {
	s := new(ModParserString)
	s.p = p
	s.Set(b)
	return s
}

func (s *ModParserString) Set(b []byte) {
	s.Byte = b
	s.String = s.p.sanatize_name(string(b))
}

func (p *Parser) ModParserPlayerNew(b []byte, stat interface{}) *ModParserPlayer {
	pl := new(ModParserPlayer)
	pl.Name = p.ModParserStringNew(b)
	pl.Stat = stat
	return pl
}

type ModParserPlayer struct {
	Name *ModParserString
	Stat interface{}
}

func (p *Parser) ModParserTeamNew(b []byte, stat interface{}) *ModParserTeam {
	t := new(ModParserTeam)
	t.Name = p.ModParserStringNew(b)
	t.Stat = stat
	t.Players = make([]*ModParserPlayer, 0)
	return t
}

func (t *ModParserTeam) PlayerAdd(p *ModParserPlayer) {
	t.Players = append(t.Players, p)
}

type ModParserTeam struct {
	Name    *ModParserString
	Stat    interface{}
	Players []*ModParserPlayer
}

type ModState struct {
	parsing       bool
	parsingState  int
	currentPlayer *ModParserPlayer
	currentTeam   *ModParserTeam
}

var AvailableModParsers []Mod

func (p *Parser) ModInfoParserInit() {
	AvailableModParsers = append(
		AvailableModParsers,
		Mod{"ktx", "1.40-beta-quakecon-release3", KTX_1_4b_Check, KTX_1_4b_Frame, KTX_1_4b_End, nil},
	)
}

func (p *Parser) ModInfoParserFind() {
	if p.mvd.Frame > 100 {
		return
	}
	for _, m := range AvailableModParsers {
		if m.Check(p.mvd.Server.Serverinfo) {
			p.mod_parser = &m
			return
		}
	}
}
