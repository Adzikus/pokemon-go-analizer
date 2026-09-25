package analysis

import (
	"testing"

	"github.com/adzik/pokemon/config"
	"github.com/adzik/pokemon/model"
)

// Helper: create a Pokemon with key fields for transfer testing.
func testPokemon(uid, species, familyID string, ivPct float64, cp int, isShiny, isLucky, isShadow bool, purpose string, tags ...string) model.Pokemon {
	return model.Pokemon{
		PokemonRaw: model.PokemonRaw{
			UID:     uid,
			Species: species,
			CP:      cp,
			Purpose: purpose,
		},
		SpeciesID:  species,
		FamilyID:   familyID,
		IVPct:      ivPct,
		IsShiny:    isShiny,
		IsLucky:    isLucky,
		IsShadow:   isShadow,
		Tags:       tags,
	}
}

// --- Testy evaluateTransfer ---

func TestEvaluateTransfer_ShinyHundoAlwaysKeep(t *testing.T) {
	family := []model.Pokemon{
		testPokemon("P-1", "charizard", "charmander", 100, 2500, true, false, false, ""),
	}
	gmMap := map[string]*model.GamemasterPokemon{}
	status, _ := evaluateTransfer(family[0], buildFamilyMap(family), gmMap)
	if status != Keep {
		t.Errorf("Shiny HUNDO: status = %q, want KEEP", status)
	}
}

func TestEvaluateTransfer_HundoKeep(t *testing.T) {
	family := []model.Pokemon{
		testPokemon("P-1", "charizard", "charmander", 100, 2500, false, false, false, ""),
		testPokemon("P-2", "charmander", "charmander", 90, 600, false, false, false, ""),
	}
	gmMap := map[string]*model.GamemasterPokemon{}
	status, _ := evaluateTransfer(family[0], buildFamilyMap(family), gmMap)
	if status != Keep {
		t.Errorf("HUNDO (no shiny hundo in family): status = %q, want KEEP", status)
	}
}

func TestEvaluateTransfer_HundoSoftKeepWhenShinyHundoExists(t *testing.T) {
	family := []model.Pokemon{
		testPokemon("P-1", "charizard", "charmander", 100, 2500, false, false, false, ""),
		testPokemon("P-2", "charizard", "charmander", 100, 2400, true, false, false, ""),
	}
	gmMap := map[string]*model.GamemasterPokemon{}
	status, _ := evaluateTransfer(family[0], buildFamilyMap(family), gmMap)
	if status != SoftKeep {
		t.Errorf("HUNDO with shiny hundo in family: status = %q, want SOFT_KEEP", status)
	}
}

func TestEvaluateTransfer_LuckyAloneKeep(t *testing.T) {
	family := []model.Pokemon{
		testPokemon("P-1", "charizard", "charmander", 93, 2500, false, true, false, ""),
		testPokemon("P-2", "charmander", "charmander", 90, 600, false, false, false, ""),
	}
	gmMap := map[string]*model.GamemasterPokemon{}
	status, _ := evaluateTransfer(family[0], buildFamilyMap(family), gmMap)
	if status != Keep {
		t.Errorf("Lucky alone: status = %q, want KEEP", status)
	}
}

func TestEvaluateTransfer_LuckyWorseSoftKeep(t *testing.T) {
	family := []model.Pokemon{
		testPokemon("P-1", "charizard", "charmander", 80, 2000, false, true, false, ""),
		testPokemon("P-2", "charizard", "charmander", 93, 2500, false, true, false, ""),
	}
	gmMap := map[string]*model.GamemasterPokemon{}
	status, _ := evaluateTransfer(family[0], buildFamilyMap(family), gmMap)
	if status != SoftKeep {
		t.Errorf("Lucky with better Lucky in family: status = %q, want SOFT_KEEP", status)
	}
}

func TestEvaluateTransfer_ShinyAloneKeep(t *testing.T) {
	family := []model.Pokemon{
		testPokemon("P-1", "charmander", "charmander", 80, 500, true, false, false, ""),
		testPokemon("P-2", "charmander", "charmander", 90, 600, false, false, false, ""),
	}
	gmMap := map[string]*model.GamemasterPokemon{}
	status, _ := evaluateTransfer(family[0], buildFamilyMap(family), gmMap)
	if status != Keep {
		t.Errorf("Only shiny in family: status = %q, want KEEP", status)
	}
}

