package main

import (
	"fmt"
	"strconv"
	"strings"
)

type WpsStats struct {
	Attacks int
	Hits    int
}

func (p *Parser) WpsParserParse(message string) error {
	if !strings.HasPrefix(message, "//wps") {
		return nil
	}
	s := strings.TrimPrefix(message, "//wps")
	s = strings.TrimRight(s, "\n")
	s = strings.TrimLeft(s, " ")
	info := strings.Split(s, " ")
	if len(info) != 4 {
		return fmt.Errorf("wrong amount of wps fields")
	}
	i, err := strconv.Atoi(info[0])
	if err != nil {
		return err
	}
	player := p.FindPlayerPnum(i)
	if player == nil {
		return fmt.Errorf("player not found")
	}

	attacks, err := strconv.Atoi(info[2])
	if err != nil {
		return err
	}
	hits, err := strconv.Atoi(info[3])
	if err != nil {
		return err
	}
	player.WpsStats[info[1]] = WpsStats{attacks, hits}
	return nil
}
