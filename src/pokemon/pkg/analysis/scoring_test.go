package analysis

import (
	"fmt"
	"math"
	"testing"

	"github.com/adzik/pokemon/config"
	"github.com/adzik/pokemon/model"
	"github.com/adzik/pokemon/pkg/gamemaster"
)

func TestCalculateCP(t *testing.T) {
	tests := []struct {
		name                     string
		baseAtk, baseDef, baseHP int
		ivAtk, ivDef, ivHP       int
		level                    float64
		wantCP                   int
	}{
		{"bulbasaur 15/15/15 lvl 50", 118, 111, 128, 15, 15, 15, 50, 1260},
		{"bulbasaur 10/10/10 lvl 30", 118, 111, 128, 10, 10, 10, 30, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cp := CalculateCP(tt.baseAtk, tt.baseDef, tt.baseHP, tt.ivAtk, tt.ivDef, tt.ivHP, tt.level)
			if tt.wantCP > 0 && cp != tt.wantCP {
				t.Errorf("CalculateCP() = %d, want %d", cp, tt.wantCP)
			}
			if cp <= 0 {
				t.Errorf("CalculateCP() = %d, want > 0", cp)
			}
		})
	}
}

func TestCalculateCPZeroIVs(t *testing.T) {
	cp := CalculateCP(100, 100, 100, 0, 0, 0, 50)
	if cp <= 0 {
		t.Errorf("CalculateCP(0 IVs) = %d, want > 0", cp)
	}
}

func TestCalculateCPMaxIVsHigher(t *testing.T) {
	cpLow := CalculateCP(100, 100, 100, 0, 0, 0, 50)
	cpHigh := CalculateCP(100, 100, 100, 15, 15, 15, 50)
	if cpHigh <= cpLow {
		t.Errorf("max IVs CP (%d) should be > min IVs CP (%d)", cpHigh, cpLow)
	}
}

func TestCalculateMaxCP(t *testing.T) {
	cp := CalculateMaxCP(118, 111, 128, 15, 15, 15)
	if cp <= 0 {
		t.Errorf("CalculateMaxCP() = %d, want > 0", cp)
	}
}

func TestIVPercentage(t *testing.T) {
	tests := []struct {
		atk, def, hp int
		wantPct      float64
	}{
		{15, 15, 15, 100.0},
		{0, 0, 0, 0.0},
		{10, 10, 10, 66.66666666666667},
		{5, 5, 5, 33.33333333333333},
	}
	for _, tt := range tests {
		name := fmt.Sprintf("%d/%d/%d", tt.atk, tt.def, tt.hp)
		t.Run(name, func(t *testing.T) {
			pct := float64(tt.atk+tt.def+tt.hp) / 45.0 * 100.0
			if math.Abs(pct-tt.wantPct) > 0.001 {
				t.Errorf("IV%% = %f, want %f", pct, tt.wantPct)
			}
		})
	}
}

func TestGetCPM(t *testing.T) {
	cpm1 := GetCPM(1)
	if cpm1 < 0.093 || cpm1 > 0.095 {
		t.Errorf("GetCPM(1) = %f, want ~0.094", cpm1)
	}
	cpm25 := GetCPM(25)
	if cpm25 < 0.667 || cpm25 > 0.669 {
		t.Errorf("GetCPM(25) = %f, want ~0.667934", cpm25)
	}
	cpm50 := GetCPM(50)
	if cpm50 < 0.840 || cpm50 > 0.841 {
		t.Errorf("GetCPM(50) = %f, want ~0.8403", cpm50)
	}
}

func TestGetCPMNonDecreasing(t *testing.T) {
	var prev float64
	for level := 1.0; level <= 50; level += 0.5 {
		cpm := GetCPM(level)
		if cpm < prev {
			t.Errorf("GetCPM(%f) = %f < previous %f", level, cpm, prev)
		}
		prev = cpm
	}
}

func TestCalculateContextScore_MegaPts(t *testing.T) {
	config.Global = config.Default()
	gmMap := map[string]*model.GamemasterPokemon{}

	// Mega Charizard Y — should get MegaPts
	mega := model.Pokemon{
		PokemonRaw: model.PokemonRaw{Species: "Charizard"},
		SpeciesID:  "charizard_mega_y",
		IVPct:      100,
	}
	result := CalculateContextScore(mega, "S", gmMap)
	if result.MegaPts != 5 {
		t.Errorf("MegaPts = %f, want 5", result.MegaPts)
	}

	// Regular Charizard — no MegaPts
	regular := model.Pokemon{
		PokemonRaw: model.PokemonRaw{Species: "Charizard"},
		SpeciesID:  "charizard",
		IVPct:      100,
	}
	result2 := CalculateContextScore(regular, "S", gmMap)
	if result2.MegaPts != 0 {
		t.Errorf("Regular MegaPts = %f, want 0", result2.MegaPts)
	}
}

