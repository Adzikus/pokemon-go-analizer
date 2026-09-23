package analysis

import (
	"testing"

	"github.com/adzik/pokemon/config"
	"github.com/adzik/pokemon/model"
)

func TestGetLeagueCaps_Default(t *testing.T) {
	config.Global = nil
	caps := GetLeagueCaps()
	if caps["GL"] != 1500 {
		t.Errorf("GL = %d, want 1500", caps["GL"])
	}
	if caps["UL"] != 2500 {
		t.Errorf("UL = %d, want 2500", caps["UL"])
	}
	if caps["ML"] != 0 {
		t.Errorf("ML = %d, want 0", caps["ML"])
	}
}

func TestGetLeagueCaps_FromConfig(t *testing.T) {
	config.Global = &config.Config{
		PVP: config.PVPConfig{
			LeagueCaps: map[string]int{"GL": 1500, "UL": 2500, "ML": 0},
		},
	}
	caps := GetLeagueCaps()
	if caps["GL"] != 1500 {
		t.Errorf("GL = %d, want 1500", caps["GL"])
	}
}

func TestLeagueTier(t *testing.T) {
	p := model.Pokemon{
		GLTier: "S",
		ULTier: "A",
		MLTier: "B+",
	}

	tests := []struct {
		league string
		want   string
	}{
		{"GL", "S"},
		{"UL", "A"},
		{"ML", "B+"},
		{"unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.league, func(t *testing.T) {
			got := LeagueTier(p, tt.league)
			if got != tt.want {
				t.Errorf("LeagueTier(%s) = %q, want %q", tt.league, got, tt.want)
			}
		})
	}
}