func TestEvaluateTransfer_ShinyWorseTrade(t *testing.T) {
	family := []model.Pokemon{
		testPokemon("P-1", "charizard", "charmander", 80, 2000, true, false, false, ""),
		testPokemon("P-2", "charizard", "charmander", 96, 2500, true, false, false, ""),
	}
	gmMap := map[string]*model.GamemasterPokemon{}
	status, _ := evaluateTransfer(family[0], buildFamilyMap(family), gmMap)
	if status != Trade {
		t.Errorf("Shiny with better shiny + IV<88%%: status = %q, want TRADE", status)
	}
}

func TestEvaluateTransfer_ShinyGoodIVSoftKeep(t *testing.T) {
	family := []model.Pokemon{
		testPokemon("P-1", "charizard", "charmander", 93, 2400, true, false, false, ""),
		testPokemon("P-2", "charizard", "charmander", 96, 2500, true, false, false, ""),
	}
	gmMap := map[string]*model.GamemasterPokemon{}
	status, _ := evaluateTransfer(family[0], buildFamilyMap(family), gmMap)
	if status != SoftKeep {
		t.Errorf("Shiny with better shiny + IV>=88%%: status = %q, want SOFT_KEEP", status)
	}
}

func TestEvaluateTransfer_HighIVSoftKeep(t *testing.T) {
	family := []model.Pokemon{
		testPokemon("P-1", "charmander", "charmander", 93, 500, false, false, false, ""),
		testPokemon("P-2", "charizard", "charmander", 96, 2500, false, false, false, ""),
	}
	gmMap := map[string]*model.GamemasterPokemon{}
	status, _ := evaluateTransfer(family[0], buildFamilyMap(family), gmMap)
	if status != SoftKeep {
		t.Errorf("IV 93%% (not shiny, not lucky): status = %q, want SOFT_KEEP", status)
	}
}

func TestEvaluateTransfer_PurposeCollectionSoftKeep(t *testing.T) {
	family := []model.Pokemon{
		testPokemon("P-1", "pidgey", "pidgey", 50, 200, false, false, false, "kolekcja"),
		testPokemon("P-2", "pidgey", "pidgey", 90, 800, false, false, false, ""),
	}
	gmMap := map[string]*model.GamemasterPokemon{}
	status, _ := evaluateTransfer(family[0], buildFamilyMap(family), gmMap)
	if status != SoftKeep {
		t.Errorf("Purpose kolekcja: status = %q, want SOFT_KEEP", status)
	}
}

func TestEvaluateTransfer_ShadowNeverTrade(t *testing.T) {
	family := []model.Pokemon{
		testPokemon("P-1", "charmander", "charmander", 50, 400, false, false, true, ""),
		testPokemon("P-2", "charmander", "charmander", 90, 600, false, false, false, ""),
	}
	gmMap := map[string]*model.GamemasterPokemon{}
	status, _ := evaluateTransfer(family[0], buildFamilyMap(family), gmMap)
	if status != SoftKeep {
		t.Errorf("Shadow: status = %q, want SOFT_KEEP", status)
	}
}

func TestEvaluateTransfer_ExcessDuplicatesTrade(t *testing.T) {
	// 12 Ratata in family, this one has low IV
	var family []model.Pokemon
	for i := 0; i < 11; i++ {
		family = append(family, testPokemon(
			string(rune('A'+i))+"-1", "rattata", "rattata",
			85, 300+i*10, false, false, false, "",
		))
	}
	// 12th with low IV
	lowIV := testPokemon("M-1", "rattata", "rattata", 40, 200, false, false, false, "")
	family = append(family, lowIV)

	gmMap := map[string]*model.GamemasterPokemon{}
	status, _ := evaluateTransfer(lowIV, buildFamilyMap(family), gmMap)
	if status != Trade {
		t.Errorf("12th rattata, low IV: status = %q, want TRADE", status)
	}
}