func TestCalculateContextScore_TransferKolekcjaZero(t *testing.T) {
	config.Global = config.Default()
	gmMap := map[string]*model.GamemasterPokemon{}

	p := model.Pokemon{
		PokemonRaw: model.PokemonRaw{Purpose: "Transfer"},
		SpeciesID:  "bulbasaur",
		IVPct:      100,
		RaidTier:   "S",
	}
	result := CalculateContextScore(p, "S", gmMap)
	if result.Score != 0 {
		t.Errorf("Transfer score = %f, want 0", result.Score)
	}

	p.Purpose = "Kolekcja"
	result2 := CalculateContextScore(p, "S", gmMap)
	if result2.Score != 0 {
		t.Errorf("Kolekcja score = %f, want 0", result2.Score)
	}
}

func TestCalculateContextScore_ScoreComponents(t *testing.T) {
	config.Global = config.Default()
	gmMap := map[string]*model.GamemasterPokemon{
		"bulbasaur": {
			SpeciesID:    "bulbasaur",
			FastMoves:    []string{"VINE_WHIP"},
			ChargedMoves: []string{"SLUDGE_BOMB", "SEED_BOMB"},
		},
	}

	p := model.Pokemon{
		PokemonRaw: model.PokemonRaw{
			Species:   "Bulbasaur",
			CP:        1500,
			FastMove:  "Vine Whip",
			Charged1:  "Sludge Bomb",
			Charged2:  "Seed Bomb",
		},
		SpeciesID: "bulbasaur",
		IVPct:     100,
		IsLucky:   true,
		BaseStats: model.BaseStats{Atk: 118, Def: 111, HP: 128},
	}

	result := CalculateContextScore(p, "S", gmMap)

	if result.IVScore != 15 {
		t.Errorf("IVScore = %f, want 15", result.IVScore)
	}
	if result.LuckyPts != 20 {
		t.Errorf("LuckyPts = %f, want 20", result.LuckyPts)
	}
	if result.MetaPts != 20 {
		t.Errorf("MetaPts = %f, want 20", result.MetaPts)
	}
	if result.MovePts != 15 {
		t.Errorf("MovePts = %f, want 15", result.MovePts)
	}
	if result.CPPts <= 0 {
		t.Errorf("CPPts = %f, want > 0", result.CPPts)
	}
	if result.Score <= 0 {
		t.Errorf("Score = %f, want > 0", result.Score)
	}
}

func TestCalculateContextScore_MoveSuggestions(t *testing.T) {
	config.Global = config.Default()
	gmMap := map[string]*model.GamemasterPokemon{
		"bulbasaur": {
			SpeciesID:    "bulbasaur",
			FastMoves:    []string{"VINE_WHIP"},
			ChargedMoves: []string{"SLUDGE_BOMB", "SEED_BOMB"},
		},
	}

	// Unknown fast move → note to change
	p := model.Pokemon{
		PokemonRaw: model.PokemonRaw{
			Species:  "Bulbasaur",
			FastMove: "TACKLE",
			Charged1: "SLUDGE_BOMB",
			Charged2: "SEED_BOMB",
		},
		SpeciesID: "bulbasaur",
		IVPct:     100,
	}

	result := CalculateContextScore(p, "S", gmMap)
	if len(result.Notes) == 0 {
		t.Error("expected note about changing fast move")
	}
	if result.MovePts != 10 {
		t.Errorf("MovePts = %f, want 10 (2 charged matches, fast wrong)", result.MovePts)
	}
}

func TestGetCPM_Interpolation(t *testing.T) {
	cpm1 := GetCPM(1.0)
	cpm15 := GetCPM(1.5)
	cpm2 := GetCPM(2.0)

	if cpm1 >= cpm15 {
		t.Errorf("GetCPM(1.0)=%f should be < GetCPM(1.5)=%f", cpm1, cpm15)
	}
	if cpm15 >= cpm2 {
		t.Errorf("GetCPM(1.5)=%f should be < GetCPM(2.0)=%f", cpm15, cpm2)
	}
	// Interpolation: GetCPM(1.25) should be between GetCPM(1.0) and GetCPM(1.5)
	cpm125 := GetCPM(1.25)
	if cpm125 <= cpm1 || cpm125 >= cpm15 {
		t.Errorf("GetCPM(1.25)=%f not between %f and %f", cpm125, cpm1, cpm15)
	}
}

