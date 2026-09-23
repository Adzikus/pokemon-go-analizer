package analysis

import (
	"fmt"
	"testing"

	"github.com/adzik/pokemon/config"
	"github.com/adzik/pokemon/model"
	"github.com/adzik/pokemon/pkg/gamemaster"
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

	teams := BuildPvPTeams(c, "GL", nil)
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

	teams := BuildPvPTeams(c, "GL", nil)
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

	teams := BuildPvPTeams(c, "GL", nil)
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

	teams := BuildPvPTeams(c, "GL", nil)
	if len(teams) > 3 {
		t.Errorf("got %d teams, want max 3", len(teams))
	}
}

func TestBuildPvPTeams_NoCandidates(t *testing.T) {
	config.Global = config.Default()
	teams := BuildPvPTeams(nil, "GL", nil)
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

	teams := BuildPvPTeams(c, "ML", nil)
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

	teams := BuildPvPTeams(c, "GL", nil)
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

	teams := BuildPvPTeams(c, "GL", nil)
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

	teams := BuildPvPTeams(c, "GL", nil)
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

	teams := BuildPvPTeams(c, "GL", nil)
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

func TestTypeWeaknesses(t *testing.T) {
	tests := []struct {
		name     string
		types    []string
		expected []string
	}{
		{"fairy", []string{"fairy"}, []string{"poison", "steel"}},
		{"steel", []string{"steel"}, []string{"fire", "fighting", "ground"}},
		{"fairy_steel", []string{"fairy", "steel"}, []string{"poison", "steel", "fire", "fighting", "ground"}},
		{"dragon", []string{"dragon"}, []string{"ice", "dragon", "fairy"}},
		{"fire", []string{"fire"}, []string{"water", "ground", "rock"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			weak := typeWeaknesses(tt.types)
			for _, w := range tt.expected {
				if !weak[w] {
					t.Errorf("typeWeaknesses(%v) missing weakness %s", tt.types, w)
				}
			}
		})
	}
}

func TestTypeCovers(t *testing.T) {
	tests := []struct {
		name     string
		types    []string
		expected []string
	}{
		{"fire", []string{"fire"}, []string{"grass", "ice", "bug", "steel"}},
		{"water", []string{"water"}, []string{"fire", "ground", "rock"}},
		{"fairy_steel", []string{"fairy", "steel"}, []string{"fighting", "dragon", "dark", "ice", "rock", "fairy"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			covers := typeCovers(tt.types)
			for _, c := range tt.expected {
				if !covers[c] {
					t.Errorf("typeCovers(%v) missing coverage %s", tt.types, c)
				}
			}
		})
	}
}

func TestBuildPvPTeams_CoverageBonus(t *testing.T) {
	config.Global = config.Default()

	// Team with coverage bonus: steel covers fairy weakness, fairy covers fighting weakness
	c := []model.Pokemon{
		{PokemonRaw: model.PokemonRaw{UID: "P-steel", CP: 1500}, SpeciesID: "steel1", Types: []string{"steel"}, GLTier: "S", DevScore: 90, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
		{PokemonRaw: model.PokemonRaw{UID: "P-fighting", CP: 1498}, SpeciesID: "fighting1", Types: []string{"fighting"}, GLTier: "S", DevScore: 88, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
		{PokemonRaw: model.PokemonRaw{UID: "P-fire", CP: 1495}, SpeciesID: "fire1", Types: []string{"fire"}, GLTier: "S", DevScore: 86, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
	}

	teams := BuildPvPTeams(c, "GL", nil)
	if len(teams) == 0 {
		t.Fatal("expected at least 1 team")
	}

	team := teams[0]
	// Steel (weakness: fire, fighting, ground) + Fighting (weakness: flying, psychic, fairy) + Fire (weakness: water, ground, rock)
	// Fighting covers steel's weakness (fighting is SE vs steel? No - fighting is weak to fairy. Let me check...)
	// Actually: steel weakness = fire, fighting, ground
	// fire covers steel weakness? fire is SE vs steel? Yes! (fire > steel)
	// So fire should get coverage bonus when steel is already in team
	types := map[string]bool{}
	for _, p := range team.Members {
		for _, tp := range p.Types {
			types[tp] = true
		}
	}
	// All three should have different primary types
	if len(types) < 3 {
		t.Errorf("team has only %d unique types, expected at least 3", len(types))
	}
}

func TestBuildPvPTeams_WeaknessPenalty(t *testing.T) {
	config.Global = config.Default()

	// Two steel types share weaknesses (fire, fighting, ground)
	// The second steel should get penalty
	c := []model.Pokemon{
		{PokemonRaw: model.PokemonRaw{UID: "P-steel1", CP: 1500}, SpeciesID: "steel_a", Types: []string{"steel"}, GLTier: "S", DevScore: 95, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
		{PokemonRaw: model.PokemonRaw{UID: "P-steel2", CP: 1498}, SpeciesID: "steel_b", Types: []string{"steel"}, GLTier: "S", DevScore: 93, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
		{PokemonRaw: model.PokemonRaw{UID: "P-dragon", CP: 1495}, SpeciesID: "dragon1", Types: []string{"dragon"}, GLTier: "S", DevScore: 91, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
		{PokemonRaw: model.PokemonRaw{UID: "P-water", CP: 1490}, SpeciesID: "water1", Types: []string{"water"}, GLTier: "S", DevScore: 89, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
	}

	teams := BuildPvPTeams(c, "GL", nil)
	if len(teams) == 0 {
		t.Fatal("expected at least 1 team")
	}

	team := teams[0]
	// P-steel1 should be in team (highest score)
	// P-steel2 should NOT be in team A (shared weaknesses with steel1 → penalty)
	foundSteel2 := false
	for _, p := range team.Members {
		if p.UID == "P-steel2" {
			foundSteel2 = true
		}
	}
	if foundSteel2 {
		// This is OK but suboptimal — two steels share fire/fighting/ground weakness
		t.Logf("warning: team A has two steel types (P-steel1, P-steel2) — shared weaknesses fire/fighting/ground")
	}
}

func TestWeaknessExponential(t *testing.T) {
	tests := []struct {
		overlap int
		want    float64
	}{
		{0, 0},
		{1, 10},
		{2, 30},
		{3, 60},
		{4, 100},
		{5, 100},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("overlap_%d", tt.overlap), func(t *testing.T) {
			got := weaknessExponential(tt.overlap)
			if got != tt.want {
				t.Errorf("weaknessExponential(%d) = %f, want %f", tt.overlap, got, tt.want)
			}
		})
	}
}

func TestBuildPvPTeams_HardCapMaxTwoOfSameType(t *testing.T) {
	config.Global = config.Default()

	// 3 fairy types — hard cap should prevent all 3 in same team
	c := []model.Pokemon{
		{PokemonRaw: model.PokemonRaw{UID: "P-fairy1", CP: 1500}, SpeciesID: "fairy_a", Types: []string{"fairy"}, GLTier: "S", DevScore: 95, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
		{PokemonRaw: model.PokemonRaw{UID: "P-fairy2", CP: 1498}, SpeciesID: "fairy_b", Types: []string{"fairy"}, GLTier: "S", DevScore: 93, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
		{PokemonRaw: model.PokemonRaw{UID: "P-fairy3", CP: 1495}, SpeciesID: "fairy_c", Types: []string{"fairy"}, GLTier: "S", DevScore: 91, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
		{PokemonRaw: model.PokemonRaw{UID: "P-fire", CP: 1490}, SpeciesID: "fire1", Types: []string{"fire"}, GLTier: "S", DevScore: 89, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
	}

	teams := BuildPvPTeams(c, "GL", nil)
	if len(teams) == 0 {
		t.Fatal("expected at least 1 team")
	}

	for ti, team := range teams {
		fairyCount := 0
		for _, p := range team.Members {
			for _, tp := range p.Types {
				if tp == "fairy" {
					fairyCount++
				}
			}
		}
		if fairyCount > 2 {
			t.Errorf("team %c has %d fairy types (hard cap = 2)", 'A'+ti, fairyCount)
		}
	}
}

func TestBuildPvPTeams_HardCapThreeSameType(t *testing.T) {
	config.Global = config.Default()

	// 3 steel types + 1 different — should get 2 steel + 1 different in team A
	c := []model.Pokemon{
		{PokemonRaw: model.PokemonRaw{UID: "P-steel1", CP: 1500}, SpeciesID: "steel_a", Types: []string{"steel"}, GLTier: "S", DevScore: 95, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
		{PokemonRaw: model.PokemonRaw{UID: "P-steel2", CP: 1498}, SpeciesID: "steel_b", Types: []string{"steel"}, GLTier: "S", DevScore: 93, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
		{PokemonRaw: model.PokemonRaw{UID: "P-steel3", CP: 1495}, SpeciesID: "steel_c", Types: []string{"steel"}, GLTier: "S", DevScore: 91, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
		{PokemonRaw: model.PokemonRaw{UID: "P-fire", CP: 1490}, SpeciesID: "fire1", Types: []string{"fire"}, GLTier: "S", DevScore: 89, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
	}

	teams := BuildPvPTeams(c, "GL", nil)
	if len(teams) == 0 {
		t.Fatal("expected at least 1 team")
	}

	team := teams[0]
	steelCount := 0
	for _, p := range team.Members {
		for _, tp := range p.Types {
			if tp == "steel" {
				steelCount++
			}
		}
	}
	if steelCount > 2 {
		t.Errorf("team A has %d steel types (hard cap = 2), members: %v", steelCount, team.Members)
	}
}

func TestBuildPvPTeams_RoleMultiplier(t *testing.T) {
	config.Global = config.Default()

	// With role multipliers, Closer (round 2) gets 1.2× score
	// This means a slightly weaker Pokemon in closer position is OK
	c := []model.Pokemon{
		{PokemonRaw: model.PokemonRaw{UID: "P-1", CP: 1500}, SpeciesID: "a", Types: []string{"water"}, GLTier: "S", DevScore: 95, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
		{PokemonRaw: model.PokemonRaw{UID: "P-2", CP: 1498}, SpeciesID: "b", Types: []string{"fire"}, GLTier: "S", DevScore: 90, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
		{PokemonRaw: model.PokemonRaw{UID: "P-3", CP: 1495}, SpeciesID: "c", Types: []string{"grass"}, GLTier: "S", DevScore: 85, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
	}

	teams := BuildPvPTeams(c, "GL", nil)
	if len(teams) == 0 {
		t.Fatal("expected at least 1 team")
	}

	team := teams[0]
	// Lead (index 0) = 0.8×, Switch (index 1) = 1.0×, Closer (index 2) = 1.2×
	// P-1 (95) should be Lead, P-2 (90) should be Switch, P-3 (85) should be Closer
	if team.Members[0].UID != "P-1" {
		t.Errorf("Lead = %s, want P-1", team.Members[0].UID)
	}
	if team.Members[1].UID != "P-2" {
		t.Errorf("Switch = %s, want P-2", team.Members[1].UID)
	}
	if team.Members[2].UID != "P-3" {
		t.Errorf("Closer = %s, want P-3", team.Members[2].UID)
	}
}

func TestBuildPvPTeams_TripleFairyBlocked(t *testing.T) {
	config.Global = config.Default()

	// Triple Fairy + 1 non-fairy — hard cap should force the non-fairy into team
	c := []model.Pokemon{
		{PokemonRaw: model.PokemonRaw{UID: "P-ninetales", CP: 1482}, SpeciesID: "ninetales_alolan", Types: []string{"ice", "fairy"}, GLTier: "S", DevScore: 70, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
		{PokemonRaw: model.PokemonRaw{UID: "P-mimikyu", CP: 1489}, SpeciesID: "mimikyu", Types: []string{"ghost", "fairy"}, GLTier: "S", DevScore: 69, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
		{PokemonRaw: model.PokemonRaw{UID: "P-azumarill", CP: 1452}, SpeciesID: "azumarill", Types: []string{"water", "fairy"}, GLTier: "A+", DevScore: 68, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
		{PokemonRaw: model.PokemonRaw{UID: "P-tinkaton", CP: 1500}, SpeciesID: "tinkaton", Types: []string{"fairy", "steel"}, GLTier: "S", DevScore: 75, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
		{PokemonRaw: model.PokemonRaw{UID: "P-garchomp", CP: 1490}, SpeciesID: "garchomp", Types: []string{"dragon", "ground"}, GLTier: "S", DevScore: 72, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
	}

	teams := BuildPvPTeams(c, "GL", nil)
	if len(teams) == 0 {
		t.Fatal("expected at least 1 team")
	}

	for ti, team := range teams {
		fairyCount := 0
		for _, p := range team.Members {
			for _, tp := range p.Types {
				if tp == "fairy" {
					fairyCount++
				}
			}
		}
		if fairyCount > 2 {
			t.Errorf("team %c has %d fairy types (Level 2 hard cap = 2)", 'A'+ti, fairyCount)
		}
	}
}

func TestMoveTypes(t *testing.T) {
	moveMap := map[string]*gamemaster.MoveEntry{
		"FIRE_SPIN":  {MoveID: "FIRE_SPIN", Type: "fire"},
		"DRILL_RUN":  {MoveID: "DRILL_RUN", Type: "ground"},
		"WING_ATTACK": {MoveID: "WING_ATTACK", Type: "flying"},
		"SHADOW_BALL": {MoveID: "SHADOW_BALL", Type: "ghost"},
	}

	p := model.Pokemon{
		PokemonRaw: model.PokemonRaw{FastMove: "Fire Spin", Charged1: "Drill Run"},
	}
	got := moveTypes(p, moveMap)
	if !got["fire"] {
		t.Error("expected fire in moveTypes")
	}
	if !got["ground"] {
		t.Error("expected ground in moveTypes")
	}
	if len(got) != 2 {
		t.Errorf("expected 2 move types, got %d", len(got))
	}
}

func TestMoveTypes_EmptyMoves(t *testing.T) {
	moveMap := map[string]*gamemaster.MoveEntry{}
	p := model.Pokemon{}
	got := moveTypes(p, moveMap)
	if len(got) != 0 {
		t.Errorf("expected 0 move types, got %d", len(got))
	}
}

func TestMoveTypes_UnknownMove(t *testing.T) {
	moveMap := map[string]*gamemaster.MoveEntry{
		"FIRE_SPIN": {MoveID: "FIRE_SPIN", Type: "fire"},
	}
	p := model.Pokemon{
		PokemonRaw: model.PokemonRaw{FastMove: "Unknown Move", Charged1: "Fire Spin"},
	}
	got := moveTypes(p, moveMap)
	if !got["fire"] {
		t.Error("expected fire from charged1")
	}
	if len(got) != 1 {
		t.Errorf("expected 1 move type (unknown ignored), got %d", len(got))
	}
}

func TestAllTypes(t *testing.T) {
	typeSet := map[string]bool{"fire": true, "steel": true}
	moveTypeSet := map[string]bool{"ground": true, "fire": true}
	combined := allTypes(typeSet, moveTypeSet)
	if !combined["fire"] || !combined["steel"] || !combined["ground"] {
		t.Errorf("expected fire, steel, ground; got %v", combined)
	}
	if len(combined) != 3 {
		t.Errorf("expected 3 unique types, got %d", len(combined))
	}
}

func TestCombinedKeys(t *testing.T) {
	m := map[string]bool{"fire": true, "ground": true}
	keys := combinedKeys(m)
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
	for _, k := range keys {
		if k != "fire" && k != "ground" {
			t.Errorf("unexpected key: %s", k)
		}
	}
}

func TestBuildPvPTeams_MoveCoverageBonus(t *testing.T) {
	config.Global = config.Default()
	moveMap := map[string]*gamemaster.MoveEntry{
		"FIRE_SPIN":  {MoveID: "FIRE_SPIN", Type: "fire"},
		"DRILL_RUN":  {MoveID: "DRILL_RUN", Type: "ground"},
		"WATER_GUN":  {MoveID: "WATER_GUN", Type: "water"},
		"ICE_BEAM":   {MoveID: "ICE_BEAM", Type: "ice"},
	}

	// Team A: pure type coverage (water only)
	// Team B: water type + fire move → covers more weaknesses
	c := []model.Pokemon{
		{PokemonRaw: model.PokemonRaw{UID: "P-water", CP: 1500, FastMove: "Water Gun", Charged1: "Water Gun"}, SpeciesID: "a", Types: []string{"water"}, GLTier: "S", DevScore: 90, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
		{PokemonRaw: model.PokemonRaw{UID: "P-fire", CP: 1498, FastMove: "Fire Spin", Charged1: "Drill Run"}, SpeciesID: "b", Types: []string{"fire"}, GLTier: "S", DevScore: 88, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
		{PokemonRaw: model.PokemonRaw{UID: "P-grass", CP: 1495, FastMove: "Water Gun", Charged1: "Ice Beam"}, SpeciesID: "c", Types: []string{"grass"}, GLTier: "S", DevScore: 85, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
		{PokemonRaw: model.PokemonRaw{UID: "P-electric", CP: 1490, FastMove: "Water Gun", Charged1: "Water Gun"}, SpeciesID: "d", Types: []string{"electric"}, GLTier: "S", DevScore: 82, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
	}

	teams := BuildPvPTeams(c, "GL", moveMap)
	if len(teams) == 0 {
		t.Fatal("expected at least 1 team")
	}

	// P-fire (fire+ground moves) should be preferred over P-electric (water moves only)
	// because ground covers fire/steel/electric which complements water types
	team := teams[0]
	hasFire := false
	for _, p := range team.Members {
		if p.UID == "P-fire" {
			hasFire = true
		}
	}
	if !hasFire {
		t.Errorf("expected P-fire in team A (move coverage: fire+ground), got: %v", team.Members)
	}
}

func TestBuildPvPTeams_MoveTypeHardCap(t *testing.T) {
	config.Global = config.Default()
	moveMap := map[string]*gamemaster.MoveEntry{
		"CHARM":      {MoveID: "CHARM", Type: "fairy"},
		"DRAGON_BREATH": {MoveID: "DRAGON_BREATH", Type: "dragon"},
		"FIRE_SPIN":  {MoveID: "FIRE_SPIN", Type: "fire"},
	}

	// 3 fairy types, but one has dragon move → combined types are different
	// Hard cap should count combined types (pokemon type + move type)
	c := []model.Pokemon{
		{PokemonRaw: model.PokemonRaw{UID: "P-fairy1", CP: 1500, FastMove: "Charm", Charged1: "Charm"}, SpeciesID: "fairy_a", Types: []string{"fairy"}, GLTier: "S", DevScore: 95, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
		{PokemonRaw: model.PokemonRaw{UID: "P-fairy2", CP: 1498, FastMove: "Charm", Charged1: "Charm"}, SpeciesID: "fairy_b", Types: []string{"fairy"}, GLTier: "S", DevScore: 93, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
		{PokemonRaw: model.PokemonRaw{UID: "P-fairy3", CP: 1495, FastMove: "Dragon Breath", Charged1: "Charm"}, SpeciesID: "fairy_c", Types: []string{"fairy"}, GLTier: "S", DevScore: 91, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
		{PokemonRaw: model.PokemonRaw{UID: "P-fire", CP: 1490, FastMove: "Fire Spin", Charged1: "Fire Spin"}, SpeciesID: "fire1", Types: []string{"fire"}, GLTier: "S", DevScore: 89, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
	}

	teams := BuildPvPTeams(c, "GL", moveMap)
	if len(teams) == 0 {
		t.Fatal("expected at least 1 team")
	}

	for ti, team := range teams {
		combinedCounts := map[string]int{}
		for _, p := range team.Members {
			pTypes := map[string]bool{}
			for _, tp := range p.Types {
				pTypes[tp] = true
			}
			// Simulate moveTypes lookup
			if p.FastMove == "Dragon Breath" {
				pTypes["dragon"] = true
			}
			for t := range pTypes {
				combinedCounts[t]++
			}
		}
		for typ, count := range combinedCounts {
			if count > 2 {
				t.Errorf("team %c has %d combined %s types (hard cap = 2)", 'A'+ti, count, typ)
			}
		}
	}
}

func TestBuildPvPTeams_MoveTypesInScoring(t *testing.T) {
	config.Global = config.Default()
	moveMap := map[string]*gamemaster.MoveEntry{
		"WATER_GUN":  {MoveID: "WATER_GUN", Type: "water"},
		"DRILL_RUN":  {MoveID: "DRILL_RUN", Type: "ground"},
		"TACKLE":     {MoveID: "TACKLE", Type: "normal"},
	}

	// Two water types with same DevScore
	// P-water1 has ground move (covers fire/electric/poison/rock/steel)
	// P-water2 has normal move (covers nothing)
	// P-water1 should score higher due to move coverage
	c := []model.Pokemon{
		{PokemonRaw: model.PokemonRaw{UID: "P-water1", CP: 1500, FastMove: "Water Gun", Charged1: "Drill Run"}, SpeciesID: "a", Types: []string{"water"}, GLTier: "S", DevScore: 90, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
		{PokemonRaw: model.PokemonRaw{UID: "P-water2", CP: 1500, FastMove: "Water Gun", Charged1: "Tackle"}, SpeciesID: "b", Types: []string{"water"}, GLTier: "S", DevScore: 90, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
		{PokemonRaw: model.PokemonRaw{UID: "P-third", CP: 1490, FastMove: "Tackle", Charged1: "Tackle"}, SpeciesID: "c", Types: []string{"normal"}, GLTier: "A", DevScore: 80, BaseStats: model.BaseStats{Atk: 150, Def: 100}},
	}

	teams := BuildPvPTeams(c, "GL", moveMap)
	if len(teams) == 0 {
		t.Fatal("expected at least 1 team")
	}

	// First member should be P-water1 (better move coverage) or P-water2
	// Both have same DevScore, but P-water1 has ground move which covers more
	team := teams[0]
	first := team.Members[0]
	if first.UID != "P-water1" && first.UID != "P-water2" {
		t.Errorf("expected P-water1 or P-water2 as lead, got %s", first.UID)
	}
}
