package main

import (
	"bytes"
	"fmt"
	//"os"
	"regexp"
	"strings"
)

type ktxPlayerInfo struct {
	ArmorMh     map[string]string
	WeaponStats map[string]string
	RL_Skills   map[string]string
	Powerups    map[string]string
	RL          map[string]string
	Damage      map[string]string
	Time        map[string]string
	Streaks     map[string]string
	Spawnfrags  string
	Frags       string
	Rank        string
	FriendKills string
	Efficiency  string
}

type ktxTeamInfo struct {
	WeaponStats map[string]string
	Powerups    map[string]string
	ArmorMh     map[string]string
	Rl          map[string]string
	Damage      map[string]string
	Time        map[string]string
	Frags       string
	Percentage  string
}

type ktx1bRegexsStruct struct {
	start_parsing           *regexp.Regexp
	team                    *regexp.Regexp
	player_frags            *regexp.Regexp
	player_wp               *regexp.Regexp
	player_stats_general    *regexp.Regexp
	player_stats_spawnfrags *regexp.Regexp
	team_stats_general      *regexp.Regexp
}

var ktx1bRegexs ktx1bRegexsStruct

var (
	ktx1b_message_player_statistics = []byte{
		10,
		208,
		236,
		225,
		249,
		229,
		242,
		32,
		243,
		244,
		225,
		244,
		233,
		243,
		244,
		233,
		227,
		243,
		58,
		10,
		157,
		158,
		158,
		158,
		158,
		158,
		158,
		158,
		158,
		158,
		158,
		158,
		158,
		158,
		158,
		158,
		158,
		158,
		158,
		158,
		158,
		158,
		158,
		158,
		158,
		158,
		158,
		158,
		158,
		158,
		158,
		158,
		158,
		158,
		158,
		158,
		158,
		159,
		10,
	}
	ktx1b_message_player_name_start         = []byte{135, 32}
	ktx1b_message_player_wp_stat            = []byte{215, 240, 58, 32}
	ktx1b_message_player_rl_skill           = []byte{210, 204, 32, 243, 235, 233, 236, 236, 58, 32}
	ktx1b_message_player_armor_mh           = []byte{193, 242, 237, 242, 166, 237, 232, 243, 58, 32}
	ktx1b_message_player_powerups           = []byte{208, 239, 247, 229, 242, 245, 240, 243, 58, 32}
	ktx1b_message_player_rl                 = []byte{32, 32, 32, 32, 32, 32, 210, 204, 58, 32}
	ktx1b_message_player_damage             = []byte{32, 32, 196, 225, 237, 225, 231, 229, 58, 32}
	ktx1b_message_player_time               = []byte{32, 32, 32, 32, 212, 233, 237, 229, 58, 32}
	ktx1b_message_player_streaks            = []byte{32, 211, 244, 242, 229, 225, 235, 243, 58, 32}
	ktx1b_message_player_spawnfrags         = []byte{32, 32, 211, 240, 225, 247, 238, 198, 242, 225, 231, 243, 58, 32}
	ktx1b_message_team_match_statistics_end = []byte{
		32,
		237,
		225,
		244,
		227,
		232,
		32,
		243,
		244,
		225,
		244,
		233,
		243,
		244,
		233,
		227,
		243,
		58,
		10,
	}
	ktx1b_message_team_match_statistics_start = []byte{58, 32, 215, 240, 58, 32}
	ktx1b_message_team_powerups               = []byte{208, 239, 247, 229, 242, 245, 240, 243, 58, 32}
	ktx1b_message_team_armormh                = []byte{
		193,
		242,
		237,
		242,
		166,
		237,
		232,
		243,
		58,
		32,
		231,
		225,
		58,
		48,
		32,
	}
	ktx1b_message_team_rl     = []byte{32, 32, 32, 32, 32, 32, 210, 204, 58, 32}
	ktx1b_message_team_damage = []byte{32, 32, 196, 225, 237, 225, 231, 229, 58, 32}
	ktx1b_message_team_time   = []byte{32, 32, 32, 32, 212, 233, 237, 229, 58, 32}

	ktx1b_message_top_scorers = []byte{
		32, 244, 239, 240, 32, 243, 227, 239, 242, 229, 242, 243, 58,
	}
)

