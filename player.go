package main

import (
    fragfile "github.com/jogi1/golang-fragfile"
	"github.com/jogi1/mvdreader"
)

type ParserPlayer struct {
	pnum              int
	Name              mvdreader.ReaderString `json:"name"`
	Team              mvdreader.ReaderString `json:"team"`
	ParserStats       *Stats `json:"parser_stats,omitempty"`
	FragmessagesStats *fragfile.FragMessage `json:"frag_messages_stats,omitempty"`
	ModStats          interface{} `json:"mod_stats,omitempty"`
	WpsStats          map[string]WpsStats `json:"wps_stats,omitempty"`
}

func (p *Parser) FindPlayerPnum(pnum int) *ParserPlayer {
	for _, player := range p.PlayersFrameCurrent {
        if player.pnum == pnum {
            return player
        }
    }
    return nil
}

func (p *Parser) FindPlayer(name, team interface{}) *ParserPlayer {
	if name == nil {
		return nil
	}
	for _, player := range p.PlayersFrameCurrent {
		if player.Name.Equal(name) {
			if team != nil {
				if player.Team.Equal(team) {
					return player
				}
			} else {
				return player
			}
		}
	}
	return nil
}

func (p *Parser) PlayersNewFrame() error {
    if !p.Flags.AggregatePlayerInfo {
        return nil
    }
	p.PlayersFrameLast = p.PlayersFrameCurrent
	p.PlayersFrameCurrent = nil
	for i, player := range p.mvd.State.Players {
		if player.Spectator {
			continue
		}
		if len(player.Name.String) == 0 {
			continue
		}
		parserPlayer := new(ParserPlayer)
		parserPlayer.Name = player.Name
		parserPlayer.Team = player.Team
		parserPlayer.pnum = i
        if p.Flags.WpsParserEnabled {
            player := p.FindPlayerPnum(i)
            if player != nil {
                parserPlayer.WpsStats = player.WpsStats
            } else {
                parserPlayer.WpsStats = make(map[string]WpsStats)
            }
        }
		p.PlayersFrameCurrent = append(p.PlayersFrameCurrent, parserPlayer)
	}
	return nil
}
