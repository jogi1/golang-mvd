package main

func (parser *Parser) handlePlayerDisconnects() {
	for _, p := range parser.mvd.State.Players {

		if p.Spectator == true {
			continue
		}
		if len(p.Name.String) == 0 {
			continue
		}

		// TODO: not so sure
		if p.Userid == 0 {
			continue
		}

		parser.players[p.Userid] = p
	}
}
