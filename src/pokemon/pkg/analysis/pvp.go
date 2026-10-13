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
	pokemon     model.Pokemon
	score       float64
	typeSet     map[string]bool
	moveTypeSet map[string]bool
	cpEff       float64
	typeBonus   float64
	tierBonus   float64
	weaknesses  map[string]bool
	covers      map[string]bool
}

var tierBonusMap = map[string]float64{
	"S": 10, "A+": 8, "A": 6, "B+": 4, "B": 2, "C": 1, "D": 0,
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

func typeWeaknesses(types []string) map[string]bool {
	weaknesses := map[string]bool{}
	for _, t := range types {
		for _, w := range CounterType(t) {
			weaknesses[w] = true
		}
	}
	return weaknesses
}

func typeCovers(types []string) map[string]bool {
	covers := map[string]bool{}
	for _, t := range types {
		for _, c := range SuperEffective[t] {
			covers[c] = true
		}
	}
	return covers
}

func moveTypes(p model.Pokemon, moveMap map[string]*gamemaster.MoveEntry) map[string]bool {
	types := map[string]bool{}
	for _, moveName := range []string{p.FastMove, p.Charged1, p.Charged2} {
		if moveName == "" {
			continue
		}
		entry := moveMap[normalizeMoveID(moveName)]
		if entry != nil {
			types[entry.Type] = true
		}
	}
	return types
}

func allTypes(typeSet, moveTypeSet map[string]bool) map[string]bool {
	combined := map[string]bool{}
	for t := range typeSet {
		combined[t] = true
	}
	for t := range moveTypeSet {
		combined[t] = true
	}
	return combined
}

func combinedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func BuildPvPTeams(c []model.Pokemon, league string, moveMap map[string]*gamemaster.MoveEntry) []PvPTeam {
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
		if gamemaster.TierRank(tier) >= gamemaster.TierRank("D") {
			candidates = append(candidates, p)
		}
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

		moveTypeSet := moveTypes(p, moveMap)

		var typeBonus float64
		if len(typeSet) > 1 {
			typeBonus = 5
		}

		combined := allTypes(typeSet, moveTypeSet)

		tb := tierBonusMap[LeagueTier(p, league)]

		item := pvpScored{
			pokemon:     p,
			cpEff:       cpEff,
			typeSet:     typeSet,
			moveTypeSet: moveTypeSet,
			typeBonus:   typeBonus,
			tierBonus:   tb,
			score:       p.DevScore + cpEff*10 + typeBonus + tb,
			weaknesses:  typeWeaknesses(p.Types),
			covers:      typeCovers(combinedKeys(combined)),
		}
		items = append(items, item)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].score > items[j].score
	})

	roleMultipliers := [3]float64{0.8, 1.0, 1.2}

	used := make([]bool, len(items))
	var teams []PvPTeam

	for len(teams) < 3 {
		var teamMembers []pvpScored
		teamTypeCounts := map[string]int{}
		teamWeaknesses := map[string]bool{}

		for round := 0; round < 3; round++ {
			bestIdx := -1
			bestScore := -1.0

			for i, item := range items {
				if used[i] {
					continue
				}

				// Hard cap: max 2 of any single type (combined = pokemon type + move types)
				combined := allTypes(item.typeSet, item.moveTypeSet)
				skip := false
				for t := range combined {
					if teamTypeCounts[t] >= 2 {
						skip = true
						break
					}
				}
				if skip {
					continue
				}

				// Type overlap penalty (combined)
				overlap := 0
				for t := range combined {
					overlap += teamTypeCounts[t]
				}

				// Weakness overlap (exponential)
				weaknessOverlap := 0
				for w := range teamWeaknesses {
					if item.weaknesses[w] {
						weaknessOverlap++
					}
				}

				// Coverage bonus
				coverageCount := 0
				for w := range teamWeaknesses {
					if item.covers[w] {
						coverageCount++
					}
				}

				score := item.score
				score -= float64(overlap) * 20.0
				score -= weaknessExponential(weaknessOverlap)
				score += float64(coverageCount) * 5.0
				score *= roleMultipliers[round]

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
			for t := range allTypes(items[bestIdx].typeSet, items[bestIdx].moveTypeSet) {
				teamTypeCounts[t]++
			}
			for w := range items[bestIdx].weaknesses {
				teamWeaknesses[w] = true
			}
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

func weaknessExponential(overlap int) float64 {
	switch overlap {
	case 0:
		return 0
	case 1:
		return 10
	case 2:
		return 30
	case 3:
		return 60
	default:
		return 100
	}
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
