package gamemaster

import (
	"testing"

	"github.com/adzik/pokemon/model"
	"github.com/adzik/pokemon/ods"
)

func TestParseCollectionBasic(t *testing.T) {
	sheet := &ods.Sheet{
		Rows: [][]string{
			{"UID", "Species", "Nick", "Type1", "Type2", "CP", "Atk", "Def", "HP", "IV%", "Fast", "Charged1", "Charged2", "Shiny", "Lucky", "Shadow", "Purified", "Dynamax", "Traded", "Tag", "Date", "Purpose"},
			{"P-00001", "Bulbasaur", "Groszek", "Grass", "Poison", "500", "15", "14", "13", "93%", "Vine Whip", "Sludge Bomb", "Seed Bomb", "X", "", "X", "", "", "", "", "2024-01-15", "Rajdy"},
		},
	}
	gmMap := map[string]*model.GamemasterPokemon{
		"bulbasaur": {
			SpeciesID:   "bulbasaur",
			SpeciesName: "Bulbasaur",
			BaseStats:   model.BaseStats{Atk: 118, Def: 111, HP: 128},
			Family:      &model.Family{ID: "bulbasaur"},
		},
	}
	metaCats := MetaTiersByCategory{
		Raid: map[string]string{"bulbasaur": "A"},
		GL:   map[string]string{"bulbasaur": "B"},
		UL:   map[string]string{},
		ML:   map[string]string{},
		RaidRank: map[string]int{"bulbasaur": 25},
		GLRank:   map[string]int{"bulbasaur": 120},
		ULRank:   map[string]int{},
		MLRank:   map[string]int{},
	}

	collection := ParseCollection(sheet, gmMap, metaCats)

	if len(collection) != 1 {
		t.Fatalf("collection length = %d, want 1", len(collection))
	}

	p := collection[0]
	if p.UID != "P-00001" {
		t.Errorf("UID = %q, want P-00001", p.UID)
	}
	if p.Species != "Bulbasaur" {
		t.Errorf("Species = %q, want Bulbasaur", p.Species)
	}
	if p.Nick != "Groszek" {
		t.Errorf("Nick = %q, want Groszek", p.Nick)
	}
	if p.CP != 500 {
		t.Errorf("CP = %d, want 500", p.CP)
	}
	if p.IVAtk != 15 || p.IVDef != 14 || p.IVHP != 13 {
		t.Errorf("IV = %d/%d/%d, want 15/14/13", p.IVAtk, p.IVDef, p.IVHP)
	}
	if p.FastMove != "Vine Whip" {
		t.Errorf("FastMove = %q, want Vine Whip", p.FastMove)
	}
	if p.Charged1 != "Sludge Bomb" {
		t.Errorf("Charged1 = %q, want Sludge Bomb", p.Charged1)
	}
	if !p.IsShiny {
		t.Error("IsShiny should be true")
	}
	if !p.IsShadow {
		t.Error("IsShadow should be true")
	}
	if p.IsLucky {
		t.Error("IsLucky should be false")
	}
	if p.RaidTier != "A" {
		t.Errorf("RaidTier = %q, want A", p.RaidTier)
	}
	if p.GLTier != "B" {
		t.Errorf("GLTier = %q, want B", p.GLTier)
	}
	if p.FamilyID != "bulbasaur" {
		t.Errorf("FamilyID = %q, want bulbasaur", p.FamilyID)
	}
}

func TestParseCollectionSkipsHeader(t *testing.T) {
	sheet := &ods.Sheet{
		Rows: [][]string{
			{"UID", "Species", "Nick", "Type1", "Type2", "CP", "Atk", "Def", "HP", "IV%", "Fast", "Charged1"},
			{"P-00001", "Bulbasaur", "Groszek", "Grass", "Poison", "500", "15", "14", "13", "93%", "Vine Whip", "Sludge Bomb"},
		},
	}
	gmMap := map[string]*model.GamemasterPokemon{}
	metaCats := MetaTiersByCategory{
		Raid: map[string]string{}, GL: map[string]string{}, UL: map[string]string{}, ML: map[string]string{},
		RaidRank: map[string]int{}, GLRank: map[string]int{}, ULRank: map[string]int{}, MLRank: map[string]int{},
	}

	collection := ParseCollection(sheet, gmMap, metaCats)

	// Header row "UID" doesn't start with "P-" → skipped
	if len(collection) != 1 {
		t.Fatalf("collection length = %d, want 1", len(collection))
	}
	if collection[0].UID != "P-00001" {
		t.Errorf("UID = %q, want P-00001", collection[0].UID)
	}
}

func TestParseCollectionSkipsShort(t *testing.T) {
	sheet := &ods.Sheet{
		Rows: [][]string{
			{"P-00001", "Bulbasaur"}, // only 2 cells, < 11
		},
	}
	gmMap := map[string]*model.GamemasterPokemon{}
	metaCats := MetaTiersByCategory{
		Raid: map[string]string{}, GL: map[string]string{}, UL: map[string]string{}, ML: map[string]string{},
		RaidRank: map[string]int{}, GLRank: map[string]int{}, ULRank: map[string]int{}, MLRank: map[string]int{},
	}

	collection := ParseCollection(sheet, gmMap, metaCats)

	if len(collection) != 0 {
		t.Errorf("collection length = %d, want 0 (row too short)", len(collection))
	}
}

func TestParseCollectionEmpty(t *testing.T) {
	sheet := &ods.Sheet{Rows: [][]string{}}
	gmMap := map[string]*model.GamemasterPokemon{}
	metaCats := MetaTiersByCategory{
		Raid: map[string]string{}, GL: map[string]string{}, UL: map[string]string{}, ML: map[string]string{},
		RaidRank: map[string]int{}, GLRank: map[string]int{}, ULRank: map[string]int{}, MLRank: map[string]int{},
	}

	collection := ParseCollection(sheet, gmMap, metaCats)

	if len(collection) != 0 {
		t.Errorf("collection length = %d, want 0 (empty sheet)", len(collection))
	}
}