func KTX_1_4b_Check(serverinfo map[string]string) bool {
	ktxver, found := serverinfo["ktxver"]
	if !found {
		return false
	}
	if strings.HasPrefix(ktxver, "1.40") { //"1.40-beta-quakecon-release3" {
		ktx1bRegexs.start_parsing = regexp.MustCompile("The match is over\n")
		ktx1bRegexs.team = regexp.MustCompile("Team (.*)\n")
		ktx1bRegexs.player_frags = regexp.MustCompile(
			`([+-]?\d+) \(([+-]?\d+)\) ([+-]?\d+) ([+-]?[0-9]*[.][0-9]+)%`,
		)
		ktx1bRegexs.player_wp = regexp.MustCompile(
			`([a-zA-Z]+)(\d+[.]\d+)`,
		)
		ktx1bRegexs.player_stats_general = regexp.MustCompile(
			`([a-zA-Z]+)[:]?(\d+[.]?[\d+]?)[%]?`,
		)
		ktx1bRegexs.team_stats_general = regexp.MustCompile(
			`(\d+[.]?[\d+]?)`,
		)
		ktx1bRegexs.player_stats_spawnfrags = regexp.MustCompile(
			`(\d+)`,
		)
		return true
	}

	return false
}

const (
	ktx1b_state_seeking = iota
	ktx1b_state_parsing
	ktx1b_state_parsing_team_name
	ktx1b_state_parsing_player
	ktx1b_state_parsing_team
	ktx1b_state_parsing_team_statistics
	ktx1b_state_parsing_top_scorers_frags
	ktx1b_state_parsing_top_scorers_deaths
	ktx1b_state_parsing_top_scorers_friendkills
	ktx1b_state_parsing_top_scorers_fragstreak
	ktx1b_state_parsing_top_scorers_quadrun
	ktx1b_state_parsing_top_scorers_rl_kill
	ktx1b_state_parsing_top_scorers_boomsticker
	ktx1b_state_parsing_top_scorers_survivor
	ktx1b_state_parsing_top_scorers_annihilator
	ktx1b_state_parsing_top_scorers
)

const (
	ktx_state_seeking = iota
	ktx_state_collecting
	ktx_state_player_stats
	ktx_state_player_stats_special
	ktx_state_team_statistics
	ktx_state_top_scorers
	ktx_state_team_scores
)

type KTXState struct {
	state  int
	buffer *bytes.Buffer

	Teams      []*ModParserTeam
	TopScorers map[string][]*ktx_top_scorer
}

func KTX_1_4b_Frame(p *Parser) error {
	mpState := p.mod_parser.State
	if mpState == nil {
		ks := new(KTXState)
		ks.buffer = bytes.NewBuffer([]byte{})
		ks.TopScorers = make(map[string][]*ktx_top_scorer)
		p.mod_parser.State = ks
	}

	state, ok := p.mod_parser.State.(*KTXState)
	if !ok {
		return fmt.Errorf("initialization failed")
	}

	// info := &p.mod_parser_info
	for _, m := range p.mvd.State.Messages {
		if m.From != 2 {
			continue
		}

		if state.state == ktx_state_seeking {
			if ktx1bRegexs.start_parsing.Match([]byte(m.Message)) {
				state.state = ktx_state_collecting
			}
			continue
		}
		_, err := state.buffer.Write([]byte(m.Message))
		if err != nil {
			return err
		}

	}
	return nil
}

type KTXfxes struct {
	prefix, suffix []byte
}

func (f *KTXfxes) Check(line []byte) bool {
	return bytes.HasPrefix(line, f.prefix) && bytes.HasSuffix(line, f.suffix)
}

func (f *KTXfxes) Get(line []byte) []byte {
	return line[len(f.prefix) : len(line)-len(f.suffix)]
}

