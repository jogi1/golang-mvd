package main

import fragfile "github.com/jogi1/golang-fragfile"

type ParserPlayer struct {
	pnum              int
	Name              *ParserString
	Team              *ParserString
	ParserStats       *Stats                // `json:",omitempty"`
	FragmessagesStats *fragfile.FragMessage // `json:",omitempty"`
	ModStats          interface{}           // `json:",omitempty"`
	WpsStats          map[string]WpsStats   // `json:",omitempty"`
}

func (p *Parser) FindPlayerPnum(pnum int) *ParserPlayer {
	for _, player := range p.PlayersFrameCurrent {
        if player.pnum == pnum {
            return player
        }
    }
    return nil
}

func (p *Parser) FindPlayer(name, team *ParserString) *ParserPlayer {
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
	p.PlayersFrameLast = p.PlayersFrameCurrent
	p.PlayersFrameCurrent = nil
	for i, player := range p.mvd.State.Players {
		if player.Spectator {
			continue
		}
		if len(player.Name) == 0 {
			continue
		}
		parserPlayer := new(ParserPlayer)
		parserPlayer.Name = p.ParserStringNew([]byte(player.Name))
		parserPlayer.Team = p.ParserStringNew([]byte(player.Team))
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
