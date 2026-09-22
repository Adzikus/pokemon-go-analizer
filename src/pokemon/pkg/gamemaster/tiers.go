package gamemaster

import (
	"strings"

	"github.com/adzik/pokemon/config"
)

func BuildMetaTiers(m *MetaData) map[string]string {
	tiers := make(map[string]string)

	for _, list := range [][]string{
		m.Rankings.GreatLeague,
		m.Rankings.UltraLeague,
		m.Rankings.MasterLeague,
	} {
		for pos, sid := range list {
			if existing, ok := tiers[sid]; ok && existing == "S" {
				continue
			}
			switch {
			case pos < config.Global.MetaTiers.S:
				tiers[sid] = "S"
			case pos < config.Global.MetaTiers.A:
				if tiers[sid] != "S" {
					tiers[sid] = "A"
				}
			case pos < config.Global.MetaTiers.B:
				if tiers[sid] == "" {
					tiers[sid] = "B"
				}
			default:
				if tiers[sid] == "" {
					tiers[sid] = "C"
				}
			}
		}
	}

	for pos, sid := range m.Rankings.Raid {
		if existing, ok := tiers[sid]; ok && existing == "S" {
			continue
		}
		switch {
		case pos < config.Global.MetaTiers.S:
			tiers[sid] = "S"
		case pos < config.Global.MetaTiers.A:
			if tiers[sid] != "S" {
				tiers[sid] = "A"
			}
		}
	}

	suffixes := []string{"", "_shadow", "_mega", "_mega_x", "_mega_y"}
	baseTiers := make(map[string]string)
	for sid, tier := range tiers {
		base := sid
		for _, sfx := range suffixes[1:] {
			base = strings.TrimSuffix(base, sfx)
		}
		if existing, ok := baseTiers[base]; !ok || tier == "S" || (tier == "A" && existing != "S") {
			baseTiers[base] = tier
		}
	}
	for base, tier := range baseTiers {
		for _, sfx := range suffixes {
			variant := base + sfx
			if _, ok := tiers[variant]; !ok {
				tiers[variant] = tier
			}
		}
	}

	return tiers
}

// RankToTier maps position in a ranking list to a tier following pokemongohub.net convention.
func RankToTier(pos int) string {
	switch {
	case pos < config.Global.Tiers.S:
		return "S"
	case pos < config.Global.Tiers.Ap:
		return "A+"
	case pos < config.Global.Tiers.A:
		return "A"
	case pos < config.Global.Tiers.Bp:
		return "B+"
	case pos < config.Global.Tiers.B:
		return "B"
	case pos < config.Global.Tiers.C:
		return "C"
	default:
		return "D"
	}
}

// TierToScore converts a tier string to a numeric score for DevScore computation.
func TierToScore(tier string) float64 {
	switch tier {
	case "S":
		return 20
	case "A+":
		return 17
	case "A":
		return 14
	case "B+":
		return 11
	case "B":
		return 8
	case "C":
		return 4
	default:
		return 0
	}
}

// TierRank returns ordering for tier comparison (higher = better).
func TierRank(tier string) int {
	switch tier {
	case "S":
		return 6
	case "A+":
		return 5
	case "A":
		return 4
	case "B+":
		return 3
	case "B":
		return 2
	case "C":
		return 1
	default:
		return 0
	}
}

// BestTier returns the better of two tiers.
func BestTier(a, b string) string {
	if TierRank(a) >= TierRank(b) {
		return a
	}
	return b
}

func BuildMetaTiersByCategory(m *MetaData) MetaTiersByCategory {
	result := MetaTiersByCategory{
		Raid:     make(map[string]string),
		GL:       make(map[string]string),
		UL:       make(map[string]string),
		ML:       make(map[string]string),
		RaidRank: make(map[string]int),
		GLRank:   make(map[string]int),
		ULRank:   make(map[string]int),
		MLRank:   make(map[string]int),
	}

	for pos, sid := range m.Rankings.Raid {
		result.Raid[sid] = RankToTier(pos)
		result.RaidRank[sid] = pos
	}
	for pos, sid := range m.Rankings.GreatLeague {
		result.GL[sid] = RankToTier(pos)
		result.GLRank[sid] = pos
	}
	for pos, sid := range m.Rankings.UltraLeague {
		result.UL[sid] = RankToTier(pos)
		result.ULRank[sid] = pos
	}
	for pos, sid := range m.Rankings.MasterLeague {
		result.ML[sid] = RankToTier(pos)
		result.MLRank[sid] = pos
	}

	suffixes := []string{"", "_shadow", "_mega", "_mega_x", "_mega_y"}
	type catPair struct {
		tiers map[string]string
		ranks map[string]int
	}
	categories := []catPair{
		{result.Raid, result.RaidRank},
		{result.GL, result.GLRank},
		{result.UL, result.ULRank},
		{result.ML, result.MLRank},
	}

	for _, cat := range categories {
		baseTiers := make(map[string]string)
		baseRanks := make(map[string]int)
		for sid, tier := range cat.tiers {
			base := sid
			for _, sfx := range suffixes[1:] {
				base = strings.TrimSuffix(base, sfx)
			}
			if existing, ok := baseTiers[base]; !ok || TierRank(tier) > TierRank(existing) {
				baseTiers[base] = tier
				baseRanks[base] = cat.ranks[sid]
			}
		}
		for base, tier := range baseTiers {
			for _, sfx := range suffixes {
				variant := base + sfx
				if _, ok := cat.tiers[variant]; !ok {
					cat.tiers[variant] = tier
					cat.ranks[variant] = baseRanks[base]
				}
			}
		}
	}

	return result
}

// BestMetaTier returns the best tier across all categories for a Pokemon.
func (m MetaTiersByCategory) BestMetaTier(speciesID string) string {
	best := "D"
	best = BestTier(best, m.Raid[speciesID])
	best = BestTier(best, m.GL[speciesID])
	best = BestTier(best, m.UL[speciesID])
	best = BestTier(best, m.ML[speciesID])
	return best
}
