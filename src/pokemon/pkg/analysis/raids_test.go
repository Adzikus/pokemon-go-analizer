package analysis

import (
	"testing"

	"github.com/adzik/pokemon/config"
	"github.com/adzik/pokemon/model"
	"github.com/adzik/pokemon/pkg/gamemaster"
)

func TestCounterType(t *testing.T) {
	tests := []struct {
		bossType string
		want     []string
	}{
		{"fire", []string{"water", "ground", "rock"}},
		{"water", []string{"electric", "grass"}},
		{"grass", []string{"fire", "ice", "poison", "flying", "bug"}},
		{"dragon", []string{"ice", "dragon", "fairy"}},
		{"normal", []string{"fighting"}},
		{"unknown", nil},
	}

	for _, tt := range tests {
		t.Run(tt.bossType, func(t *testing.T) {
			got := CounterType(tt.bossType)
			if len(got) != len(tt.want) {
				t.Errorf("CounterType(%q) = %v, want %v", tt.bossType, got, tt.want)
				return
			}
			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("CounterType(%q)[%d] = %q, want %q", tt.bossType, i, v, tt.want[i])
				}
			}
		})
	}
}

func TestIsSuperEffective(t *testing.T) {
	tests := []struct {
		moveType string
		bossType string
		want     bool
	}{
		{"water", "fire", true},
		{"fire", "grass", true},
		{"ice", "dragon", true},
		{"fairy", "dragon", true},
		{"normal", "ghost", false},
		{"fire", "water", false},
		{"water", "water", false},
	}

	for _, tt := range tests {
		t.Run(tt.moveType+"_vs_"+tt.bossType, func(t *testing.T) {
			got := IsSuperEffective(tt.moveType, tt.bossType)
			if got != tt.want {
				t.Errorf("IsSuperEffective(%q, %q) = %v, want %v", tt.moveType, tt.bossType, got, tt.want)
			}
		})
	}
}

func TestIsSTAB(t *testing.T) {
	tests := []struct {
		moveType string
		types    []string
		want     bool
	}{
		{"fire", []string{"fire", "flying"}, true},
		{"flying", []string{"fire", "flying"}, true},
		{"water", []string{"fire", "flying"}, false},
		{"fire", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.moveType, func(t *testing.T) {
			got := IsSTAB(tt.moveType, tt.types)
			if got != tt.want {
				t.Errorf("IsSTAB(%q, %v) = %v, want %v", tt.moveType, tt.types, got, tt.want)
			}
		})
	}
}

func TestEffectiveness(t *testing.T) {
	config.Global = config.Default()

	tests := []struct {
		moveType string
		bossType string
		want     float64
	}{
		{"water", "fire", 1.6},
		{"fire", "grass", 1.6},
		{"fire", "water", 1.0},
		{"water", "", 1.0},
		{"normal", "fire", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.moveType+"_vs_"+tt.bossType, func(t *testing.T) {
			got := Effectiveness(tt.moveType, tt.bossType)
			if got != tt.want {
				t.Errorf("Effectiveness(%q, %q) = %f, want %f", tt.moveType, tt.bossType, got, tt.want)
			}
		})
	}
}

func TestCalculateMoveDPS_Basic(t *testing.T) {
	config.Global = config.Default()

	dps := CalculateMoveDPS(
		150, 15,
		10, 500, 10,
		100, 2000, 50,
		"water", "fire",
		[]string{"water"}, "fire",
		false,
	)

	if dps <= 0 {
		t.Errorf("DPS = %f, want > 0", dps)
	}
}

func TestCalculateMoveDPS_ShadowBonus(t *testing.T) {
	config.Global = config.Default()

	dpsNormal := CalculateMoveDPS(150, 15, 10, 500, 10, 100, 2000, 50, "water", "fire", []string{"water"}, "fire", false)
	dpsShadow := CalculateMoveDPS(150, 15, 10, 500, 10, 100, 2000, 50, "water", "fire", []string{"water"}, "fire", true)

	if dpsShadow <= dpsNormal {
		t.Errorf("Shadow DPS (%f) should be > normal DPS (%f)", dpsShadow, dpsNormal)
	}
}

func TestCalculateMoveDPS_STABBonus(t *testing.T) {
	config.Global = config.Default()

	dpsSTAB := CalculateMoveDPS(150, 15, 10, 500, 10, 100, 2000, 50, "water", "fire", []string{"water"}, "fire", false)
	dpsNoSTAB := CalculateMoveDPS(150, 15, 10, 500, 10, 100, 2000, 50, "fire", "grass", []string{"water"}, "fire", false)

	if dpsSTAB <= dpsNoSTAB {
		t.Errorf("STAB DPS (%f) should be > non-STAB DPS (%f)", dpsSTAB, dpsNoSTAB)
	}
}

