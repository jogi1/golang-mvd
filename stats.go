package main

import (
	"github.com/jogi1/mvdreader"
)

type Weapon_Stat struct {
	Pickup, Drop, Damage int
}

type Armor_Stat struct {
	Pickup, Damage_Absorbed int
}

type Item_Stat struct {
	Pickup, Drop int
}

type RunKill struct {
	Time float64
}

type RunPickup struct {
	Time float64
	Item mvdreader.IT_TYPE
}

type Run struct {
	Start, Stop float64
	Kills       []RunKill
	Pickups     []RunPickup
}

type Stats struct {
	Axe, Shotgun, SuperShotgun, NailGun, SuperNailGun, GrenadeLauncher, RocketLauncher, LightningGun Weapon_Stat
	GreenArmor, YellowArmor, RedArmor                                                                Armor_Stat
	MegaHealth, Quad, Pentagram, Ring                                                                Item_Stat
	Kills, Deaths, Suicides, Teamkills                                                               int
	Runs                                                                                             []*Run
	currentRun                                                                                       *Run

	Strength []float64
}

func (s *Stats) StrengthCalculate(parser *Parser, player *mvdreader.Player) {
	armorModifier := 0.0
	if player.HasItem(mvdreader.IT_ARMOR1) {
		armorModifier = 1.0
	}
	if player.HasItem(mvdreader.IT_ARMOR2) {
		armorModifier = 1.5
	}
	if player.HasItem(mvdreader.IT_ARMOR3) {
		armorModifier = 2.0
	}
	quad := player.HasItem(mvdreader.IT_QUAD)
	pent := player.HasItem(mvdreader.IT_INVULNERABILITY)
	ring := player.HasItem(mvdreader.IT_INVISIBILITY)
	powerupStrength := 0.0
	if ring {
		powerupStrength = 200
	}
	if quad {
		powerupStrength += 500
	}
	if pent {
		powerupStrength += 800
	}
	if quad && pent {
		powerupStrength += 2
	}

	strength := float64(
		float64(player.Health) +
			(float64(player.Armor)*armorModifier)*2 +
			powerupStrength,
	)

	s.Strength = append(s.Strength, strength)
}

func (s *Stats) Death(parser *Parser, player, playerLastFrame *mvdreader.Player) {
	if s.currentRun == nil {
		return
	}
	s.currentRun.Stop = parser.mvd.State.Time
}

func (s *Stats) Kill(parser *Parser, player, playerLastFrame *mvdreader.Player) {
	if s.currentRun == nil {
		return
	}
	cr := s.currentRun
	cr.Kills = append(cr.Kills, RunKill{parser.mvd.State.Time})
}

func (s *Stats) Spawn(parser *Parser, player, playerLastFrame *mvdreader.Player) {
	cr := new(Run)
	s.currentRun = cr
	s.Runs = append(s.Runs, cr)
	cr.Start = parser.mvd.State.Time
}

func (s *Stats) Pickup(parser *Parser, player, playerLastFrame *mvdreader.Player, item mvdreader.IT_TYPE) {
	if s.currentRun == nil {
		return
	}

	cr := s.currentRun
	cr.Pickups = append(cr.Pickups, RunPickup{parser.mvd.State.Time, item})
}

func (s *Stats) Drop(parser *Parser, player, playerLastFrame *mvdreader.Player, item mvdreader.IT_TYPE) {
	if s.currentRun == nil {
		return
	}
}

func (s *Stats) Suicide(parser *Parser, player, playerLastFrame *mvdreader.Player) {
	if s.currentRun == nil {
		return
	}
}

func (s *Stats) Teamkill(parser *Parser, player, playerLastFrame *mvdreader.Player) {
	if s.currentRun == nil {
		return
	}
}
