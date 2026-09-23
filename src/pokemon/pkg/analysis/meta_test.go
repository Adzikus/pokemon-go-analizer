package analysis

import (
	"strings"
	"testing"

	"github.com/adzik/pokemon/config"
	"github.com/adzik/pokemon/model"
)

func TestBuildMetaComparison_OwnedPokemon(t *testing.T) {
	config.Global = config.Default()
	gmMap := map[string]*model.GamemasterPokemon{
		"bulbasaur": {SpeciesID: "bulbasaur", SpeciesName: "Bulbasaur"},
	}
	c := []model.Pokemon{
		{
			PokemonRaw: model.PokemonRaw{UID: "P-1", Species: "Bulbasaur", CP: 1500},
			SpeciesID:  "bulbasaur",
			IVPct:      96,
			DevScore:   85,
		},
	}

	rows := BuildMetaComparison(c, []string{"bulbasaur"}, gmMap)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if !rows[0].Owned {
		t.Error("expected Owned=true")
	}
	if rows[0].Copies != 1 {
		t.Errorf("Copies = %d, want 1", rows[0].Copies)
	}
}

func TestBuildMetaComparison_NotOwned(t *testing.T) {
	config.Global = config.Default()
	gmMap := map[string]*model.GamemasterPokemon{
		"charizard": {SpeciesID: "charizard", SpeciesName: "Charizard"},
	}

	rows := BuildMetaComparison(nil, []string{"charizard"}, gmMap)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0].Owned {
		t.Error("expected Owned=false")
	}
	if rows[0].Copies != 0 {
		t.Errorf("Copies = %d, want 0", rows[0].Copies)
	}
}

func TestBuildMetaComparison_MultipleCopies(t *testing.T) {
	config.Global = config.Default()
	gmMap := map[string]*model.GamemasterPokemon{
		"bulbasaur": {SpeciesID: "bulbasaur", SpeciesName: "Bulbasaur"},
	}
	c := []model.Pokemon{
		{PokemonRaw: model.PokemonRaw{Species: "Bulbasaur", CP: 1500}, SpeciesID: "bulbasaur", IVPct: 93},
		{PokemonRaw: model.PokemonRaw{Species: "Bulbasaur", CP: 1200}, SpeciesID: "bulbasaur", IVPct: 96},
		{PokemonRaw: model.PokemonRaw{Species: "Bulbasaur", CP: 1800}, SpeciesID: "bulbasaur", IVPct: 100},
	}

	rows := BuildMetaComparison(c, []string{"bulbasaur"}, gmMap)
	if rows[0].Copies != 3 {
		t.Errorf("Copies = %d, want 3", rows[0].Copies)
	}
}

func TestBuildMetaComparison_BestIVSelected(t *testing.T) {
	config.Global = config.Default()
	gmMap := map[string]*model.GamemasterPokemon{
		"bulbasaur": {SpeciesID: "bulbasaur", SpeciesName: "Bulbasaur"},
	}
	c := []model.Pokemon{
		{PokemonRaw: model.PokemonRaw{Species: "Bulbasaur", CP: 1200}, SpeciesID: "bulbasaur", IVPct: 93},
		{PokemonRaw: model.PokemonRaw{Species: "Bulbasaur", CP: 1800}, SpeciesID: "bulbasaur", IVPct: 100},
	}

	rows := BuildMetaComparison(c, []string{"bulbasaur"}, gmMap)
	if rows[0].BestIV != 100 {
		t.Errorf("BestIV = %.1f, want 100", rows[0].BestIV)
	}
	if rows[0].BestCP != 1800 {
		t.Errorf("BestCP = %d, want 1800", rows[0].BestCP)
	}
}

func TestBuildMetaComparison_BestIVTieBreakByCP(t *testing.T) {
	config.Global = config.Default()
	gmMap := map[string]*model.GamemasterPokemon{
		"bulbasaur": {SpeciesID: "bulbasaur", SpeciesName: "Bulbasaur"},
	}
	c := []model.Pokemon{
		{PokemonRaw: model.PokemonRaw{Species: "Bulbasaur", CP: 1200}, SpeciesID: "bulbasaur", IVPct: 100},
		{PokemonRaw: model.PokemonRaw{Species: "Bulbasaur", CP: 2000}, SpeciesID: "bulbasaur", IVPct: 100},
	}

	rows := BuildMetaComparison(c, []string{"bulbasaur"}, gmMap)
	if rows[0].BestCP != 2000 {
		t.Errorf("BestCP = %d, want 2000 (tie break by CP)", rows[0].BestCP)
	}
}

func TestBuildMetaComparison_TierFromRankToTier(t *testing.T) {
	config.Global = config.Default()
	gmMap := map[string]*model.GamemasterPokemon{
		"a": {SpeciesID: "a"},
		"b": {SpeciesID: "b"},
		"c": {SpeciesID: "c"},
		"d": {SpeciesID: "d"},
	}
	attackers := []string{"a", "b", "c", "d"}

	rows := BuildMetaComparison(nil, attackers, gmMap)

	// RankToTier uses Tiers config: S < 20, A+ < 50, A < 100, ...
	// pos 0-3 → S (since 0 < 20, 1 < 20, 2 < 20, 3 < 20)
	for i, row := range rows {
		if row.Tier != "S" {
			t.Errorf("row %d Tier = %q, want S (pos %d < 20)", i, row.Tier, i)
		}
	}
}