var (
	ktxfxes_bar  = KTXfxes{[]byte{157, 158}, []byte{158, 159}}
	ktxfxes_team = KTXfxes{[]byte{
		84, 101, 97, 109, 32, 144,
	}, []byte{145, 58}}
	ktxfxes_player             = KTXfxes{[]byte{135, 32}, []byte{58}}
	ktx_line_player_statistics = []byte{
		208, 236, 225, 249, 229, 242, 32, 243, 244, 225, 244, 233, 243, 244, 233, 227, 243, 58,
	}
	ktx_line_player_statistics_special = []byte{
		70, 114, 97, 103, 115, 32, 40, 114, 97, 110, 107, 41, 32, 102, 114, 105, 101, 110, 100, 107, 105, 108, 108, 115, 32, 15, 32, 101, 102, 102, 105, 99, 105, 101, 110, 99, 121,
	}

	ktx_line_team_statistics_suffix = []byte{
		32,
		237,
		225,
		244,
		227,
		232,
		32,
		243,
		244,
		225,
		244,
		233,
		243,
		244,
		233,
		227,
		243,
		58,
	}
	ktx_top_scorers_frags = []byte{
		32, 32, 32, 32, 32, 32, 70, 114, 97, 103, 115, 58, 32,
	}

	ktx_top_scorers_deaths = []byte{
		32, 32, 32, 32, 32, 68, 101, 97, 116, 104, 115, 58, 32,
	}
	ktx_top_scorers_friendkills = []byte{
		70, 114, 105, 101, 110, 100, 107, 105, 108, 108, 115, 58, 32,
	}
	ktx_top_scorers_efficiency = []byte{
		32, 69, 102, 102, 105, 99, 105, 101, 110, 99, 121, 58, 32,
	}

	ktx_top_scorers_fragstreak = []byte{
		32, 70, 114, 97, 103, 83, 116, 114, 101, 97, 107, 58, 32,
	}
	ktx_top_scorers_quadrun = []byte{
		32, 32, 32, 32, 81, 117, 97, 100, 82, 117, 110, 58, 32,
	}
	ktx_top_scorers_rl_killer = []byte{
		32, 32, 82, 76, 32, 75, 105, 108, 108, 101, 114, 58, 32,
	}
	ktx_top_scorers_boomsticker = []byte{
		66, 111, 111, 109, 115, 116, 105, 99, 107, 101, 114, 58, 32,
	}
	ktx_top_scorers_survivor = []byte{
		32, 32, 32, 83, 117, 114, 118, 105, 118, 111, 114, 58, 32,
	}
	ktx_top_scorers_annihilator = []byte{
		65, 110, 110, 105, 104, 105, 108, 97, 116, 111, 114, 58, 32,
	}

	ktx_line_team_scores = []byte{
		212, 229, 225, 237, 32, 243, 227, 239, 242, 229, 243, 58, 32,
	}
)

type KTX_TopScorers struct {
	start []byte
	name  string
}

var ktx_top_scorers_list = []KTX_TopScorers{
	{ktx_top_scorers_frags, "frags"},
	{ktx_top_scorers_deaths, "deaths"},
	{ktx_top_scorers_friendkills, "friendkills"},
	{ktx_top_scorers_efficiency, "efficiency"},
	{ktx_top_scorers_fragstreak, "fragstreak"},
	{ktx_top_scorers_fragstreak, "fragstreak"},
	{ktx_top_scorers_quadrun, "quadrun"},
	{ktx_top_scorers_rl_killer, "rl_killer"},
	{ktx_top_scorers_boomsticker, "boomsticker"},
	{ktx_top_scorers_survivor, "survivor"},
	{ktx_top_scorers_annihilator, "annihilator"},
}

type ktx_top_scorer struct {
	Name  *ParserString
	Value string
}

