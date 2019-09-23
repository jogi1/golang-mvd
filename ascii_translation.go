package main

import (
	"fmt"
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

func sanatize_name(name string) string {
	r := []byte(name)
	var b strings.Builder
	for _, ri := range r {
		fmt.Fprintf(&b, "%c", qw_ascii_table[uint(ri)])
	}
	return b.String()
}

/*
func unicode_string(str string) string {
	r := []byte(str)
	var b strings.Builder
	for _, ri := range r {
		var rt rune
		if ri >= 32 && ri <= 127 {
			rt = rune(ri) | 0xe080
		} else {
			rt = rune(ri)
		}
		fmt.Fprintf(&b, "%c", rt)
	}
	return b.String()
}
*/

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