func TestGetCPM_EdgeCases(t *testing.T) {
	// Level below 1 → clamp to index 0
	cpm05 := GetCPM(0.5)
	if cpm05 < 0.09 {
		t.Errorf("GetCPM(0.5) = %f, should clamp to >= L1 value", cpm05)
	}
	// Level above 50 → clamp to last value in table
	cpm55 := GetCPM(55)
	cpmLast := GetCPM(99) // beyond table, same clamp
	if cpm55 != cpmLast {
		t.Errorf("GetCPM(55) = %f should equal last table value %f", cpm55, cpmLast)
	}
}

func TestResolveTierForPurpose(t *testing.T) {
	metaCats := gamemaster.MetaTiersByCategory{
		Raid: map[string]string{"pidgey": "C"},
		GL:   map[string]string{"pidgey": "B"},
		UL:   map[string]string{"pidgey": "A"},
		ML:   map[string]string{"pidgey": "S"},
	}

	tests := []struct {
		purpose  string
		cP       int
		raidT    string
		glT      string
		ulT      string
		mlT      string
		wantTier string
	}{
		{"rajdy", 0, "S", "", "", "", "S"},
		{"raid", 0, "A", "", "", "", "A"},
		{"gl", 0, "", "B", "", "", "B"},
		{"ul", 0, "", "", "A", "", "A"},
		{"ml", 0, "", "", "", "S", "S"},
		{"pvp", 1400, "", "A", "", "", "A"},     // CP≤1500 → GL
		{"pvp", 2000, "", "", "B", "", "B"},     // CP≤2500 → UL
		{"pvp", 3000, "", "", "", "S", "S"},     // CP>2500 → ML
		{"gym", 0, "S", "", "", "", "S"},
		{"kolekcja", 0, "", "", "", "", "S"},    // default → BestMetaTier → S (from metaCats ML)
	}

	for _, tt := range tests {
		t.Run(tt.purpose, func(t *testing.T) {
			p := model.Pokemon{
				PokemonRaw: model.PokemonRaw{
					Purpose: tt.purpose,
					CP:      tt.cP,
				},
				SpeciesID: "pidgey",
				RaidTier:  tt.raidT,
				GLTier:    tt.glT,
				ULTier:    tt.ulT,
				MLTier:    tt.mlT,
			}
			got := resolveTierForPurpose(p, metaCats)
			if got != tt.wantTier {
				t.Errorf("resolveTierForPurpose(%q) = %q, want %q", tt.purpose, got, tt.wantTier)
			}
		})
	}
}

func TestCalculateDevScore(t *testing.T) {
	config.Global = config.Default()
	gmMap := map[string]*model.GamemasterPokemon{}
	metaCats := gamemaster.MetaTiersByCategory{
		Raid: map[string]string{"bulbasaur": "S"},
	}

	// Transfer → 0 with note
	p := model.Pokemon{
		PokemonRaw: model.PokemonRaw{Purpose: "Transfer"},
		SpeciesID:  "bulbasaur",
		IVPct:      100,
	}
	result := CalculateDevScore(p, gmMap, metaCats)
	if result.Score != 0 {
		t.Errorf("Transfer DevScore = %f, want 0", result.Score)
	}
	if len(result.Notes) == 0 {
		t.Error("expected note for Transfer")
	}

	// Raid → uses RaidTier
	p2 := model.Pokemon{
		PokemonRaw: model.PokemonRaw{Purpose: "rajdy"},
		SpeciesID:  "bulbasaur",
		IVPct:      100,
		RaidTier:   "S",
	}
	result2 := CalculateDevScore(p2, gmMap, metaCats)
	if result2.MetaPts != 20 {
		t.Errorf("Raid DevScore MetaPts = %f, want 20", result2.MetaPts)
	}
}

func TestIsMoveAvailable(t *testing.T) {
	available := []string{"VINE_WHIP", "TACKLE", "RAZOR_LEAF"}

	tests := []struct {
		move string
		want bool
	}{
		{"VINE_WHIP", true},
		{"vine_whip", true},
		{"Vine Whip", true},
		{"TACKLE", true},
		{"BODY_SLAM", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.move, func(t *testing.T) {
			got := isMoveAvailable(tt.move, available)
			if got != tt.want {
				t.Errorf("isMoveAvailable(%q) = %v, want %v", tt.move, got, tt.want)
			}
		})
	}
}

