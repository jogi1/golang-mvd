package main

import (
	"io/ioutil"
	"strings"
)

var qw_ascii_table []rune

func (mvd *Mvd) Ascii_Init() {
	ascii_table, err := ioutil.ReadFile("ascii.table")
	if err != nil {
		err = nil
		s, err := Asset("data/ascii.table")
		if err != nil {
			mvd.Error.Fatal(err)
		}
		qw_ascii_table = []rune(string(s))
		return
	}
	s := string(ascii_table)
	s = strings.TrimRight(s, "\r\n")
	qw_ascii_table = []rune(string(ascii_table))
}
