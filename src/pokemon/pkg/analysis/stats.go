package analysis

import (
	"strings"

	"github.com/adzik/pokemon/model"
	"github.com/adzik/pokemon/pkg/gamemaster"
)

type CollectionStats struct {
	Total        int
	Shiny        int
	Lucky        int
	Shadow       int
	Purified     int
	Dynamax      int
	Traded       int
	Hundo        int
	HighIV       int
	Families     int
	TierS        int
	TierA        int
	TierB        int
	TierC        int
	PurposeRaid  int
	PurposePvP   int
	PurposeColl  int
	PurposeOther int
}

func ComputeStats(c []model.Pokemon) CollectionStats {
	var s CollectionStats
	s.Total = len(c)
	seenFamilies := map[string]bool{}

	for _, p := range c {
		if p.IsShiny {
			s.Shiny++
		}
		if p.IsLucky {
			s.Lucky++
		}
		if p.IsShadow {
			s.Shadow++
		}
		if p.IsPurified {
			s.Purified++
		}
		if p.IsDynamax {
			s.Dynamax++
		}
		if p.IsTraded {
			s.Traded++
		}
		if p.IVPct == 100.0 {
			s.Hundo++
		}
		if p.IVPct >= 90 {
			s.HighIV++
		}
		if p.FamilyID != "" {
			seenFamilies[p.FamilyID] = true
		}
		bestTier := gamemaster.BestTier(gamemaster.BestTier(p.RaidTier, p.GLTier), gamemaster.BestTier(p.ULTier, p.MLTier))
		switch bestTier {
		case "S":
			s.TierS++
		case "A+", "A":
			s.TierA++
		case "B+", "B":
			s.TierB++
		default:
			s.TierC++
		}
		purpose := strings.ToLower(p.Purpose)
		if purpose == "rajdy" || purpose == "raid" {
			s.PurposeRaid++
		} else if purpose == "pvp" || purpose == "gl" || purpose == "ul" || purpose == "ml" {
			s.PurposePvP++
		} else if purpose == "kolekcja" {
			s.PurposeColl++
		} else {
			s.PurposeOther++
		}
	}
	s.Families = len(seenFamilies)
	return s
}

func TopN(c []model.Pokemon, n int) []model.Pokemon {
	if len(c) <= n {
		out := make([]model.Pokemon, len(c))
		copy(out, c)
		return out
	}
	out := make([]model.Pokemon, n)
	copy(out, c[:n])
	return out
}

func SortByCP(c []model.Pokemon) []model.Pokemon {
	sorted := make([]model.Pokemon, len(c))
	copy(sorted, c)
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].CP > sorted[i].CP {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	return sorted
}