func TestCalculateMoveDPS_SuperEffectiveBonus(t *testing.T) {
	config.Global = config.Default()

	dpsSE := CalculateMoveDPS(150, 15, 10, 500, 10, 100, 2000, 50, "water", "fire", []string{"water"}, "fire", false)
	dpsNeutral := CalculateMoveDPS(150, 15, 10, 500, 10, 100, 2000, 50, "water", "fire", []string{"water"}, "normal", false)

	if dpsSE <= dpsNeutral {
		t.Errorf("SE DPS (%f) should be > neutral DPS (%f)", dpsSE, dpsNeutral)
	}
}

func TestCalculateMoveDPS_ZeroEnergy(t *testing.T) {
	config.Global = config.Default()

	dps := CalculateMoveDPS(150, 15, 10, 500, 10, 100, 2000, 0, "water", "fire", []string{"water"}, "fire", false)

	if dps <= 0 {
		t.Errorf("DPS with zero fast energy = %f, want > 0 (fast-only)", dps)
	}
}

func TestCalculateMoveDPSRaw(t *testing.T) {
	fastDPS, chargedDPS := CalculateMoveDPSRaw(10, 500, 10, 100, 2000, 50)

	if fastDPS != 20.0 {
		t.Errorf("fastDPS = %f, want 20.0", fastDPS)
	}
	if chargedDPS != 40.0 {
		t.Errorf("chargedDPS = %f, want 40.0", chargedDPS)
	}
}

func TestCalculateMoveDPSRaw_ZeroEnergy(t *testing.T) {
	fastDPS, chargedDPS := CalculateMoveDPSRaw(10, 500, 10, 100, 2000, 0)

	if fastDPS != 20.0 {
		t.Errorf("fastDPS = %f, want 20.0", fastDPS)
	}
	if chargedDPS != 0 {
		t.Errorf("chargedDPS = %f, want 0 (zero energy)", chargedDPS)
	}
}

