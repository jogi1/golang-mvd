package main

func (parser *Parser) handlePlayerDisconnects() {
	for _, p := range parser.mvd.State.Players {
		if p.Spectator == true {
			continue
		}
		if len(p.Name) == 0 {
			continue
		}
		// TODO: not sure about this
		if p.Entertime == 0 {
			continue
		}
		parser.players[p.Userid] = p
	}
}