func TestSortByScore(t *testing.T) {
	c := []model.Pokemon{
		{PokemonRaw: model.PokemonRaw{UID: "P-1", CP: 2000}, DevScore: 50, IVPct: 90},
		{PokemonRaw: model.PokemonRaw{UID: "P-2", CP: 2500}, DevScore: 80, IVPct: 95},
		{PokemonRaw: model.PokemonRaw{UID: "P-3", CP: 2000}, DevScore: 80, IVPct: 95},
		{PokemonRaw: model.PokemonRaw{UID: "P-4", CP: 2500}, DevScore: 80, IVPct: 90},
		{PokemonRaw: model.PokemonRaw{UID: "P-5", CP: 3000}, DevScore: 30, IVPct: 100},
	}

	sorted := SortByScore(c)

	// P-2 (80/95/2500) > P-3 (80/95/2000) > P-4 (80/90/2500) > P-1 (50/90/2000) > P-5 (30/100/3000)
	if sorted[0].UID != "P-2" {
		t.Errorf("sorted[0] = %s, want P-2", sorted[0].UID)
	}
	if sorted[1].UID != "P-3" {
		t.Errorf("sorted[1] = %s, want P-3", sorted[1].UID)
	}
	if sorted[2].UID != "P-4" {
		t.Errorf("sorted[2] = %s, want P-4", sorted[2].UID)
	}
	if sorted[3].UID != "P-1" {
		t.Errorf("sorted[3] = %s, want P-1", sorted[3].UID)
	}
	if sorted[4].UID != "P-5" {
		t.Errorf("sorted[4] = %s, want P-5", sorted[4].UID)
	}
}

func TestSortByScore_DoesNotMutateOriginal(t *testing.T) {
	c := []model.Pokemon{
		{PokemonRaw: model.PokemonRaw{UID: "P-1"}, DevScore: 50},
		{PokemonRaw: model.PokemonRaw{UID: "P-2"}, DevScore: 80},
	}
	origUID := c[0].UID
	origScore := c[0].DevScore

	SortByScore(c)

	if c[0].UID != origUID || c[0].DevScore != origScore {
		t.Error("SortByScore mutated original")
	}
}

func TestEnrichWithScores(t *testing.T) {
	config.Global = config.Default()
	gmMap := map[string]*model.GamemasterPokemon{
		"bulbasaur": {
			SpeciesID:    "bulbasaur",
			BaseStats:    model.BaseStats{Atk: 118, Def: 111, HP: 128},
			FastMoves:    []string{"VINE_WHIP"},
			ChargedMoves: []string{"SLUDGE_BOMB", "SEED_BOMB"},
			Family:       &model.Family{ID: "bulbasaur"},
		},
	}
	moveMap := map[string]*gamemaster.MoveEntry{
		"VINE_WHIP": {MoveID: "VINE_WHIP", Type: "grass", Power: 7, Cooldown: 600, EnergyGain: 12},
		"SLUDGE_BOMB": {MoveID: "SLUDGE_BOMB", Type: "poison", Power: 80, Cooldown: 2300, Energy: 50},
		"SEED_BOMB": {MoveID: "SEED_BOMB", Type: "grass", Power: 65, Cooldown: 2100, Energy: 45},
	}
	metaCats := gamemaster.MetaTiersByCategory{
		Raid: map[string]string{"bulbasaur": "S"},
	}

	c := []model.Pokemon{
		{
			PokemonRaw: model.PokemonRaw{
				UID: "P-1", Species: "Bulbasaur",
				CP: 1500, FastMove: "Vine Whip", Charged1: "Sludge Bomb", Charged2: "Seed Bomb",
			},
			SpeciesID: "bulbasaur",
			FamilyID:  "bulbasaur",
			IVPct:     100,
			BaseStats: model.BaseStats{Atk: 118, Def: 111, HP: 128},
			RaidTier:  "S",
		},
	}

	result := EnrichWithScores(c, gmMap, moveMap, metaCats)

	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	p := result[0]
	if p.MaxCP <= 0 {
		t.Errorf("MaxCP = %d, want > 0", p.MaxCP)
	}
	if p.EffectiveATK <= 0 {
		t.Errorf("EffectiveATK = %f, want > 0", p.EffectiveATK)
	}
	if p.RaidScore <= 0 {
		t.Errorf("RaidScore = %f, want > 0", p.RaidScore)
	}
	if p.DevScore <= 0 {
		t.Errorf("DevScore = %f, want > 0", p.DevScore)
	}
	if p.MoveDPSFast <= 0 {
		t.Errorf("MoveDPSFast = %f, want > 0", p.MoveDPSFast)
	}
	if p.MoveDPSCharged <= 0 {
		t.Errorf("MoveDPSCharged = %f, want > 0", p.MoveDPSCharged)
	}
	if p.RaidDPS <= 0 {
		t.Errorf("RaidDPS = %f, want > 0", p.RaidDPS)
	}
	if p.RaidDPSOptimal <= 0 {
		t.Errorf("RaidDPSOptimal = %f, want > 0", p.RaidDPSOptimal)
	}
}
