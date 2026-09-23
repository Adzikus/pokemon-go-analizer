package gamemaster

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adzik/pokemon/model"
)

const testGamemasterJSON = `{
  "pokemon": [
    {
      "speciesId": "bulbasaur",
      "speciesName": "Bulbasaur",
      "types": ["grass", "poison"],
      "baseStats": {"atk": 118, "def": 111, "hp": 128},
      "fastMoves": ["VINE_WHIP", "TACKLE"],
      "chargedMoves": ["SLUDGE_BOMB", "SEED_BOMB", "POWER_WHIP"],
      "tags": ["starter"],
      "family": {"id": "bulbasaur"}
    },
    {
      "speciesId": "bulbasaur_shadow",
      "speciesName": "Bulbasaur",
      "types": ["grass", "poison"],
      "baseStats": {"atk": 118, "def": 111, "hp": 128},
      "fastMoves": ["VINE_WHIP"],
      "chargedMoves": ["SLUDGE_BOMB"],
      "tags": ["shadoweligible"],
      "family": {"id": "bulbasaur"}
    }
  ],
  "moves": [
    {"moveId": "VINE_WHIP", "name": "Vine Whip", "type": "grass", "power": 7, "energy": 6, "energyGain": 12, "cooldown": 600},
    {"moveId": "TACKLE", "name": "Tackle", "type": "normal", "power": 5, "energy": 5, "energyGain": 10, "cooldown": 500},
    {"moveId": "SLUDGE_BOMB", "name": "Sludge Bomb", "type": "poison", "power": 80, "energy": 50, "energyGain": 0, "cooldown": 2300},
    {"moveId": "SEED_BOMB", "name": "Seed Bomb", "type": "grass", "power": 65, "energy": 45, "energyGain": 0, "cooldown": 2100},
    {"moveId": "POWER_WHIP", "name": "Power Whip", "type": "grass", "power": 90, "energy": 55, "energyGain": 0, "cooldown": 2600}
  ]
}`

const testMetaJSON = `{
  "rankings": {
    "great_league_overall": ["mewtwo", "groudon", "kyogre", "rayquaza", "dialga"],
    "ultra_league_overall": ["mewtwo", "groudon", "kyogre", "dialga", "giratina"],
    "master_league_overall": ["mewtwo", "rayquaza", "groudon", "kyogre", "dialga"],
    "raid_attackers": ["rayquaza", "mewtwo", "groudon", "kyogre", "dialga"]
  }
}`

func writeTestFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadGamemaster(t *testing.T) {
	dir := t.TempDir()
	path := writeTestFile(t, dir, "gamemaster.json", testGamemasterJSON)

	gm, err := LoadGamemaster(path)
	if err != nil {
		t.Fatalf("LoadGamemaster() error = %v", err)
	}
	if len(gm.Pokemon) != 2 {
		t.Errorf("Pokemon count = %d, want 2", len(gm.Pokemon))
	}
	if len(gm.Moves) != 5 {
		t.Errorf("Moves count = %d, want 5", len(gm.Moves))
	}
	if gm.Pokemon[0].SpeciesID != "bulbasaur" {
		t.Errorf("Pokemon[0].SpeciesID = %q, want bulbasaur", gm.Pokemon[0].SpeciesID)
	}
	if gm.Pokemon[0].BaseStats.Atk != 118 {
		t.Errorf("Pokemon[0].BaseStats.Atk = %d, want 118", gm.Pokemon[0].BaseStats.Atk)
	}
}

func TestLoadGamemasterMissing(t *testing.T) {
	_, err := LoadGamemaster("/nonexistent/path.json")
	if err == nil {
		t.Error("LoadGamemaster() with missing file should return error")
	}
}

func TestLoadGamemasterBadJSON(t *testing.T) {
	dir := t.TempDir()
	path := writeTestFile(t, dir, "bad.json", "not json {{{")

	_, err := LoadGamemaster(path)
	if err == nil {
		t.Error("LoadGamemaster() with bad JSON should return error")
	}
}

func TestLoadMeta(t *testing.T) {
	dir := t.TempDir()
	path := writeTestFile(t, dir, "meta.json", testMetaJSON)

	meta, err := LoadMeta(path)
	if err != nil {
		t.Fatalf("LoadMeta() error = %v", err)
	}
	if len(meta.Rankings.GreatLeague) != 5 {
		t.Errorf("GreatLeague count = %d, want 5", len(meta.Rankings.GreatLeague))
	}
	if meta.Rankings.GreatLeague[0] != "mewtwo" {
		t.Errorf("GreatLeague[0] = %q, want mewtwo", meta.Rankings.GreatLeague[0])
	}
	if len(meta.Rankings.Raid) != 5 {
		t.Errorf("Raid count = %d, want 5", len(meta.Rankings.Raid))
	}
}

func TestLoadMetaMissing(t *testing.T) {
	_, err := LoadMeta("/nonexistent/path.json")
	if err == nil {
		t.Error("LoadMeta() with missing file should return error")
	}
}

func TestBuildGamemasterMap(t *testing.T) {
	gm := &GamemasterData{
		Pokemon: []model.GamemasterPokemon{
			{SpeciesID: "bulbasaur", SpeciesName: "Bulbasaur"},
			{SpeciesID: "charmander", SpeciesName: "Charmander"},
		},
	}

	m := BuildGamemasterMap(gm)
	if len(m) != 2 {
		t.Fatalf("map length = %d, want 2", len(m))
	}
	if m["bulbasaur"] == nil {
		t.Error("map[bulbasaur] is nil")
	}
	if m["charmander"] == nil {
		t.Error("map[charmander] is nil")
	}
	if m["missing"] != nil {
		t.Error("map[missing] should be nil")
	}
}

func TestBuildMoveMap(t *testing.T) {
	gm := &GamemasterData{
		Moves: []MoveEntry{
			{MoveID: "TACKLE", Name: "Tackle"},
			{MoveID: "VINE_WHIP", Name: "Vine Whip"},
		},
	}

	m := BuildMoveMap(gm)
	if len(m) != 2 {
		t.Fatalf("map length = %d, want 2", len(m))
	}
	if m["TACKLE"] == nil {
		t.Error("map[TACKLE] is nil")
	}
	if m["TACKLE"].Name != "Tackle" {
		t.Errorf("map[TACKLE].Name = %q, want Tackle", m["TACKLE"].Name)
	}
}

func TestBuildGamemasterMapEmpty(t *testing.T) {
	gm := &GamemasterData{}
	m := BuildGamemasterMap(gm)
	if len(m) != 0 {
		t.Errorf("map length = %d, want 0", len(m))
	}
}
