package analysis

import (
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

func BuildFullRanking(c []model.Pokemon) []RankingRow {
	sorted := SortByScore(c)
	rows := make([]RankingRow, len(sorted))
	for i, p := range sorted {
		rows[i] = RankingRow{Rank: i + 1, Pokemon: p}
	}
	return rows
}
