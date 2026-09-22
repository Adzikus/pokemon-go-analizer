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

func GetLeagueCaps() map[string]int {
	if config.Global != nil && config.Global.PVP.LeagueCaps != nil {
		return config.Global.PVP.LeagueCaps
	}
	return map[string]int{"GL": 1500, "UL": 2500, "ML": 0}
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
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].DevScore > candidates[j].DevScore
	})

	var teams []PvPTeam
	for i := 0; i+2 < len(candidates) && len(teams) < 3; i += 3 {
		team := PvPTeam{
			Name:    fmt.Sprintf("Druzyna %c", 'A'+len(teams)),
			Style:   "balanced",
			Members: [3]model.Pokemon{candidates[i], candidates[i+1], candidates[i+2]},
		}
		teams = append(teams, team)
	}
	return teams
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