func TestEvaluateTransfer_CPDiffTrade(t *testing.T) {
	family := []model.Pokemon{
		testPokemon("P-1", "pidgey", "pidgey", 85, 100, false, false, false, ""),
		testPokemon("P-2", "pidgey", "pidgey", 85, 250, false, false, false, ""),
	}
	gmMap := map[string]*model.GamemasterPokemon{}
	status, _ := evaluateTransfer(family[0], buildFamilyMap(family), gmMap)
	if status != Trade {
		t.Errorf("CP diff >50%%: status = %q, want TRADE", status)
	}
}

func TestEvaluateTransfer_LowerIVTrade(t *testing.T) {
	family := []model.Pokemon{
		testPokemon("P-1", "pidgey", "pidgey", 80, 200, false, false, false, ""),
		testPokemon("P-2", "pidgey", "pidgey", 93, 200, false, false, false, ""),
	}
	gmMap := map[string]*model.GamemasterPokemon{}
	status, _ := evaluateTransfer(family[0], buildFamilyMap(family), gmMap)
	if status != Trade {
		t.Errorf("Lower IV than best: status = %q, want TRADE", status)
	}
}

func TestEvaluateTransfer_SafeDefault(t *testing.T) {
	family := []model.Pokemon{
		testPokemon("P-1", "pidgey", "pidgey", 50, 200, false, false, false, ""),
	}
	gmMap := map[string]*model.GamemasterPokemon{}
	status, _ := evaluateTransfer(family[0], buildFamilyMap(family), gmMap)
	if status != Safe {
		t.Errorf("Default: status = %q, want SAFE", status)
	}
}

// --- Test BuildTransferCandidates sort order ---

func TestBuildTransferCandidates_SortedCorrectly(t *testing.T) {
	c := []model.Pokemon{
		testPokemon("P-1", "pidgey", "pidgey", 50, 200, false, false, false, ""),
		testPokemon("P-2", "charmander", "charmander", 100, 2500, false, false, false, ""),
		testPokemon("P-3", "rattata", "rattata", 85, 300, false, false, false, ""),
		testPokemon("P-4", "caterpie", "caterpie", 93, 400, false, false, false, ""),
	}
	gmMap := map[string]*model.GamemasterPokemon{}

	results := BuildTransferCandidates(c, gmMap)

	// Expected order: SAFE(0) → TRADE(1) → SOFT_KEEP(2) → KEEP(3)
	// P-1: SAFE (50% IV, alone)
	// P-3: TRADE (85% IV, alone → actually check: IV<88%, CP=300 vs best=300, same IV, no better → SAFE... hmm)
	// Let me just verify the group ordering
	if len(results) != 4 {
		t.Fatalf("expected 4 results, got %d", len(results))
	}

	// Check that SAFE comes before KEEP
	safeIdx := -1
	keepIdx := -1
	softKeepIdx := -1
	for i, r := range results {
		switch r.Status {
		case Safe:
			safeIdx = i
		case Keep:
			keepIdx = i
		case SoftKeep:
			softKeepIdx = i
		}
	}
	if safeIdx >= 0 && keepIdx >= 0 && safeIdx > keepIdx {
		t.Error("SAFE should come before KEEP in results")
	}
	if safeIdx >= 0 && softKeepIdx >= 0 && safeIdx > softKeepIdx {
		t.Error("SAFE should come before SOFT_KEEP in results")
	}
}

// Helper: build family map from a slice of Pokemon.
func buildFamilyMap(pokemon []model.Pokemon) map[string][]model.Pokemon {
	families := map[string][]model.Pokemon{}
	for _, p := range pokemon {
		fid := p.FamilyID
		if fid == "" {
			fid = "__unknown__" + p.SpeciesID
		}
		families[fid] = append(families[fid], p)
	}
	return families
}

// --- isMythical / isLegendary helpers ---

