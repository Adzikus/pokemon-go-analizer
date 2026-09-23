package analysis

import (
	"fmt"
	"sort"

	"github.com/adzik/pokemon/config"
	"github.com/adzik/pokemon/model"
	"github.com/adzik/pokemon/pkg/gamemaster"
)

type PvPTeam struct {
	Name    string
	Style   string
	Members [3]model.Pokemon
}

type pvpScored struct {
	pokemon   model.Pokemon
	score     float64
	typeSet   map[string]bool
	cpEff     float64
	typeBonus float64
}

func GetLeagueCaps() map[string]int {
	if config.Global != nil && config.Global.PVP.LeagueCaps != nil {
		return config.Global.PVP.LeagueCaps
	}
	return map[string]int{"GL": 1500, "UL": 2500, "ML": 0}
}

func LeagueTier(p model.Pokemon, league string) string {
	switch league {
	case "GL":
		return p.GLTier
	case "UL":
		return p.ULTier
	case "ML":
		return p.MLTier
	}
	return ""
}

func BuildPvPTeams(c []model.Pokemon, league string) []PvPTeam {
	cap := GetLeagueCaps()[league]

	var candidates []model.Pokemon
	for _, p := range c {
		if cap > 0 && p.CP > cap {
			continue
		}
		if cap > 0 && p.CP < cap-200 {
			continue
		}
		tier := LeagueTier(p, league)
		if gamemaster.TierRank(tier) >= gamemaster.TierRank("B") {
			candidates = append(candidates, p)
		}
	}

	type scored struct {
		pokemon  model.Pokemon
		score    float64
		typeSet  map[string]bool
		cpEff    float64
		typeBonus float64
	}

	var items []pvpScored
	for _, p := range candidates {
		var cpEff float64
		if cap > 0 {
			cpEff = float64(p.CP) / float64(cap)
			if cpEff > 1.0 {
				cpEff = 1.0
			}
		} else {
			cpEff = 1.0
		}

		typeSet := make(map[string]bool)
		for _, t := range p.Types {
			if t != "" {
				typeSet[t] = true
			}
		}

		var typeBonus float64
		if len(typeSet) > 1 {
			typeBonus = 5
		}

		item := pvpScored{
			pokemon:   p,
			cpEff:     cpEff,
			typeBonus: typeBonus,
			score:     p.DevScore + cpEff*10 + typeBonus,
		}
		items = append(items, item)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].score > items[j].score
	})

	used := make([]bool, len(items))
	var teams []PvPTeam

	for len(teams) < 3 {
		var teamMembers []pvpScored
		var teamTypes []map[string]bool

		for round := 0; round < 3; round++ {
			bestIdx := -1
			bestScore := -1.0

			for i, item := range items {
				if used[i] {
					continue
				}

				overlap := false
				for _, tt := range teamTypes {
					for t := range item.typeSet {
						if tt[t] {
							overlap = true
							break
						}
					}
					if overlap {
						break
					}
				}

				score := item.score
				if overlap {
					score -= 20
				}

				if score > bestScore {
					bestScore = score
					bestIdx = i
				}
			}

			if bestIdx == -1 {
				break
			}

			used[bestIdx] = true
			teamMembers = append(teamMembers, items[bestIdx])
			teamTypes = append(teamTypes, items[bestIdx].typeSet)
		}

		if len(teamMembers) < 3 {
			break
		}

		team := PvPTeam{
			Name:  fmt.Sprintf("Druzyna %c", 'A'+len(teams)),
			Style: detectStyle(teamMembers),
			Members: [3]model.Pokemon{
				teamMembers[0].pokemon,
				teamMembers[1].pokemon,
				teamMembers[2].pokemon,
			},
		}
		teams = append(teams, team)
	}

	return teams
}

func detectStyle(members []pvpScored) string {
	if len(members) == 0 {
		return "balanced"
	}

	var atkTotal, defTotal float64
	for _, m := range members {
		if m.pokemon.BaseStats.Atk > 0 {
			atkTotal += float64(m.pokemon.BaseStats.Atk)
			defTotal += float64(m.pokemon.BaseStats.Def)
		}
	}

	if atkTotal == 0 || defTotal == 0 {
		return "balanced"
	}

	ratio := atkTotal / defTotal
	if ratio > 1.2 {
		return "aggressive"
	}
	if ratio < 0.8 {
		return "defensive"
	}
	return "balanced"
}
