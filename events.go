package main

import (
	"github.com/jogi1/mvdreader"
)

//go:generate stringer -type=Event_Type
type Event_Type uint

const (
	EPT_Spawn Event_Type = iota
	EPT_Death
	EPT_Suicide
	EPT_Kill
	EPT_Teamkill
	EPT_Pickup
	EPT_Drop
)

// EPT_Spawn, EPT_Death, EPT_Suicide
type Event_Player struct {
	Type          Event_Type
	Player_Number int
}

// EPT_Kill, EPT_Teamkill
type Event_Player_Kill struct {
	Type          Event_Type
	Player_Number int
}

type Event_Player_Item struct {
	Type          Event_Type
	Player_Number int
	Item_Type     mvdreader.IT_TYPE
}

type Event_Player_Stat struct {
	Type          Event_Type
	Player_Number int
	Stat          mvdreader.STAT_TYPE
	Amount        int
}


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


        var stat *Stats;
        if parser.Flags.StatsEnabled {
            stat, ok := parser.stats[player.Userid]
            if !ok || stat == nil {
                stat = new(Stats)
                parser.stats[player.Userid] = stat
            }

            if parser.Flags.StatsStrengthEnabled {
                if stat != nil {
                    stat.StrengthCalculate(parser, player)
                }
            }
        }

		if player.EventInfo.Events&mvdreader.PE_STATS == mvdreader.PE_STATS {
			if player.Health > 0 && p.Health <= 0 {
				e := Event_Player{EPT_Spawn, player_num}
				parser.events = append(parser.events, e)
                if stat != nil {
                    stat.Spawn(parser, player, p)
                }
			}

			if player.Frags > p.Frags {
				e := Event_Player{EPT_Kill, player_num}
				parser.events = append(parser.events, e)
                if stat != nil {
                    stat.Kills += player.Frags - p.Frags
                    stat.Kill(parser, player, p)
                }
			}
			if player.Health <= 0 && p.Health > 0 {
				e := Event_Player{EPT_Death, player_num}
				parser.events = append(parser.events, e)
                if stat != nil {
                    stat.Deaths += 1
                    stat.Death(parser, player, p)
                }
			}

			if player.Frags < p.Frags {
				if player.Health <= 0 {
					e := Event_Player{EPT_Suicide, player_num}
					parser.events = append(parser.events, e)
                    if stat != nil {
                        stat.Suicide(parser, player, p)
                        stat.Suicides += 1
                    }
				} else {
					e := Event_Player{EPT_Teamkill, player_num}
					parser.events = append(parser.events, e)
                    if stat != nil {
                        stat.Teamkill(parser, player, p)
                        stat.Teamkills += 1
                    }
				}
			}

			if player.Items != p.Items || true {
				itemstat := parser.stats[int(player.Userid)]
                if itemstat != nil {
                    itemstat.SuperShotgun.CheckItem(parser, mvdreader.IT_SUPER_SHOTGUN, player, p, stat)
                    itemstat.NailGun.CheckItem(parser, mvdreader.IT_NAILGUN, player, p, stat)
                    itemstat.SuperNailGun.CheckItem(parser, mvdreader.IT_SUPER_NAILGUN, player, p, stat)
                    itemstat.GrenadeLauncher.CheckItem(parser, mvdreader.IT_GRENADE_LAUNCHER, player, p, stat)
                    itemstat.RocketLauncher.CheckItem(parser, mvdreader.IT_ROCKET_LAUNCHER, player, p, stat)
                    itemstat.LightningGun.CheckItem(parser, mvdreader.IT_LIGHTNING, player, p, stat)

                    itemstat.GreenArmor.CheckItem(parser, mvdreader.IT_ARMOR1, player, p, stat)
                    itemstat.YellowArmor.CheckItem(parser, mvdreader.IT_ARMOR2, player, p, stat)
                    itemstat.RedArmor.CheckItem(parser, mvdreader.IT_ARMOR3, player, p, stat)

                    itemstat.MegaHealth.CheckItem(parser, mvdreader.IT_SUPERHEALTH, player, p, stat)
                    itemstat.Quad.CheckItem(parser, mvdreader.IT_QUAD, player, p, stat)
                    itemstat.Pentagram.CheckItem(parser, mvdreader.IT_INVULNERABILITY, player, p, stat)
                    itemstat.Ring.CheckItem(parser, mvdreader.IT_INVISIBILITY, player, p, stat)
                }
			}

            if parser.Flags.AggregatePlayerInfo {
                ps := parser.FindPlayerPnum(player_num)
                if ps != nil && stat != nil {
                    ps.ParserStats = stat
                }
            }
		}
	}
}

func (s *Weapon_Stat) CheckItem(parser *Parser, iitem mvdreader.IT_TYPE, cf, lf *mvdreader.Player, stat *Stats) int {
	item := int(iitem)
	if cf.Items&item == item && lf.Items&item == 0 {
		s.Pickup += 1
        if stat != nil {
            stat.Pickup(parser, lf, cf, iitem)
        }
		e := Event_Player_Item{EPT_Pickup, int(cf.EventInfo.Pnum), iitem}
		parser.events = append(parser.events, e)
		return 1
	}
	if cf.Items&item == 0 && lf.Items&item == item {
		s.Drop += 1
        if stat != nil {
            stat.Drop(parser, lf, cf, iitem)
        }
		e := Event_Player_Item{EPT_Drop, int(cf.EventInfo.Pnum), iitem}
		parser.events = append(parser.events, e)
		return -1
	}
	return 0
}

func (s *Armor_Stat) CheckItem(parser *Parser, iitem mvdreader.IT_TYPE, cf, lf *mvdreader.Player, stat *Stats) {
	item := int(iitem)
	if cf.Items&item == item && lf.Items&item == 0 {
		s.Pickup++
        if stat != nil {
            stat.Pickup(parser, lf, cf, iitem)
        }
		e := Event_Player_Item{EPT_Drop, int(cf.EventInfo.Pnum), iitem}
		parser.events = append(parser.events, e)
	}
	if cf.Items&item == item && lf.Items&item == item {
		if cf.Armor > lf.Armor {
			s.Pickup++
            if stat != nil {
                stat.Pickup(parser, lf, cf, iitem)
            }
			e := Event_Player_Item{EPT_Drop, int(cf.EventInfo.Pnum), iitem}
			parser.events = append(parser.events, e)
		}
		if cf.Armor < lf.Armor {
			s.Damage_Absorbed += lf.Armor - cf.Armor
		}
	}
}

func (s *Item_Stat) CheckItem(parser *Parser, iitem mvdreader.IT_TYPE, cf, lf *mvdreader.Player, stat *Stats) {
	item := int(iitem)
	if cf.Items&item == item && lf.Items&item == 0 {
        if stat != nil {
            stat.Pickup(parser, lf, cf, iitem)
        }
		s.Pickup++
	}

	if iitem == mvdreader.IT_SUPERHEALTH {
		if cf.Items&item == item && lf.Health > 100 && cf.Health > lf.Health {
            if stat != nil {
                stat.Pickup(parser, lf, cf, iitem)
            }
			s.Pickup++
		}
	}
}
