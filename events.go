package main

import (
	"fmt"
)

//go:generate
//stringer -type=PE_TYPE
type PE_TYPE int

const (
	PE_MOVEMENT    PE_TYPE = 1 << 1
	PE_STATS       PE_TYPE = 1 << 2
	PE_ANIMATION   PE_TYPE = 1 << 3
	PE_NETWORKINFO PE_TYPE = 1 << 4
	PE_USERINFO    PE_TYPE = 1 << 5
)

type event struct {
	player *Player
}

func (mvd *Mvd) EmitEvent(event interface{}) {
	fmt.Println("fuck you parser!")
}

func (mvd *Mvd) EmitEventPlayer(player *Player, pnum byte, pe_type PE_TYPE) {
	//fmt.Printf("player (%s) changed in frame(%d)\n", player.Name, mvd.frame)
	player.event_info.pnum = pnum
	player.event_info.events |= pe_type
}

func (mvd *Mvd) HandlePlayerEvents() {
	for _, __player := range mvd.state.Players {
		player := &__player
		player_num := int(player.event_info.pnum)
		if player.event_info.events == 0 {
			continue
		}
		p := &mvd.state_last_frame.Players[player.event_info.pnum]
		if p.Userid != player.Userid {
			// @TODO: Handle this better
			return
		}
		if player.event_info.events&PE_STATS == PE_STATS {
			if player.Health > 0 && p.Health <= 0 {
				e := Event_Player{EPT_Spawn, player_num}
				mvd.state.Events = append(mvd.state.Events, e)
			}
			if player.Health <= 0 && p.Health > 0 {
				player.Deaths += 1
				e := Event_Player{EPT_Death, player_num}
				mvd.state.Events = append(mvd.state.Events, e)
			}

			if player.Frags < p.Frags {
				if player.Health <= 0 {
					player.Suicides += 1
					e := Event_Player{EPT_Suicide, player_num}
					mvd.state.Events = append(mvd.state.Events, e)
				} else {
					player.Teamkills += 1
					e := Event_Player{EPT_Teamkill, player_num}
					mvd.state.Events = append(mvd.state.Events, e)
				}
			}

			if player.Items != p.Items {
				player.Itemstats.SuperShotgun.CheckItem(mvd, IT_SUPER_SHOTGUN, player, p)
				player.Itemstats.NailGun.CheckItem(mvd, IT_NAILGUN, player, p)
				player.Itemstats.SuperNailGun.CheckItem(mvd, IT_SUPER_NAILGUN, player, p)
				player.Itemstats.GrenadeLauncher.CheckItem(mvd, IT_GRENADE_LAUNCHER, player, p)
				player.Itemstats.RocketLauncher.CheckItem(mvd, IT_ROCKET_LAUNCHER, player, p)
				player.Itemstats.LightningGun.CheckItem(mvd, IT_LIGHTNING, player, p)

				player.Itemstats.GreenArmor.CheckItem(mvd, IT_ARMOR1, player, p)
				player.Itemstats.YellowArmor.CheckItem(mvd, IT_ARMOR2, player, p)
				player.Itemstats.RedArmor.CheckItem(mvd, IT_ARMOR3, player, p)

				player.Itemstats.MegaHealth.CheckItem(mvd, IT_SUPERHEALTH, player, p)
				player.Itemstats.Quad.CheckItem(mvd, IT_QUAD, player, p)
				player.Itemstats.Pentagram.CheckItem(mvd, IT_INVULNERABILITY, player, p)
				player.Itemstats.Ring.CheckItem(mvd, IT_INVISIBILITY, player, p)
			}
		}
		player.event_info.events = 0
		player.event_info.pnum = 0
	}
}

func (s *Weapon_Stat) CheckItem(mvd *Mvd, iitem IT_TYPE, cf, lf *Player) int {
	item := int(iitem)
	if cf.Items&item == item && lf.Items&item == 0 {
		s.Pickup += 1
		e := Event_Player_Item{EPT_Pickup, int(cf.event_info.pnum), iitem}
		mvd.state.Events = append(mvd.state.Events, e)
		return 1
	}
	if cf.Items&item == 0 && lf.Items&item == item {
		s.Drop += 1
		e := Event_Player_Item{EPT_Drop, int(cf.event_info.pnum), iitem}
		mvd.state.Events = append(mvd.state.Events, e)
		return -1
	}
	return 0
}

func (s *Armor_Stat) CheckItem(mvd *Mvd, iitem IT_TYPE, cf, lf *Player) {
	item := int(iitem)
	if cf.Items&item == item && lf.Items&item == 0 {
		s.Pickup += 1
		e := Event_Player_Item{EPT_Drop, int(cf.event_info.pnum), iitem}
		mvd.state.Events = append(mvd.state.Events, e)
	}
	if cf.Items&item == item && lf.Items&item == item {
		if cf.Armor > lf.Armor {
			s.Pickup += 1
			e := Event_Player_Item{EPT_Drop, int(cf.event_info.pnum), iitem}
			mvd.state.Events = append(mvd.state.Events, e)
		}
		if cf.Armor < lf.Armor {
			s.Damage_Absorbed += lf.Armor - cf.Armor
		}
	}
}

func (s *Item_Stat) CheckItem(mvd *Mvd, iitem IT_TYPE, cf, lf *Player) {
	item := int(iitem)
	if cf.Items&item == item && lf.Items&item == 0 {
		s.Pickup += 1
	}

	if iitem == IT_SUPERHEALTH {
		if cf.Items&item == item && lf.Health > 100 && cf.Health > lf.Health {
			s.Pickup += 1
		}
	}
}

func (mvd *Mvd) EmitEventSound(sound *Sound) {
	sound.Frame = mvd.frame
	mvd.state.SoundsActive = append(mvd.state.SoundsActive, *sound)
}