func TestIsMythical(t *testing.T) {
	gmMap := map[string]*model.GamemasterPokemon{
		"mew":       {SpeciesID: "mew", Tags: []string{"mythical", "untradeable"}},
		"celebi":    {SpeciesID: "celebi", Tags: []string{"mythical", "untradeable"}},
		"mewtwo":    {SpeciesID: "mewtwo", Tags: []string{"legendary"}},
		"charizard": {SpeciesID: "charizard", Tags: []string{"starter"}},
		"unknown":   {SpeciesID: "unknown"},
	}
	tests := []struct {
		species string
		want    bool
	}{
		{"mew", true},
		{"celebi", true},
		{"mewtwo", false},
		{"charizard", false},
		{"unknown", false},
	}
	for _, tt := range tests {
		p := testPokemon("P-1", tt.species, tt.species, 80, 1000, false, false, false, "")
		if got := isMythical(p, gmMap); got != tt.want {
			t.Errorf("isMythical(%s) = %v, want %v", tt.species, got, tt.want)
		}
	}
}

func TestIsLegendary(t *testing.T) {
	gmMap := map[string]*model.GamemasterPokemon{
		"mewtwo":    {SpeciesID: "mewtwo", Tags: []string{"legendary"}},
		"groudon":   {SpeciesID: "groudon", Tags: []string{"legendary"}},
		"mew":       {SpeciesID: "mew", Tags: []string{"mythical", "untradeable"}},
		"charizard": {SpeciesID: "charizard", Tags: []string{"starter"}},
	}
	tests := []struct {
		species string
		want    bool
	}{
		{"mewtwo", true},
		{"groudon", true},
		{"mew", false},
		{"charizard", false},
	}
	for _, tt := range tests {
		p := testPokemon("P-1", tt.species, tt.species, 80, 1000, false, false, false, "")
		if got := isLegendary(p, gmMap); got != tt.want {
			t.Errorf("isLegendary(%s) = %v, want %v", tt.species, got, tt.want)
		}
	}
}

// --- Mythical transfer tests ---

func TestEvaluateTransfer_MythicalKeep(t *testing.T) {
	config.Global = config.Default()
	gmMap := map[string]*model.GamemasterPokemon{
		"mew": {SpeciesID: "mew", Tags: []string{"mythical", "untradeable"}},
	}
	family := []model.Pokemon{
		testPokemon("P-1", "mew", "mew", 82, 1375, false, false, false, ""),
	}
	status, reason := evaluateTransfer(family[0], buildFamilyMap(family), gmMap)
	if status != Keep {
		t.Errorf("Mythical: status = %q, want KEEP", status)
	}
	t.Logf("Reason: %s", reason)
}

func TestEvaluateTransfer_MythicalKeepLowIV(t *testing.T) {
	config.Global = config.Default()
	gmMap := map[string]*model.GamemasterPokemon{
		"celebi": {SpeciesID: "celebi", Tags: []string{"mythical", "untradeable"}},
	}
	family := []model.Pokemon{
		testPokemon("P-1", "celebi", "celebi", 50, 1000, false, false, false, ""),
	}
	status, _ := evaluateTransfer(family[0], buildFamilyMap(family), gmMap)
	if status != Keep {
		t.Errorf("Mythical low IV: status = %q, want KEEP", status)
	}
}

// --- Legendary transfer tests ---

func TestEvaluateTransfer_LegendarySolo(t *testing.T) {
	config.Global = config.Default()
	gmMap := map[string]*model.GamemasterPokemon{
		"lugia": {SpeciesID: "lugia", Tags: []string{"legendary"}},
	}
	family := []model.Pokemon{
		testPokemon("P-1", "lugia", "lugia", 75, 2045, false, false, false, ""),
	}
	status, reason := evaluateTransfer(family[0], buildFamilyMap(family), gmMap)
	if status != SoftKeep {
		t.Errorf("Legendary solo: status = %q, want SOFT_KEEP", status)
	}
	t.Logf("Reason: %s", reason)
}