func TestBuildPvPTeams_FiltersByCP(t *testing.T) {
	config.Global = config.Default()
	c := []model.Pokemon{
		{PokemonRaw: model.PokemonRaw{UID: "P-1", CP: 1500}, SpeciesID: "a", Types: []string{"water"}, GLTier: "S", DevScore: 90, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
		{PokemonRaw: model.PokemonRaw{UID: "P-2", CP: 1000}, SpeciesID: "b", Types: []string{"fire"}, GLTier: "S", DevScore: 85, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
		{PokemonRaw: model.PokemonRaw{UID: "P-3", CP: 1400}, SpeciesID: "c", Types: []string{"grass"}, GLTier: "S", DevScore: 80, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
		{PokemonRaw: model.PokemonRaw{UID: "P-4", CP: 1350}, SpeciesID: "d", Types: []string{"electric"}, GLTier: "S", DevScore: 75, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
	}

	teams := BuildPvPTeams(c, "GL")
	if len(teams) == 0 {
		t.Fatal("expected at least 1 team")
	}

	for _, team := range teams {
		for _, p := range team.Members {
			if p.CP > 1500 || p.CP < 1300 {
				t.Errorf("member %s CP=%d outside GL range [1300,1500]", p.UID, p.CP)
			}
		}
	}
}

func TestBuildPvPTeams_FiltersByTier(t *testing.T) {
	config.Global = config.Default()
	c := []model.Pokemon{
		{PokemonRaw: model.PokemonRaw{UID: "P-1", CP: 1500}, SpeciesID: "a", Types: []string{"water"}, GLTier: "S", DevScore: 90, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
		{PokemonRaw: model.PokemonRaw{UID: "P-2", CP: 1480}, SpeciesID: "b", Types: []string{"fire"}, GLTier: "D", DevScore: 85, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
		{PokemonRaw: model.PokemonRaw{UID: "P-3", CP: 1460}, SpeciesID: "c", Types: []string{"grass"}, GLTier: "C", DevScore: 80, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
	}

	teams := BuildPvPTeams(c, "GL")
	for _, team := range teams {
		for _, p := range team.Members {
			if p.GLTier == "D" || p.GLTier == "C" {
				t.Errorf("member %s has tier %s (below B), should be excluded", p.UID, p.GLTier)
			}
		}
	}
}

func TestBuildPvPTeams_SortsByScore(t *testing.T) {
	config.Global = config.Default()
	c := []model.Pokemon{
		{PokemonRaw: model.PokemonRaw{UID: "P-1", CP: 1500}, SpeciesID: "a", Types: []string{"water"}, GLTier: "S", DevScore: 70, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
		{PokemonRaw: model.PokemonRaw{UID: "P-2", CP: 1498}, SpeciesID: "b", Types: []string{"fire"}, GLTier: "S", DevScore: 90, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
		{PokemonRaw: model.PokemonRaw{UID: "P-3", CP: 1495}, SpeciesID: "c", Types: []string{"grass"}, GLTier: "S", DevScore: 80, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
	}

	teams := BuildPvPTeams(c, "GL")
	if len(teams) == 0 {
		t.Fatal("expected at least 1 team")
	}
	// P-2 (90) should be first
	if teams[0].Members[0].UID != "P-2" {
		t.Errorf("first member = %s, want P-2 (highest score)", teams[0].Members[0].UID)
	}
}

func TestBuildPvPTeams_Max3Teams(t *testing.T) {
	config.Global = config.Default()
	c := make([]model.Pokemon, 12)
	for i := range c {
		c[i] = model.Pokemon{
			PokemonRaw: model.PokemonRaw{UID: "P-" + string(rune('A'+i)), CP: 1500},
			SpeciesID:  "species_" + string(rune('a'+i)),
			Types:      []string{"type_" + string(rune('A'+i))},
			GLTier:     "S",
			DevScore:   float64(90 - i*5),
			BaseStats:  model.BaseStats{Atk: 150, Def: 100},
		}
	}

	teams := BuildPvPTeams(c, "GL")
	if len(teams) > 3 {
		t.Errorf("got %d teams, want max 3", len(teams))
	}
}

func TestBuildPvPTeams_NoCandidates(t *testing.T) {
	config.Global = config.Default()
	teams := BuildPvPTeams(nil, "GL")
	if len(teams) != 0 {
		t.Errorf("expected 0 teams, got %d", len(teams))
	}
}

func TestBuildPvPTeams_MasterLeague(t *testing.T) {
	config.Global = config.Default()
	c := []model.Pokemon{
		{PokemonRaw: model.PokemonRaw{UID: "P-1", CP: 2500}, SpeciesID: "a", Types: []string{"dragon"}, MLTier: "S", DevScore: 95, BaseStats: model.BaseStats{Atk: 200, Def: 100}},
		{PokemonRaw: model.PokemonRaw{UID: "P-2", CP: 4000}, SpeciesID: "b", Types: []string{"psychic"}, MLTier: "S", DevScore: 90, BaseStats: model.BaseStats{Atk: 200, Def: 100}},
		{PokemonRaw: model.PokemonRaw{UID: "P-3", CP: 6000}, SpeciesID: "c", Types: []string{"steel"}, MLTier: "A", DevScore: 85, BaseStats: model.BaseStats{Atk: 200, Def: 100}},
	}

	teams := BuildPvPTeams(c, "ML")
	if len(teams) == 0 {
		t.Fatal("expected at least 1 team (ML has no CP limit)")
	}
	for _, team := range teams {
		for _, p := range team.Members {
			if p.CP < 2500 {
				t.Errorf("ML member %s CP=%d below 2500 (should pass without cap)", p.UID, p.CP)
			}
		}
	}
}

func TestBuildPvPTeams_TeamNaming(t *testing.T) {
	config.Global = config.Default()
	c := make([]model.Pokemon, 9)
	for i := range c {
		c[i] = model.Pokemon{
			PokemonRaw: model.PokemonRaw{UID: "P-" + string(rune('A'+i)), CP: 1500},
			SpeciesID:  "species_" + string(rune('a'+i)),
			Types:      []string{"type_" + string(rune('A'+i))},
			GLTier:     "S",
			DevScore:   float64(90 - i*5),
			BaseStats:  model.BaseStats{Atk: 150, Def: 100},
		}
	}

	teams := BuildPvPTeams(c, "GL")
	expected := []string{"Druzyna A", "Druzyna B", "Druzyna C"}
	for i, team := range teams {
		if i >= len(expected) {
			break
		}
		if team.Name != expected[i] {
			t.Errorf("team %d Name = %q, want %q", i, team.Name, expected[i])
		}
	}
}

func TestBuildPvPTeams_TypeDiversity(t *testing.T) {
	config.Global = config.Default()
	c := []model.Pokemon{
		{PokemonRaw: model.PokemonRaw{UID: "P-1", CP: 1500}, SpeciesID: "a", Types: []string{"water", "ice"}, GLTier: "S", DevScore: 95, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
		{PokemonRaw: model.PokemonRaw{UID: "P-2", CP: 1498}, SpeciesID: "b", Types: []string{"fire"}, GLTier: "S", DevScore: 90, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
		{PokemonRaw: model.PokemonRaw{UID: "P-3", CP: 1495}, SpeciesID: "c", Types: []string{"grass"}, GLTier: "S", DevScore: 85, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
		{PokemonRaw: model.PokemonRaw{UID: "P-4", CP: 1490}, SpeciesID: "d", Types: []string{"water"}, GLTier: "S", DevScore: 80, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
		{PokemonRaw: model.PokemonRaw{UID: "P-5", CP: 1485}, SpeciesID: "e", Types: []string{"electric"}, GLTier: "S", DevScore: 75, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
		{PokemonRaw: model.PokemonRaw{UID: "P-6", CP: 1480}, SpeciesID: "f", Types: []string{"dragon"}, GLTier: "S", DevScore: 70, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
	}

	teams := BuildPvPTeams(c, "GL")
	if len(teams) == 0 {
		t.Fatal("expected at least 1 team")
	}

	// Check first team has diverse types
	team := teams[0]
	types := map[string]bool{}
	for _, p := range team.Members {
		for _, tp := range p.Types {
			if types[tp] {
				t.Errorf("team A has duplicate type %s", tp)
			}
			types[tp] = true
		}
	}
}

func TestDetectStyle_Aggressive(t *testing.T) {
	members := []pvpScored{
		{pokemon: model.Pokemon{BaseStats: model.BaseStats{Atk: 200, Def: 80}}},
		{pokemon: model.Pokemon{BaseStats: model.BaseStats{Atk: 180, Def: 90}}},
	}
	got := detectStyle(members)
	if got != "aggressive" {
		t.Errorf("detectStyle = %q, want 'aggressive'", got)
	}
}

func TestDetectStyle_Defensive(t *testing.T) {
	members := []pvpScored{
		{pokemon: model.Pokemon{BaseStats: model.BaseStats{Atk: 80, Def: 200}}},
		{pokemon: model.Pokemon{BaseStats: model.BaseStats{Atk: 90, Def: 180}}},
	}
	got := detectStyle(members)
	if got != "defensive" {
		t.Errorf("detectStyle = %q, want 'defensive'", got)
	}
}

func TestDetectStyle_Balanced(t *testing.T) {
	members := []pvpScored{
		{pokemon: model.Pokemon{BaseStats: model.BaseStats{Atk: 150, Def: 130}}},
		{pokemon: model.Pokemon{BaseStats: model.BaseStats{Atk: 140, Def: 140}}},
	}
	got := detectStyle(members)
	if got != "balanced" {
		t.Errorf("detectStyle = %q, want 'balanced'", got)
	}
}

func TestDetectStyle_Empty(t *testing.T) {
	got := detectStyle(nil)
	if got != "balanced" {
		t.Errorf("detectStyle(nil) = %q, want 'balanced'", got)
	}
}

func TestBuildPvPTeams_StyleDetection(t *testing.T) {
	config.Global = config.Default()
	c := []model.Pokemon{
		{PokemonRaw: model.PokemonRaw{UID: "P-1", CP: 1500}, SpeciesID: "a", Types: []string{"water"}, GLTier: "S", DevScore: 95, BaseStats: model.BaseStats{Atk: 200, Def: 80}},
		{PokemonRaw: model.PokemonRaw{UID: "P-2", CP: 1498}, SpeciesID: "b", Types: []string{"fire"}, GLTier: "S", DevScore: 90, BaseStats: model.BaseStats{Atk: 180, Def: 90}},
		{PokemonRaw: model.PokemonRaw{UID: "P-3", CP: 1495}, SpeciesID: "c", Types: []string{"grass"}, GLTier: "S", DevScore: 85, BaseStats: model.BaseStats{Atk: 190, Def: 85}},
	}

	teams := BuildPvPTeams(c, "GL")
	if len(teams) == 0 {
		t.Fatal("expected at least 1 team")
	}
	// High ATK/DEF ratio → aggressive
	if teams[0].Style != "aggressive" {
		t.Errorf("team style = %q, want 'aggressive'", teams[0].Style)
	}
}

func TestBuildPvPTeams_TypeOverlapPenalty(t *testing.T) {
	config.Global = config.Default()
	c := []model.Pokemon{
		{PokemonRaw: model.PokemonRaw{UID: "P-1", CP: 1500}, SpeciesID: "a", Types: []string{"water"}, GLTier: "S", DevScore: 95, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
		{PokemonRaw: model.PokemonRaw{UID: "P-2", CP: 1498}, SpeciesID: "b", Types: []string{"fire"}, GLTier: "S", DevScore: 90, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
		{PokemonRaw: model.PokemonRaw{UID: "P-3", CP: 1495}, SpeciesID: "c", Types: []string{"grass"}, GLTier: "S", DevScore: 85, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
		{PokemonRaw: model.PokemonRaw{UID: "P-4", CP: 1490}, SpeciesID: "d", Types: []string{"fire"}, GLTier: "S", DevScore: 80, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
	}

	teams := BuildPvPTeams(c, "GL")
	if len(teams) == 0 {
		t.Fatal("expected at least 1 team")
	}

	// P-4 (fire) has lower score than P-3 but P-4 has same type as P-2
	// So P-3 should be in team A, P-4 should not overlap
	teamA := teams[0]
	for i, p1 := range teamA.Members {
		for j, p2 := range teamA.Members {
			if i >= j {
				continue
			}
			for _, t1 := range p1.Types {
				for _, t2 := range p2.Types {
					if t1 == t2 {
						t.Errorf("team A has type overlap: %s and %s both have %s", p1.UID, p2.UID, t1)
					}
				}
			}
		}
	}
}