func TestBuildMetaComparison_LimitFromConfig(t *testing.T) {
	config.Global = config.Default()
	config.Global.Meta.TopN = 3
	gmMap := map[string]*model.GamemasterPokemon{
		"a": {SpeciesID: "a"},
		"b": {SpeciesID: "b"},
		"c": {SpeciesID: "c"},
		"d": {SpeciesID: "d"},
		"e": {SpeciesID: "e"},
	}

	rows := BuildMetaComparison(nil, []string{"a", "b", "c", "d", "e"}, gmMap)
	if len(rows) != 3 {
		t.Errorf("expected 3 rows (TopN=3), got %d", len(rows))
	}
}

func TestBuildMetaComparison_GamemasterNameNotOverwritten(t *testing.T) {
	config.Global = config.Default()
	gmMap := map[string]*model.GamemasterPokemon{
		"bulbasaur": {SpeciesID: "bulbasaur", SpeciesName: "Bulbasaur"},
	}
	c := []model.Pokemon{
		{PokemonRaw: model.PokemonRaw{Species: "Groszek", CP: 1500}, SpeciesID: "bulbasaur", IVPct: 100},
	}

	rows := BuildMetaComparison(c, []string{"bulbasaur"}, gmMap)
	if rows[0].Name != "Bulbasaur" {
		t.Errorf("Name = %q, want 'Bulbasaur' (gamemaster name), not 'Groszek'", rows[0].Name)
	}
}

func TestBuildMetaComparison_NoGamemasterEntry(t *testing.T) {
	config.Global = config.Default()
	gmMap := map[string]*model.GamemasterPokemon{}

	rows := BuildMetaComparison(nil, []string{"unknown_species"}, gmMap)
	if rows[0].Name != "unknown_species" {
		t.Errorf("Name = %q, want 'unknown_species' (fallback to ID)", rows[0].Name)
	}
}

func TestFormatMetaComparison_AllOwned(t *testing.T) {
	rows := []MetaCompareRow{
		{Name: "Bulbasaur", Position: 1, Tier: "S", Owned: true, Copies: 2, BestIV: 100, IsLucky: true, Score: 90},
		{Name: "Charizard", Position: 2, Tier: "S", Owned: true, Copies: 1, BestIV: 96, IsLucky: false, Score: 80},
	}

	result := FormatMetaComparison(rows)
	if !strings.Contains(result, "2/2 (100%)") {
		t.Errorf("expected 100%% coverage, got: %s", result)
	}
	if strings.Contains(result, "brak") {
		t.Error("expected no missing list")
	}
}

func TestFormatMetaComparison_SomeMissing(t *testing.T) {
	rows := []MetaCompareRow{
		{Name: "Bulbasaur", Position: 1, Tier: "S", Owned: true, Copies: 2, BestIV: 100, IsLucky: true, Score: 90},
		{Name: "Charizard", Position: 2, Tier: "S", Owned: false},
		{Name: "Blastoise", Position: 3, Tier: "A", Owned: true, Copies: 1, BestIV: 93, IsLucky: false, Score: 70},
	}

	result := FormatMetaComparison(rows)
	if !strings.Contains(result, "2/3 (67%)") {
		t.Errorf("expected 67%% coverage, got: %s", result)
	}
	if !strings.Contains(result, "brak: Charizard") {
		t.Errorf("expected missing list with Charizard, got: %s", result)
	}
}

func TestFormatMetaComparison_NoneOwned(t *testing.T) {
	rows := []MetaCompareRow{
		{Name: "Mewtwo", Position: 1, Tier: "S", Owned: false},
		{Name: "Rayquaza", Position: 2, Tier: "S", Owned: false},
	}

	result := FormatMetaComparison(rows)
	if !strings.Contains(result, "0/2 (0%)") {
		t.Errorf("expected 0%% coverage, got: %s", result)
	}
	if !strings.Contains(result, "brak: Mewtwo, Rayquaza") {
		t.Errorf("expected missing list, got: %s", result)
	}
}

func TestBuildMetaComparison_LuckyDetection(t *testing.T) {
	config.Global = config.Default()
	gmMap := map[string]*model.GamemasterPokemon{
		"bulbasaur": {SpeciesID: "bulbasaur", SpeciesName: "Bulbasaur"},
	}
	c := []model.Pokemon{
		{PokemonRaw: model.PokemonRaw{Species: "Bulbasaur", CP: 1500}, SpeciesID: "bulbasaur", IVPct: 93, IsLucky: true},
		{PokemonRaw: model.PokemonRaw{Species: "Bulbasaur", CP: 1800}, SpeciesID: "bulbasaur", IVPct: 100},
	}

	rows := BuildMetaComparison(c, []string{"bulbasaur"}, gmMap)
	// Best IV = 100 (not lucky), so IsLucky = false
	if rows[0].IsLucky {
		t.Error("Best IV=100 is not lucky, IsLucky should be false")
	}
}

func TestBuildMetaComparison_RankToTierTransitions(t *testing.T) {
	config.Global = config.Default()
	config.Global.Meta.TopN = 55
	gmMap := map[string]*model.GamemasterPokemon{}
	attackers := make([]string, 55)
	for i := range attackers {
		attackers[i] = "species_" + string(rune('a'+i%26))
	}

	rows := BuildMetaComparison(nil, attackers, gmMap)

	// Verify tier transitions match RankToTier
	expectedTiers := map[int]string{
		0:  "S",  // pos 0 < 20
		19: "S",  // pos 19 < 20
		20: "A+", // pos 20 < 50
		49: "A+", // pos 49 < 50
		50: "A",  // pos 50 < 100
		54: "A",  // pos 54 < 100
	}

	for pos, expectedTier := range expectedTiers {
		if rows[pos].Tier != expectedTier {
			t.Errorf("pos %d Tier = %q, want %q", pos, rows[pos].Tier, expectedTier)
		}
	}
}
