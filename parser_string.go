package main

import "bytes"

type ParserString struct {
	p      *Parser
	String string
	Byte   []byte
}

func (ps *ParserString) Equal(c *ParserString) bool {
	return bytes.Equal(ps.Byte, c.Byte)
}

func (p *Parser) ParserStringNew(b []byte) *ParserString {
	s := new(ParserString)
	s.p = p
	s.Set(b)
	return s
}

func (s *ParserString) Set(b []byte) {
	s.Byte = b
	s.String = s.p.SanatizeName(string(b))
}