func TestNormalizeMoveID(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Vine Whip", "VINE_WHIP"},
		{"  Tackle  ", "TACKLE"},
		{"vine_whip", "VINE_WHIP"},
		{"SLUDGE_BOMB", "SLUDGE_BOMB"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeMoveID(tt.input)
			if got != tt.want {
				t.Errorf("normalizeMoveID(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCalculatePokemonDPS(t *testing.T) {
	config.Global = config.Default()
	gmMap := map[string]*model.GamemasterPokemon{
		"bulbasaur": {SpeciesID: "bulbasaur"},
	}
	moveMap := map[string]*gamemaster.MoveEntry{
		"VINE_WHIP":   {MoveID: "VINE_WHIP", Name: "Vine Whip", Type: "grass", Power: 7, Cooldown: 600, EnergyGain: 12},
		"SLUDGE_BOMB": {MoveID: "SLUDGE_BOMB", Name: "Sludge Bomb", Type: "poison", Power: 80, Cooldown: 2300, Energy: 50},
	}

	p := model.Pokemon{
		PokemonRaw: model.PokemonRaw{
			FastMove: "Vine Whip",
			Charged1: "Sludge Bomb",
			Type1:     "grass",
			Type2:     "poison",
			IVAtk:     15,
		},
		SpeciesID: "bulbasaur",
		BaseStats: model.BaseStats{Atk: 118, Def: 111, HP: 128},
	}

	dps, fast, charged := CalculatePokemonDPS(p, gmMap, moveMap, "fire")
	if dps <= 0 {
		t.Errorf("DPS = %f, want > 0", dps)
	}
	if fast != "Vine Whip" {
		t.Errorf("fast = %q, want 'Vine Whip'", fast)
	}
	if charged != "Sludge Bomb" {
		t.Errorf("charged = %q, want 'Sludge Bomb'", charged)
	}
}

func TestCalculatePokemonDPS_UnknownSpecies(t *testing.T) {
	config.Global = config.Default()
	gmMap := map[string]*model.GamemasterPokemon{}
	moveMap := map[string]*gamemaster.MoveEntry{}

	p := model.Pokemon{
		PokemonRaw: model.PokemonRaw{FastMove: "Tackle", Charged1: "Body Slam"},
		SpeciesID:  "unknown",
	}

	dps, fast, charged := CalculatePokemonDPS(p, gmMap, moveMap, "")
	if dps != 0 {
		t.Errorf("DPS = %f, want 0 for unknown species", dps)
	}
	if fast != "" || charged != "" {
		t.Errorf("expected empty moves for unknown species")
	}
}

func TestCalculatePokemonDPS_MissingMove(t *testing.T) {
	config.Global = config.Default()
	gmMap := map[string]*model.GamemasterPokemon{
		"bulbasaur": {SpeciesID: "bulbasaur"},
	}
	moveMap := map[string]*gamemaster.MoveEntry{
		"VINE_WHIP": {MoveID: "VINE_WHIP", Name: "Vine Whip", Type: "grass", Power: 7, Cooldown: 600, EnergyGain: 12},
	}

	p := model.Pokemon{
		PokemonRaw: model.PokemonRaw{
			FastMove: "Vine Whip",
			Charged1: "Sludge Bomb",
			Type1:     "grass",
			IVAtk:     15,
		},
		SpeciesID: "bulbasaur",
		BaseStats: model.BaseStats{Atk: 118},
	}

	dps, fast, charged := CalculatePokemonDPS(p, gmMap, moveMap, "fire")
	if dps != 0 {
		t.Errorf("DPS = %f, want 0 for missing move", dps)
	}
	if fast != "Vine Whip" || charged != "Sludge Bomb" {
		t.Errorf("expected original move names returned")
	}
}

func TestCalculateEffectiveATK(t *testing.T) {
	config.Global = config.Default()

	p := model.Pokemon{
		PokemonRaw: model.PokemonRaw{IVAtk: 15},
		BaseStats:  model.BaseStats{Atk: 150},
	}
	atk := CalculateEffectiveATK(p)
	if atk <= 0 {
		t.Errorf("ATK = %f, want > 0", atk)
	}

	pShadow := model.Pokemon{
		PokemonRaw: model.PokemonRaw{IVAtk: 15},
		BaseStats:  model.BaseStats{Atk: 150},
		IsShadow:   true,
	}
	atkShadow := CalculateEffectiveATK(pShadow)
	if atkShadow <= atk {
		t.Errorf("Shadow ATK (%f) should be > normal ATK (%f)", atkShadow, atk)
	}
}

func TestFindOptimalMoves(t *testing.T) {
	config.Global = config.Default()
	gmMap := map[string]*model.GamemasterPokemon{
		"bulbasaur": {
			SpeciesID:    "bulbasaur",
			FastMoves:    []string{"VINE_WHIP", "TACKLE"},
			ChargedMoves: []string{"SLUDGE_BOMB", "SEED_BOMB"},
		},
	}
	moveMap := map[string]*gamemaster.MoveEntry{
		"VINE_WHIP":   {MoveID: "VINE_WHIP", Name: "Vine Whip", Type: "grass", Power: 7, Cooldown: 600, EnergyGain: 12},
		"TACKLE":      {MoveID: "TACKLE", Name: "Tackle", Type: "normal", Power: 5, Cooldown: 500, EnergyGain: 5},
		"SLUDGE_BOMB": {MoveID: "SLUDGE_BOMB", Name: "Sludge Bomb", Type: "poison", Power: 80, Cooldown: 2300, Energy: 50},
		"SEED_BOMB":   {MoveID: "SEED_BOMB", Name: "Seed Bomb", Type: "grass", Power: 65, Cooldown: 2100, Energy: 45},
	}

	p := model.Pokemon{
		SpeciesID: "bulbasaur",
		PokemonRaw: model.PokemonRaw{
			Type1: "grass",
			IVAtk: 15,
		},
		BaseStats: model.BaseStats{Atk: 118},
	}

	fast, charged, dps := FindOptimalMoves(p, gmMap, moveMap, "fire")
	if dps <= 0 {
		t.Errorf("optimal DPS = %f, want > 0", dps)
	}
	if fast == "" || charged == "" {
		t.Errorf("optimal moves: fast=%q, charged=%q, want non-empty", fast, charged)
	}
}

func TestFindOptimalMoves_UnknownSpecies(t *testing.T) {
	config.Global = config.Default()
	gmMap := map[string]*model.GamemasterPokemon{}
	moveMap := map[string]*gamemaster.MoveEntry{}

	p := model.Pokemon{SpeciesID: "unknown"}
	fast, charged, dps := FindOptimalMoves(p, gmMap, moveMap, "")
	if fast != "" || charged != "" || dps != 0 {
		t.Errorf("expected empty results for unknown species")
	}
}

func TestFindBestCounters(t *testing.T) {
	config.Global = config.Default()
	gmMap := map[string]*model.GamemasterPokemon{
		"blastoise": {SpeciesID: "blastoise"},
		"venusaur":  {SpeciesID: "venusaur"},
		"charizard": {SpeciesID: "charizard"},
	}
	moveMap := map[string]*gamemaster.MoveEntry{
		"WATER_GUN":    {MoveID: "WATER_GUN", Name: "Water Gun", Type: "water", Power: 5, Cooldown: 500, EnergyGain: 5},
		"HYDRO_PUMP":   {MoveID: "HYDRO_PUMP", Name: "Hydro Pump", Type: "water", Power: 130, Cooldown: 3300, Energy: 100},
		"VINE_WHIP":    {MoveID: "VINE_WHIP", Name: "Vine Whip", Type: "grass", Power: 7, Cooldown: 600, EnergyGain: 12},
		"SOLAR_BEAM":   {MoveID: "SOLAR_BEAM", Name: "Solar Beam", Type: "grass", Power: 180, Cooldown: 4900, Energy: 100},
		"EMBER":        {MoveID: "EMBER", Name: "Ember", Type: "fire", Power: 10, Cooldown: 1000, EnergyGain: 10},
		"FLAMETHROWER": {MoveID: "FLAMETHROWER", Name: "Flamethrower", Type: "fire", Power: 70, Cooldown: 2200, Energy: 50},
	}

	c := []model.Pokemon{
		{PokemonRaw: model.PokemonRaw{FastMove: "Water Gun", Charged1: "Hydro Pump", CP: 1500, Type1: "water", IVAtk: 15}, SpeciesID: "blastoise", BaseStats: model.BaseStats{Atk: 126}},
		{PokemonRaw: model.PokemonRaw{FastMove: "Vine Whip", Charged1: "Solar Beam", CP: 1500, Type1: "grass", IVAtk: 15}, SpeciesID: "venusaur", BaseStats: model.BaseStats{Atk: 136}},
		{PokemonRaw: model.PokemonRaw{FastMove: "Ember", Charged1: "Flamethrower", CP: 1500, Type1: "fire", IVAtk: 15}, SpeciesID: "charizard", BaseStats: model.BaseStats{Atk: 169}},
	}

	counters := FindBestCounters(c, gmMap, moveMap, "fire", 2)
	if len(counters) != 2 {
		t.Fatalf("expected 2 counters, got %d", len(counters))
	}
	if counters[0].DPS < counters[1].DPS {
		t.Error("counters should be sorted by DPS descending")
	}
}

func TestFindBestCounters_MinCPFilter(t *testing.T) {
	config.Global = config.Default()
	gmMap := map[string]*model.GamemasterPokemon{
		"a": {SpeciesID: "a"},
	}
	moveMap := map[string]*gamemaster.MoveEntry{
		"TACKLE":    {MoveID: "TACKLE", Name: "Tackle", Type: "normal", Power: 5, Cooldown: 500, EnergyGain: 5},
		"BODY_SLAM": {MoveID: "BODY_SLAM", Name: "Body Slam", Type: "normal", Power: 50, Cooldown: 2100, Energy: 35},
	}

	c := []model.Pokemon{
		{PokemonRaw: model.PokemonRaw{FastMove: "Tackle", Charged1: "Body Slam", CP: 500, Type1: "normal", IVAtk: 10}, SpeciesID: "a", BaseStats: model.BaseStats{Atk: 100}},
		{PokemonRaw: model.PokemonRaw{FastMove: "Tackle", Charged1: "Body Slam", CP: 2000, Type1: "normal", IVAtk: 10}, SpeciesID: "a", BaseStats: model.BaseStats{Atk: 100}},
	}

	counters := FindBestCounters(c, gmMap, moveMap, "", 10)
	if len(counters) != 1 {
		t.Fatalf("expected 1 counter (MinCP=1000), got %d", len(counters))
	}
	if counters[0].Pokemon.CP != 2000 {
		t.Errorf("expected CP=2000, got %d", counters[0].Pokemon.CP)
	}
}

func TestFindBestCounters_SuperEffFlag(t *testing.T) {
	config.Global = config.Default()
	gmMap := map[string]*model.GamemasterPokemon{
		"a": {SpeciesID: "a"},
	}
	moveMap := map[string]*gamemaster.MoveEntry{
		"WATER_GUN":  {MoveID: "WATER_GUN", Name: "Water Gun", Type: "water", Power: 5, Cooldown: 500, EnergyGain: 5},
		"HYDRO_PUMP": {MoveID: "HYDRO_PUMP", Name: "Hydro Pump", Type: "water", Power: 130, Cooldown: 3300, Energy: 100},
	}

	c := []model.Pokemon{
		{PokemonRaw: model.PokemonRaw{FastMove: "Water Gun", Charged1: "Hydro Pump", CP: 2000, Type1: "water", IVAtk: 15}, SpeciesID: "a", BaseStats: model.BaseStats{Atk: 150}},
	}

	counters := FindBestCounters(c, gmMap, moveMap, "fire", 10)
	if len(counters) != 1 {
		t.Fatalf("expected 1 counter, got %d", len(counters))
	}
	if !counters[0].SuperEff {
		t.Error("expected SuperEff=true for water vs fire boss")
	}
}

func TestBuildRaidRankings(t *testing.T) {
	config.Global = config.Default()
	gmMap := map[string]*model.GamemasterPokemon{
		"a": {SpeciesID: "a"},
	}
	moveMap := map[string]*gamemaster.MoveEntry{
		"TACKLE":    {MoveID: "TACKLE", Name: "Tackle", Type: "normal", Power: 5, Cooldown: 500, EnergyGain: 5},
		"BODY_SLAM": {MoveID: "BODY_SLAM", Name: "Body Slam", Type: "normal", Power: 50, Cooldown: 2100, Energy: 35},
	}

	c := []model.Pokemon{
		{PokemonRaw: model.PokemonRaw{FastMove: "Tackle", Charged1: "Body Slam", CP: 2000, Type1: "normal", IVAtk: 15}, SpeciesID: "a", BaseStats: model.BaseStats{Atk: 150}},
	}

	rankings := BuildRaidRankings(c, gmMap, moveMap, 10)

	if len(rankings.ByType) != 18 {
		t.Errorf("ByType has %d entries, want 18", len(rankings.ByType))
	}
	if len(rankings.Universal) != 1 {
		t.Errorf("Universal has %d entries, want 1", len(rankings.Universal))
	}
}

func TestBuildRaidRankings_Empty(t *testing.T) {
	config.Global = config.Default()
	gmMap := map[string]*model.GamemasterPokemon{}
	moveMap := map[string]*gamemaster.MoveEntry{}

	rankings := BuildRaidRankings(nil, gmMap, moveMap, 10)

	if len(rankings.ByType) != 18 {
		t.Errorf("ByType should have 18 entries even with empty collection")
	}
	for _, list := range rankings.ByType {
		if len(list) != 0 {
			t.Errorf("expected empty list for empty collection, got %d", len(list))
		}
	}
}

func TestRaidDPSResult_Fields(t *testing.T) {
	config.Global = config.Default()
	gmMap := map[string]*model.GamemasterPokemon{
		"blastoise": {
			SpeciesID:    "blastoise",
			FastMoves:    []string{"WATER_GUN"},
			ChargedMoves: []string{"HYDRO_PUMP"},
		},
	}
	moveMap := map[string]*gamemaster.MoveEntry{
		"WATER_GUN":  {MoveID: "WATER_GUN", Name: "Water Gun", Type: "water", Power: 5, Cooldown: 500, EnergyGain: 5},
		"HYDRO_PUMP": {MoveID: "HYDRO_PUMP", Name: "Hydro Pump", Type: "water", Power: 130, Cooldown: 3300, Energy: 100},
	}

	c := []model.Pokemon{
		{PokemonRaw: model.PokemonRaw{FastMove: "Water Gun", Charged1: "Hydro Pump", CP: 2000, Type1: "water", IVAtk: 15}, SpeciesID: "blastoise", BaseStats: model.BaseStats{Atk: 126}},
	}

	counters := FindBestCounters(c, gmMap, moveMap, "fire", 1)
	if len(counters) != 1 {
		t.Fatalf("expected 1 counter, got %d", len(counters))
	}

	r := counters[0]
	if r.DPS <= 0 {
		t.Errorf("DPS = %f, want > 0", r.DPS)
	}
	if r.OptimalDPS <= 0 {
		t.Errorf("OptimalDPS = %f, want > 0", r.OptimalDPS)
	}
	if r.FastMove != "Water Gun" {
		t.Errorf("FastMove = %q, want 'Water Gun'", r.FastMove)
	}
	if r.ChargedMove != "Hydro Pump" {
		t.Errorf("ChargedMove = %q, want 'Hydro Pump'", r.ChargedMove)
	}
	if r.BestFast == "" || r.BestCharged == "" {
		t.Errorf("best moves should not be empty")
	}
}
