package analysis

import (
	"sort"

	"github.com/adzik/pokemon/model"
	"github.com/adzik/pokemon/pkg/gamemaster"
)

func RatingFromScore(devScore float64) int {
	r := int(devScore / 10)
	if r < 1 {
		r = 1
	}
	if r > 10 {
		r = 10
	}
	return r
}

func Recommendation(p model.Pokemon, rating int) string {
	if p.IVPct == 100 {
		return "MAX OUT (L50)"
	}
	bestTier := gamemaster.BestTier(gamemaster.BestTier(p.RaidTier, p.GLTier), gamemaster.BestTier(p.ULTier, p.MLTier))
	if gamemaster.TierRank(bestTier) >= gamemaster.TierRank("S") && rating >= 7 {
		return "Inwestuj — meta-relevantny"
	}
	if p.IsShadow && rating >= 6 {
		return "Usun Frustration po eventach, potem inwestuj"
	}
	if p.IsLucky && rating >= 5 {
		return "Lucky — tanie ulepszanie, warto inwestowac"
	}
	if rating >= 5 {
		return "Dobra opcja do inwestycji"
	}
	if rating >= 3 {
		return "Trzymaj, ale nie inwestuj teraz"
	}
	return "Kandydat do transferu"
}

type RankingRow struct {
	Rank    int
	Pokemon model.Pokemon
}

type FullRankingResult struct {
	Rows      []RankingRow
	Total     int
	Excluded  int
}

// IsFullyDeveloped checks if a Pokemon has reached or exceeded its CP cap
// considering that the next power-up (+0.5 level) would also push it over.
func IsFullyDeveloped(p model.Pokemon, cap int) bool {
	if cap <= 0 {
		return false
	}
	if p.CP >= cap {
		return true
	}
	nextCP := NextPowerUpCP(p.BaseStats.Atk, p.BaseStats.Def, p.BaseStats.HP, p.IVAtk, p.IVDef, p.IVHP, p.CP)
	return nextCP > cap
}

func BuildFullRanking(c []model.Pokemon) FullRankingResult {
	result := FullRankingResult{Total: len(c)}
	for _, p := range c {
		cap := PurposeCap(p.Purpose, p.CelRozwoju, p.MaxCP)
		if IsFullyDeveloped(p, cap) {
			result.Excluded++
			continue
		}
		result.Rows = append(result.Rows, RankingRow{Pokemon: p})
	}
	sort.Slice(result.Rows, func(i, j int) bool {
		a, b := result.Rows[i].Pokemon, result.Rows[j].Pokemon
		if a.DevScore != b.DevScore {
			return a.DevScore > b.DevScore
		}
		if a.IVPct != b.IVPct {
			return a.IVPct > b.IVPct
		}
		return a.CP > b.CP
	})
	for i := range result.Rows {
		result.Rows[i].Rank = i + 1
	}
	return result
}