func KTX_1_4b_End(p *Parser) error {
	var currentTeam *ModParserTeam
	var currentTeamInfo *ktxTeamInfo
	var currentPlayer *ModParserPlayer
	var currentPlayerInfo *ktxPlayerInfo
	var currentTopFraggerType string
	var topScorerIndex int

	state, ok := p.mod_parser.State.(*KTXState)
	if !ok {
		return fmt.Errorf("getting info failed")
	}
	lines := bytes.Split(state.buffer.Bytes(), []byte{10})
	for _, line := range lines {

		if len(line) == 0 {
			continue
		}
		line_converted := make([]byte, len(line))
		for i, by := range line {
			if by > 128 {
				by -= 128
			}
			line_converted[i] = by
		}

		// iknore brown bars
		if ktxfxes_bar.Check(line) {
			continue
		}
		// player statistics
		if state.state == ktx_state_collecting {
			if bytes.Equal(
				line,
				ktx_line_player_statistics,
			) {
				state.state = ktx_state_player_stats
				continue
			}
		}

		if bytes.HasPrefix(line, ktx_line_team_scores) {
			state.state = ktx_state_team_scores
			continue
		}

		if state.state == ktx_state_team_scores {
			index := bytes.Index(line, []byte{145, 58})
			if index >= 0 {
				found := false
				for _, t := range state.Teams {
					if bytes.Equal(t.Name.Byte, line[1:index]) {
						currentTeam = t

						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("could not find team: %s", line[1:index])
				}
				line = line[index:]
				line_converted = line_converted[index:]
				s := ktx1bRegexs.team_stats_general.FindAllSubmatch(line, -1)
				if len(s) == 2 {
					if _info, ok := currentTeam.Stat.(*ktxTeamInfo); ok {
						currentTeamInfo = _info
					} else {
						return fmt.Errorf("could not get team info")
					}

					currentTeamInfo.Frags = string(s[0][0])
					currentTeamInfo.Percentage = string(s[1][0])
					continue
				}

				print_message(line, line_converted)
				return fmt.Errorf("error parsing stats for team: %s", line[1:index])
			}
			continue
		}

		if state.state == ktx_state_top_scorers {
			for _, e := range ktx_top_scorers_list {
				if bytes.HasPrefix(line, e.start) {
					tfm := new([]*ktx_top_scorer)
					currentTopFraggerType = e.name
					state.TopScorers[e.name] = *tfm
					topScorerIndex = len(e.start)
				}
			}
			if currentTopFraggerType == "" {
				return fmt.Errorf("getting top scorer type failed")
			}

			line = line[topScorerIndex:]
			line_converted = line_converted[topScorerIndex:]

			index := bytes.LastIndex(line, []byte{32, 144})

			if index < 0 {
				return fmt.Errorf("top scorer parsing failed")
			}
			scorer := new(ktx_top_scorer)
			scorer.Name = p.ParserStringNew(line[:index])
			scorer.Value = string(line[index+2 : len(line)-1])
			l := state.TopScorers[currentTopFraggerType]
			l = append(l, scorer)
			state.TopScorers[currentTopFraggerType] = l
		}

		if state.state == ktx_state_team_statistics {
			index := bytes.Index(line, ktx1b_message_top_scorers)
			if index != -1 {
				state.state = ktx_state_top_scorers
				continue
			}
			if bytes.HasPrefix(line, []byte{144}) {
				index := bytes.Index(line, []byte{145, 58, 32})
				if index > 0 {
					name := line[1:index]
					found := false
					for _, t := range state.Teams {
						if bytes.Equal(t.Name.Byte, name) {
							found = true
							currentTeam = t

							if _info, ok := currentTeam.Stat.(*ktxTeamInfo); ok {
								currentTeamInfo = _info
							} else {
								return fmt.Errorf("could not get team info")
							}

							break
						}
					}
					if !found {
						return fmt.Errorf("could not find team (%s)", name)
					}
					line = line[index+1:]
					line_converted = line_converted[index+1:]

					if !get_stats(
						line,
						line_converted,
						ktx1b_message_team_match_statistics_start,
						*ktx1bRegexs.player_stats_general,
						&currentTeamInfo.WeaponStats) {
						return fmt.Errorf("could not parse team weapon stats for (%v)", currentTeam.Name)
					} else {
						continue
					}
				} else {
					print_message(line, line_converted)
					return fmt.Errorf("could not parse team stats")
				}

			}
			if get_stats(
				line,
				line_converted,
				ktx1b_message_team_powerups,
				*ktx1bRegexs.player_stats_general,
				&currentTeamInfo.Powerups,
			) {
				continue
			}

			if get_stats(
				line,
				line_converted,
				ktx1b_message_team_armormh,
				*ktx1bRegexs.player_stats_general,
				&currentTeamInfo.ArmorMh,
			) {
				continue
			}

			if get_stats(
				line,
				line_converted,
				ktx1b_message_team_powerups,
				*ktx1bRegexs.player_stats_general,
				&currentTeamInfo.Powerups,
			) {
				continue
			}

			if get_stats(
				line,
				line_converted,
				ktx1b_message_team_rl,
				*ktx1bRegexs.player_stats_general,
				&currentTeamInfo.Rl,
			) {
				continue
			}

			if get_stats(
				line,
				line_converted,
				ktx1b_message_team_time,
				*ktx1bRegexs.player_stats_general,
				&currentTeamInfo.Time,
			) {
				continue
			}

			if get_stats(
				line,
				line_converted,
				ktx1b_message_team_damage,
				*ktx1bRegexs.player_stats_general,
				&currentTeamInfo.Damage,
			) {
				continue
			}
			continue
		}

		if state.state == ktx_state_player_stats {
			if bytes.HasSuffix(line, ktx_line_team_statistics_suffix) {
				state.state = ktx_state_team_statistics
				continue
			}

			// get current team
			if ktxfxes_team.Check(line) {
				ti := new(ktxTeamInfo)
				t := p.ModParserTeamNew(ktxfxes_team.Get(line), ti)
				currentTeam = t
				currentTeamInfo = ti
				state.Teams = append(state.Teams, t)
				continue
			}

			// get current player
			if ktxfxes_player.Check(line) {
				pi := new(ktxPlayerInfo)
				p := p.ModParserPlayerNew(ktxfxes_player.Get(line), pi)
				currentPlayer = p
				currentPlayerInfo = pi
				currentTeam.PlayerAdd(p)
				state.state = ktx_state_player_stats_special
				continue
			}

			if currentPlayerInfo != nil {
				if currentPlayerInfo.WeaponStats == nil {
					if get_stats(
						line,
						line_converted,
						ktx1b_message_player_wp_stat,
						*ktx1bRegexs.player_stats_general,
						&currentPlayerInfo.WeaponStats,
					) {
						continue
					}
				}

				if currentPlayerInfo.RL_Skills == nil {
					if get_stats(
						line,
						line_converted,
						ktx1b_message_player_rl_skill,
						*ktx1bRegexs.player_stats_general,
						&currentPlayerInfo.RL_Skills,
					) {
						continue
					}
				}

				if currentPlayerInfo.ArmorMh == nil {
					if get_stats(
						line,
						line_converted,
						ktx1b_message_player_armor_mh,
						*ktx1bRegexs.player_stats_general,
						&currentPlayerInfo.ArmorMh,
					) {
						continue
					}
				}

				if currentPlayerInfo.Powerups == nil {
					if get_stats(
						line,
						line_converted,
						ktx1b_message_player_powerups,
						*ktx1bRegexs.player_stats_general,
						&currentPlayerInfo.Powerups,
					) {
						continue
					}
				}

				if currentPlayerInfo.RL == nil {
					if get_stats(
						line,
						line_converted,
						ktx1b_message_player_rl,
						*ktx1bRegexs.player_stats_general,
						&currentPlayerInfo.RL,
					) {
						continue
					}
				}

				if currentPlayerInfo.Damage == nil {
					if get_stats(
						line,
						line_converted,
						ktx1b_message_player_damage,
						*ktx1bRegexs.player_stats_general,
						&currentPlayerInfo.Damage,
					) {
						continue
					}
				}

				if currentPlayerInfo.Streaks == nil {
					if get_stats(
						line,
						line_converted,
						ktx1b_message_player_streaks,
						*ktx1bRegexs.player_stats_general,
						&currentPlayerInfo.Streaks,
					) {
						continue
					}
				}

				if currentPlayerInfo.Time == nil {
					if get_stats(
						line,
						line_converted,
						ktx1b_message_player_time,
						*ktx1bRegexs.player_stats_general,
						&currentPlayerInfo.Time,
					) {
						continue
					}
				}

				if len(line) > len(ktx1b_message_player_spawnfrags) {
					if bytes.Equal(
						line[:len(ktx1b_message_player_spawnfrags)],
						ktx1b_message_player_spawnfrags,
					) {
						s := ktx1bRegexs.player_stats_spawnfrags.FindSubmatch(line)
						if len(s) > 1 {
							currentPlayerInfo.Spawnfrags = string(s[0])
							continue
						}
					}
				}
			}

		}

		// handling the info straight after the player name
		if state.state == ktx_state_player_stats_special {
			s := ktx1bRegexs.player_frags.FindSubmatch(line)
			if len(s) == 5 {
				currentPlayerInfo.Frags = string(s[1])
				currentPlayerInfo.Rank = string(s[2])
				currentPlayerInfo.FriendKills = string(s[3])
				currentPlayerInfo.Efficiency = string(s[4])
				state.state = ktx_state_player_stats
				continue
			} else {
				return fmt.Errorf("could not get stats for player: %v", currentPlayer.Name)
			}
		}
		// print_message(line, line_converted)
	}
    for _, kt := range state.Teams {
        for _, kp := range kt.Players {
            pp := p.FindPlayer(kp.Name, kt.Name) 
            if pp != nil {
                pp.ModStats = kp.Stat
            }

        }
    }
	return nil
}

func print_message(message, message_converted []byte) {
	fmt.Println("Message:")
	m2 := ""
	fmt.Println(string(message_converted))
	for _, c := range message {
		m2 = m2 + fmt.Sprintf(" %3d", c)
	}
	fmt.Println(m2)
	fmt.Println(strings.Repeat("-", 20))
}

func get_stats(
	message []byte,
	message_converted []byte,
	prefix []byte,
	re regexp.Regexp,
	info *map[string]string,
) bool {
	var rs map[string]string
	if len(message) > len(prefix) {
		if bytes.HasPrefix(message, prefix) {
			s := re.FindAllSubmatch(message_converted[len(prefix):], -1)
			if len(s) > 0 {
				rs = make(map[string]string)
				for _, ss := range s {
					rs[string(ss[1])] = string(ss[2])
				}
				if info != nil {
					*info = rs
				}
				return true
			}
		}
	}
	return false
}