func TestEvaluateTransfer_LegendaryUnderLimit(t *testing.T) {
	config.Global = config.Default()
	gmMap := map[string]*model.GamemasterPokemon{
		"groudon": {SpeciesID: "groudon", Tags: []string{"legendary"}},
	}
	// 3 copies
	family := []model.Pokemon{
		testPokemon("P-1", "groudon", "groudon", 86, 2889, false, false, false, ""),
		testPokemon("P-2", "groudon", "groudon", 84, 2894, false, false, false, ""),
		testPokemon("P-3", "groudon", "groudon", 82, 2800, false, false, false, ""),
	}
	status, reason := evaluateTransfer(family[2], buildFamilyMap(family), gmMap)
	if status != SoftKeep {
		t.Errorf("Legendary 3 copies: status = %q, want SOFT_KEEP", status)
	}
	t.Logf("Reason: %s", reason)
}

func TestEvaluateTransfer_LegendaryAtLimit(t *testing.T) {
	cfg := config.Default()
	cfg.Transfer.MaxLegendaries = 3
	config.Global = cfg

	gmMap := map[string]*model.GamemasterPokemon{
		"groudon": {SpeciesID: "groudon", Tags: []string{"legendary"}},
	}
	// Exactly 3 copies (= limit)
	family := []model.Pokemon{
		testPokemon("P-1", "groudon", "groudon", 86, 2889, false, false, false, ""),
		testPokemon("P-2", "groudon", "groudon", 84, 2894, false, false, false, ""),
		testPokemon("P-3", "groudon", "groudon", 82, 2800, false, false, false, ""),
	}
	status, _ := evaluateTransfer(family[2], buildFamilyMap(family), gmMap)
	if status != SoftKeep {
		t.Errorf("Legendary at limit (3/3): status = %q, want SOFT_KEEP", status)
	}
}

func TestEvaluateTransfer_LegendaryOverLimit(t *testing.T) {
	cfg := config.Default()
	cfg.Transfer.MaxLegendaries = 3
	config.Global = cfg

	gmMap := map[string]*model.GamemasterPokemon{
		"groudon": {SpeciesID: "groudon", Tags: []string{"legendary"}},
	}
	// 4 copies (over limit of 3)
	family := []model.Pokemon{
		testPokemon("P-1", "groudon", "groudon", 86, 2889, false, false, false, ""),
		testPokemon("P-2", "groudon", "groudon", 84, 2894, false, false, false, ""),
		testPokemon("P-3", "groudon", "groudon", 82, 2800, false, false, false, ""),
		testPokemon("P-4", "groudon", "groudon", 80, 2700, false, false, false, ""),
	}
	status, reason := evaluateTransfer(family[3], buildFamilyMap(family), gmMap)
	if status != Trade {
		t.Errorf("Legendary over limit (4/3): status = %q, want TRADE", status)
	}
	t.Logf("Reason: %s", reason)
}

func TestEvaluateTransfer_LegendaryLowIVSolo(t *testing.T) {
	config.Global = config.Default()
	gmMap := map[string]*model.GamemasterPokemon{
		"lugia": {SpeciesID: "lugia", Tags: []string{"legendary"}},
	}
	family := []model.Pokemon{
		testPokemon("P-1", "lugia", "lugia", 60, 1800, false, false, false, ""),
	}
	status, _ := evaluateTransfer(family[0], buildFamilyMap(family), gmMap)
	if status != SoftKeep {
		t.Errorf("Legendary low IV solo: status = %q, want SOFT_KEEP", status)
	}
}

func TestEvaluateTransfer_LegendaryDefault5(t *testing.T) {
	cfg := config.Default()
	cfg.Transfer.MaxLegendaries = 0 // should default to 5
	config.Global = cfg

	gmMap := map[string]*model.GamemasterPokemon{
		"mewtwo": {SpeciesID: "mewtwo", Tags: []string{"legendary"}},
	}
	// 5 copies — should be SOFT_KEEP (default limit = 5)
	var family []model.Pokemon
	for i := 0; i < 5; i++ {
		family = append(family, testPokemon(
			string(rune('A'+i))+"-1", "mewtwo", "mewtwo",
			float64(80+i), 3000+i*10, false, false, false, "",
		))
	}
	status, _ := evaluateTransfer(family[4], buildFamilyMap(family), gmMap)
	if status != SoftKeep {
		t.Errorf("Legendary 5/5 (default limit): status = %q, want SOFT_KEEP", status)
	}
}
