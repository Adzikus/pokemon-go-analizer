package analysis

import (
	"testing"

	"github.com/adzik/pokemon/model"
	"github.com/adzik/pokemon/pkg/gamemaster"
)

// Helper: build minimal GamemasterData with specific Pokemon entries.
func buildTestGamemaster(entries []model.GamemasterPokemon) *gamemaster.GamemasterData {
	return &gamemaster.GamemasterData{Pokemon: entries}
}

// Helper: build minimal MetaData with Raid rankings.
func buildTestMeta(raidRanking []string) *gamemaster.MetaData {
	m := &gamemaster.MetaData{}
	m.Rankings.Raid = raidRanking
	return m
}

func legendary(id, name string) model.GamemasterPokemon {
	return model.GamemasterPokemon{
		SpeciesID:   id,
		SpeciesName: name,
		Tags:        []string{"legendary"},
	}
}

func mythical(id, name string) model.GamemasterPokemon {
	return model.GamemasterPokemon{
		SpeciesID:   id,
		SpeciesName: name,
		Tags:        []string{"mythical"},
	}
}

func shadowEligible(id, name string) model.GamemasterPokemon {
	return model.GamemasterPokemon{
		SpeciesID:   id,
		SpeciesName: name,
		Tags:        []string{"shadoweligible"},
	}
}

func common(id, name string) model.GamemasterPokemon {
	return model.GamemasterPokemon{
		SpeciesID:   id,
		SpeciesName: name,
		Tags:        []string{},
	}
}

func untradeable(id, name string) model.GamemasterPokemon {
	return model.GamemasterPokemon{
		SpeciesID:   id,
		SpeciesName: name,
		Tags:        []string{"untradeable"},
	}
}

func shadow(id, name string) model.GamemasterPokemon {
	return model.GamemasterPokemon{
		SpeciesID:   id,
		SpeciesName: name,
		Tags:        []string{"shadow"},
	}
}

func owned(speciesID string, isLucky bool) model.Pokemon {
	return model.Pokemon{
		PokemonRaw: model.PokemonRaw{Species: speciesID},
		SpeciesID:  speciesID,
		IsLucky:    isLucky,
	}
}

func testPokedex(entries ...*gamemaster.PokedexEntry) map[string]*gamemaster.PokedexEntry {
	m := map[string]*gamemaster.PokedexEntry{}
	for _, e := range entries {
		m[e.SpeciesID] = e
	}
	return m
}

func pokedexEntry(speciesID string, lucky bool) *gamemaster.PokedexEntry {
	return &gamemaster.PokedexEntry{SpeciesID: speciesID, Lucky: lucky, Caught: true}
}

func TestBuildLuckyTargets_ExcludesShadow(t *testing.T) {
	gm := buildTestGamemaster([]model.GamemasterPokemon{
		shadow("bulbasaur_shadow", "Bulbasaur"),
		common("charmander", "Charmander"),
	})
	meta := buildTestMeta([]string{"bulbasaur_shadow", "charmander"})
	collection := []model.Pokemon{}
	pokedex := testPokedex()

	targets := BuildLuckyTargets(collection, gm, meta, pokedex)

	for _, tgt := range targets {
		if tgt.SpeciesID == "bulbasaur_shadow" {
			t.Error("shadow Pokemon should be excluded from lucky targets")
		}
	}
}

func TestBuildLuckyTargets_ExcludesUntradeable(t *testing.T) {
	gm := buildTestGamemaster([]model.GamemasterPokemon{
		untradeable("mew", "Mew"),
		common("bulbasaur", "Bulbasaur"),
	})
	meta := buildTestMeta([]string{"mew", "bulbasaur"})
	collection := []model.Pokemon{}
	pokedex := testPokedex()

	targets := BuildLuckyTargets(collection, gm, meta, pokedex)

	for _, tgt := range targets {
		if tgt.SpeciesID == "mew" {
			t.Error("untradeable Pokemon should be excluded from lucky targets")
		}
	}
}

func TestBuildLuckyTargets_ExcludesOwnedLucky(t *testing.T) {
	gm := buildTestGamemaster([]model.GamemasterPokemon{
		common("bulbasaur", "Bulbasaur"),
		common("charmander", "Charmander"),
	})
	meta := buildTestMeta([]string{"bulbasaur", "charmander"})
	collection := []model.Pokemon{}
	pokedex := testPokedex(
		pokedexEntry("bulbasaur", true),
	)

	targets := BuildLuckyTargets(collection, gm, meta, pokedex)

	for _, tgt := range targets {
		if tgt.SpeciesID == "bulbasaur" {
			t.Error("species with Lucky in Pokedex should be excluded")
		}
	}
}

