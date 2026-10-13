package analysis

import (
	"testing"

	"github.com/adzik/pokemon/config"
	"github.com/adzik/pokemon/model"
)

func TestRatingFromScore_Zero(t *testing.T) {
	got := RatingFromScore(0)
	if got != 1 {
		t.Errorf("RatingFromScore(0) = %d, want 1", got)
	}
}

func TestRatingFromScore_Negative(t *testing.T) {
	got := RatingFromScore(-10)
	if got != 1 {
		t.Errorf("RatingFromScore(-10) = %d, want 1", got)
	}
}

func TestRatingFromScore_Low(t *testing.T) {
	tests := []struct {
		score float64
		want  int
	}{
		{0, 1},
		{5, 1},
		{10, 1},
		{15, 1},
		{25, 2},
		{50, 5},
		{75, 7},
		{99, 9},
	}
	for _, tt := range tests {
		got := RatingFromScore(tt.score)
		if got != tt.want {
			t.Errorf("RatingFromScore(%g) = %d, want %d", tt.score, got, tt.want)
		}
	}
}

func TestRatingFromScore_Max(t *testing.T) {
	got := RatingFromScore(100)
	if got != 10 {
		t.Errorf("RatingFromScore(100) = %d, want 10", got)
	}
}

func TestRatingFromScore_OverMax(t *testing.T) {
	got := RatingFromScore(200)
	if got != 10 {
		t.Errorf("RatingFromScore(200) = %d, want 10", got)
	}
}

func TestRecommendation_Hundo(t *testing.T) {
	config.Global = config.Default()
	p := model.Pokemon{
		PokemonRaw: model.PokemonRaw{},
		IVPct:      100,
	}
	got := Recommendation(p, 10)
	if got != "MAX OUT (L50)" {
		t.Errorf("Recommendation(hundo) = %q, want 'MAX OUT (L50)'", got)
	}
}

func TestRecommendation_MetaS(t *testing.T) {
	config.Global = config.Default()
	p := model.Pokemon{
		RaidTier: "S",
		IVPct:    96,
	}
	got := Recommendation(p, 8)
	if got != "Inwestuj — meta-relevantny" {
		t.Errorf("Recommendation(meta S) = %q", got)
	}
}

func TestRecommendation_MetaBestAcrossLeagues(t *testing.T) {
	config.Global = config.Default()
	// MLTier=S is best across all leagues
	p := model.Pokemon{
		MLTier: "S",
		IVPct:  96,
	}
	got := Recommendation(p, 7)
	if got != "Inwestuj — meta-relevantny" {
		t.Errorf("Recommendation(meta ML=S) = %q, want 'Inwestuj — meta-relevantny'", got)
	}
}

func TestRecommendation_Shadow(t *testing.T) {
	config.Global = config.Default()
	p := model.Pokemon{
		PokemonRaw: model.PokemonRaw{},
		IsShadow:   true,
		IVPct:      90,
	}
	got := Recommendation(p, 7)
	if got != "Usun Frustration po eventach, potem inwestuj" {
		t.Errorf("Recommendation(shadow) = %q", got)
	}
}

func TestRecommendation_Lucky(t *testing.T) {
	config.Global = config.Default()
	p := model.Pokemon{
		IsLucky: true,
		IVPct:   93,
	}
	got := Recommendation(p, 6)
	if got != "Lucky — tanie ulepszanie, warto inwestowac" {
		t.Errorf("Recommendation(lucky) = %q", got)
	}
}

func TestRecommendation_Good(t *testing.T) {
	config.Global = config.Default()
	p := model.Pokemon{IVPct: 90}
	got := Recommendation(p, 5)
	if got != "Dobra opcja do inwestycji" {
		t.Errorf("Recommendation(good) = %q", got)
	}
}

func TestRecommendation_Hold(t *testing.T) {
	config.Global = config.Default()
	p := model.Pokemon{IVPct: 80}
	got := Recommendation(p, 3)
	if got != "Trzymaj, ale nie inwestuj teraz" {
		t.Errorf("Recommendation(hold) = %q", got)
	}
}

func TestRecommendation_Trash(t *testing.T) {
	config.Global = config.Default()
	p := model.Pokemon{IVPct: 50}
	got := Recommendation(p, 1)
	if got != "Kandydat do transferu" {
		t.Errorf("Recommendation(trash) = %q", got)
	}
}

func TestBuildFullRanking_Sorted(t *testing.T) {
	config.Global = config.Default()
	c := []model.Pokemon{
		{PokemonRaw: model.PokemonRaw{UID: "P-1"}, DevScore: 50},
		{PokemonRaw: model.PokemonRaw{UID: "P-2"}, DevScore: 90},
		{PokemonRaw: model.PokemonRaw{UID: "P-3"}, DevScore: 70},
	}

	result := BuildFullRanking(c)
	if len(result.Rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(result.Rows))
	}
	if result.Rows[0].Rank != 1 || result.Rows[0].Pokemon.UID != "P-2" {
		t.Errorf("rank 1 = %s, want P-2", result.Rows[0].Pokemon.UID)
	}
	if result.Rows[1].Rank != 2 || result.Rows[1].Pokemon.UID != "P-3" {
		t.Errorf("rank 2 = %s, want P-3", result.Rows[1].Pokemon.UID)
	}
	if result.Rows[2].Rank != 3 || result.Rows[2].Pokemon.UID != "P-1" {
		t.Errorf("rank 3 = %s, want P-1", result.Rows[2].Pokemon.UID)
	}
}

func TestBuildFullRanking_Empty(t *testing.T) {
	result := BuildFullRanking(nil)
	if len(result.Rows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(result.Rows))
	}
	if result.Total != 0 {
		t.Errorf("expected Total 0, got %d", result.Total)
	}
}

func TestBuildFullRanking_DoesNotMutateOriginal(t *testing.T) {
	config.Global = config.Default()
	c := []model.Pokemon{
		{PokemonRaw: model.PokemonRaw{UID: "P-1"}, DevScore: 50},
		{PokemonRaw: model.PokemonRaw{UID: "P-2"}, DevScore: 90},
	}
	origUID := c[0].UID
	origScore := c[0].DevScore

	BuildFullRanking(c)

	if c[0].UID != origUID || c[0].DevScore != origScore {
		t.Error("BuildFullRanking mutated original")
	}
}
