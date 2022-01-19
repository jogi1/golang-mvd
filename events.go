package main

import (
	"github.com/jogi1/mvdreader"
)

func (parser *Parser) clearPlayerEvents() {
	parser.events = nil
}

func (parser *Parser) handlePlayerEvents() {
	for _, __player := range parser.mvd.State.Players {
		player := &__player
		player_num := int(player.EventInfo.Pnum)
		if player.EventInfo.Events == 0 {
			continue
		}
		p := &parser.mvd.State_last_frame.Players[player.EventInfo.Pnum]
		if p.Userid != player.Userid {
			// @TODO: Handle this better
			return
		}

		stat, ok := parser.stats[player.Userid]
		if !ok {
			stat := new(Stats)
			parser.stats[player.Userid] = stat
		}

		if player.EventInfo.Events&mvdreader.PE_STATS == mvdreader.PE_STATS {
			if player.Health > 0 && p.Health <= 0 {
				e := Event_Player{EPT_Spawn, player_num}
				parser.events = append(parser.events, e)
			}

			if player.Frags > p.Frags {
				stat.Kills += player.Frags - p.Frags
				e := Event_Player{EPT_Kill, player_num}
				parser.events = append(parser.events, e)
			}
			if player.Health <= 0 && p.Health > 0 {
				stat.Deaths += 1
				e := Event_Player{EPT_Death, player_num}
				parser.events = append(parser.events, e)
			}

			if player.Frags < p.Frags {
				if player.Health <= 0 {
					stat.Suicides += 1
					e := Event_Player{EPT_Suicide, player_num}
					parser.events = append(parser.events, e)
				} else {
					stat.Teamkills += 1
					e := Event_Player{EPT_Teamkill, player_num}
					parser.events = append(parser.events, e)
				}
			}

			if player.Items != p.Items || true {
				itemstat := parser.stats[int(player.Userid)]
				itemstat.SuperShotgun.CheckItem(parser, mvdreader.IT_SUPER_SHOTGUN, player, p)
				itemstat.NailGun.CheckItem(parser, mvdreader.IT_NAILGUN, player, p)
				itemstat.SuperNailGun.CheckItem(parser, mvdreader.IT_SUPER_NAILGUN, player, p)
				itemstat.GrenadeLauncher.CheckItem(parser, mvdreader.IT_GRENADE_LAUNCHER, player, p)
				itemstat.RocketLauncher.CheckItem(parser, mvdreader.IT_ROCKET_LAUNCHER, player, p)
				itemstat.LightningGun.CheckItem(parser, mvdreader.IT_LIGHTNING, player, p)

				itemstat.GreenArmor.CheckItem(parser, mvdreader.IT_ARMOR1, player, p)
				itemstat.YellowArmor.CheckItem(parser, mvdreader.IT_ARMOR2, player, p)
				itemstat.RedArmor.CheckItem(parser, mvdreader.IT_ARMOR3, player, p)

				itemstat.MegaHealth.CheckItem(parser, mvdreader.IT_SUPERHEALTH, player, p)
				itemstat.Quad.CheckItem(parser, mvdreader.IT_QUAD, player, p)
				itemstat.Pentagram.CheckItem(parser, mvdreader.IT_INVULNERABILITY, player, p)
				itemstat.Ring.CheckItem(parser, mvdreader.IT_INVISIBILITY, player, p)
			}
		}
	}
}

func (s *Weapon_Stat) CheckItem(parser *Parser, iitem mvdreader.IT_TYPE, cf, lf *mvdreader.Player) int {
	item := int(iitem)
	if cf.Items&item == item && lf.Items&item == 0 {
		s.Pickup += 1
		e := Event_Player_Item{EPT_Pickup, int(cf.EventInfo.Pnum), iitem}
		parser.events = append(parser.events, e)
		return 1
	}
	if cf.Items&item == 0 && lf.Items&item == item {
		s.Drop += 1
		e := Event_Player_Item{EPT_Drop, int(cf.EventInfo.Pnum), iitem}
		parser.events = append(parser.events, e)
		return -1
	}
	return 0
}

func (s *Armor_Stat) CheckItem(parser *Parser, iitem mvdreader.IT_TYPE, cf, lf *mvdreader.Player) {
	item := int(iitem)
	if cf.Items&item == item && lf.Items&item == 0 {
		s.Pickup++
		e := Event_Player_Item{EPT_Drop, int(cf.EventInfo.Pnum), iitem}
		parser.events = append(parser.events, e)
	}
	if cf.Items&item == item && lf.Items&item == item {
		if cf.Armor > lf.Armor {
			s.Pickup++
			e := Event_Player_Item{EPT_Drop, int(cf.EventInfo.Pnum), iitem}
			parser.events = append(parser.events, e)
		}
		if cf.Armor < lf.Armor {
			s.Damage_Absorbed += lf.Armor - cf.Armor
		}
	}
}

func (s *Item_Stat) CheckItem(parser *Parser, iitem mvdreader.IT_TYPE, cf, lf *mvdreader.Player) {
	item := int(iitem)
	if cf.Items&item == item && lf.Items&item == 0 {
		s.Pickup++
	}

	if iitem == mvdreader.IT_SUPERHEALTH {
		if cf.Items&item == item && lf.Health > 100 && cf.Health > lf.Health {
			s.Pickup++
		}
	}
}