func TestBuildLuckyTargets_IncludesUnownedNonLucky(t *testing.T) {
	gm := buildTestGamemaster([]model.GamemasterPokemon{
		common("bulbasaur", "Bulbasaur"),
		common("charmander", "Charmander"),
	})
	meta := buildTestMeta([]string{"bulbasaur", "charmander"})
	collection := []model.Pokemon{}
	pokedex := testPokedex(
		pokedexEntry("bulbasaur", false),
	)

	targets := BuildLuckyTargets(collection, gm, meta, pokedex)

	found := false
	for _, tgt := range targets {
		if tgt.SpeciesID == "bulbasaur" {
			found = true
			break
		}
	}
	if !found {
		t.Error("owned non-lucky species should still be a lucky target")
	}
}

func TestBuildLuckyTargets_LegendaryHigherThanCommon(t *testing.T) {
	gm := buildTestGamemaster([]model.GamemasterPokemon{
		legendary("mewtwo", "Mewtwo"),
		common("bulbasaur", "Bulbasaur"),
	})
	meta := buildTestMeta([]string{"mewtwo", "bulbasaur"})
	collection := []model.Pokemon{}
	pokedex := testPokedex()

	targets := BuildLuckyTargets(collection, gm, meta, pokedex)

	if len(targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(targets))
	}
	if targets[0].SpeciesID != "mewtwo" {
		t.Errorf("first target = %q, want mewtwo (legendary should be first)", targets[0].SpeciesID)
	}
	if targets[1].SpeciesID != "bulbasaur" {
		t.Errorf("second target = %q, want bulbasaur", targets[1].SpeciesID)
	}
}

func TestBuildLuckyTargets_MetaSCommonAboveMetaALegendary(t *testing.T) {
	gm := buildTestGamemaster([]model.GamemasterPokemon{
		legendary("mewtwo", "Mewtwo"),
		common("dragonite", "Dragonite"),
	})
	meta := buildTestMeta([]string{"x", "x", "x", "x", "x", "dragonite", "x", "x", "x", "x", "x", "x", "x", "x", "x", "mewtwo"})
	collection := []model.Pokemon{}
	pokedex := testPokedex()

	targets := BuildLuckyTargets(collection, gm, meta, pokedex)

	if len(targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(targets))
	}
	if targets[0].SpeciesID != "dragonite" {
		t.Errorf("first target = %q, want dragonite (common S=20 > legendary A=19)", targets[0].SpeciesID)
	}
	if targets[1].SpeciesID != "mewtwo" {
		t.Errorf("second target = %q, want mewtwo", targets[1].SpeciesID)
	}
}

func TestBuildLuckyTargets_SortedDescending(t *testing.T) {
	gm := buildTestGamemaster([]model.GamemasterPokemon{
		common("pidgey", "Pidgey"),
		common("bulbasaur", "Bulbasaur"),
		legendary("rayquaza", "Rayquaza"),
		common("charmander", "Charmander"),
	})
	meta := buildTestMeta([]string{"rayquaza", "x", "x", "x", "x", "bulbasaur", "x", "x", "x", "x", "x", "x", "x", "x", "x", "charmander"})
	collection := []model.Pokemon{}
	pokedex := testPokedex()

	targets := BuildLuckyTargets(collection, gm, meta, pokedex)

	for i := 1; i < len(targets); i++ {
		if targets[i].Score > targets[i-1].Score {
			t.Errorf("not sorted: target[%d].Score=%f > target[%d].Score=%f",
				i, targets[i].Score, i-1, targets[i-1].Score)
		}
	}
}

func TestBuildLuckyTargets_EmptyCollection(t *testing.T) {
	gm := buildTestGamemaster([]model.GamemasterPokemon{
		common("bulbasaur", "Bulbasaur"),
		legendary("mewtwo", "Mewtwo"),
	})
	meta := buildTestMeta([]string{"mewtwo", "bulbasaur"})
	collection := []model.Pokemon{}
	pokedex := testPokedex()

	targets := BuildLuckyTargets(collection, gm, meta, pokedex)

	if len(targets) != 2 {
		t.Errorf("empty collection should include all tradeable species, got %d", len(targets))
	}
}

func TestBuildLuckyTargets_EmptyGamemaster(t *testing.T) {
	gm := buildTestGamemaster([]model.GamemasterPokemon{})
	meta := buildTestMeta([]string{})
	collection := []model.Pokemon{}
	pokedex := testPokedex()

	targets := BuildLuckyTargets(collection, gm, meta, pokedex)

	if len(targets) != 0 {
		t.Errorf("empty gamemaster should produce empty targets, got %d", len(targets))
	}
}

func TestBuildLuckyTargets_AllAlreadyLucky(t *testing.T) {
	gm := buildTestGamemaster([]model.GamemasterPokemon{
		common("bulbasaur", "Bulbasaur"),
		common("charmander", "Charmander"),
	})
	meta := buildTestMeta([]string{"bulbasaur", "charmander"})
	collection := []model.Pokemon{}
	pokedex := testPokedex(
		pokedexEntry("bulbasaur", true),
		pokedexEntry("charmander", true),
	)

	targets := BuildLuckyTargets(collection, gm, meta, pokedex)

	if len(targets) != 0 {
		t.Errorf("all species already Lucky should produce empty list, got %d", len(targets))
	}
}
