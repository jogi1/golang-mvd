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
	for i, _ := range mvd.state.Players {
		player := &mvd.state.Players[i]
		if player.event_info.events == 0 {
			continue
		}
		p := &mvd.state_last_frame.Players[player.event_info.pnum]
		if p.Userid != player.Userid {
			// @TODO: Handle this better
			return
		}
		if player.event_info.events&PE_STATS == PE_STATS {
			if player.Health <= 0 && p.Health > 0 {
				player.Deaths += 1
			}

			if player.Items != p.Items {
				player.Itemstats.SuperShotgun.CheckItem(IT_SUPER_SHOTGUN, player, p)
				player.Itemstats.NailGun.CheckItem(IT_NAILGUN, player, p)
				player.Itemstats.SuperNailGun.CheckItem(IT_SUPER_NAILGUN, player, p)
				player.Itemstats.GrenadeLauncher.CheckItem(IT_GRENADE_LAUNCHER, player, p)
				player.Itemstats.RocketLauncher.CheckItem(IT_ROCKET_LAUNCHER, player, p)
				player.Itemstats.LightningGun.CheckItem(IT_LIGHTNING, player, p)

				player.Itemstats.GreenArmor.CheckItem(IT_ARMOR1, player, p)
				player.Itemstats.YellowArmor.CheckItem(IT_ARMOR2, player, p)
				player.Itemstats.RedArmor.CheckItem(IT_ARMOR3, player, p)

				player.Itemstats.MegaHealth.CheckItem(IT_SUPERHEALTH, player, p)
			}
		}
		player.event_info.events = 0
		player.event_info.pnum = 0
	}
}

func (mvd *Mvd) EmitEventSound(sound *Sound) {
}

func (s *Weapon_Stat) CheckItem(iitem IT_TYPE, cf, lf *Player) {
	item := int(iitem)
	if cf.Items&item == item && lf.Items&item == 0 {
		s.Pickup += 1
	}
	if cf.Items&item == 0 && lf.Items&item == item {
		s.Drop += 1
	}
}

func (s *Armor_Stat) CheckItem(iitem IT_TYPE, cf, lf *Player) {
	item := int(iitem)
	if cf.Items&item == item && lf.Items&item == 0 {
		s.Pickup += 1
	}
}

func (s *Item_Stat) CheckItem(iitem IT_TYPE, cf, lf *Player) {
	item := int(iitem)
	if cf.Items&item == item && lf.Items&item == 0 {
		s.Pickup += 1
	}
}
