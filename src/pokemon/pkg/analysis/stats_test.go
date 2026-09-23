package analysis

import (
	"testing"

	"github.com/adzik/pokemon/model"
)

func TestComputeStats_AllCounters(t *testing.T) {
	c := []model.Pokemon{
		{
			PokemonRaw: model.PokemonRaw{UID: "P-1", Species: "Bulbasaur", Purpose: "rajdy"},
			SpeciesID: "bulbasaur", FamilyID: "bulbasaur",
			IVPct: 100, IsShiny: true, IsLucky: true, IsShadow: true,
			RaidTier: "S",
		},
		{
			PokemonRaw: model.PokemonRaw{UID: "P-2", Species: "Ivysaur", Purpose: "pvp"},
			SpeciesID: "ivysaur", FamilyID: "bulbasaur",
			IVPct: 95, IsPurified: true, IsDynamax: true, IsTraded: true,
			GLTier: "A",
		},
		{
			PokemonRaw: model.PokemonRaw{UID: "P-3", Species: "Venusaur", Purpose: "kolekcja"},
			SpeciesID: "venusaur", FamilyID: "bulbasaur",
			IVPct: 80,
			ULTier: "B",
		},
		{
			PokemonRaw: model.PokemonRaw{UID: "P-4", Species: "Charmander", Purpose: "inny"},
			SpeciesID: "charmander", FamilyID: "charmander",
			IVPct: 50,
		},
	}
	stats := ComputeStats(c)

	if stats.Total != 4 {
		t.Errorf("Total = %d, want 4", stats.Total)
	}
	if stats.Shiny != 1 {
		t.Errorf("Shiny = %d, want 1", stats.Shiny)
	}
	if stats.Lucky != 1 {
		t.Errorf("Lucky = %d, want 1", stats.Lucky)
	}
	if stats.Shadow != 1 {
		t.Errorf("Shadow = %d, want 1", stats.Shadow)
	}
	if stats.Purified != 1 {
		t.Errorf("Purified = %d, want 1", stats.Purified)
	}
	if stats.Dynamax != 1 {
		t.Errorf("Dynamax = %d, want 1", stats.Dynamax)
	}
	if stats.Traded != 1 {
		t.Errorf("Traded = %d, want 1", stats.Traded)
	}
	if stats.Hundo != 1 {
		t.Errorf("Hundo = %d, want 1", stats.Hundo)
	}
	if stats.HighIV != 2 {
		t.Errorf("HighIV = %d, want 2 (100%% + 95%%)", stats.HighIV)
	}
	if stats.Families != 2 {
		t.Errorf("Families = %d, want 2 (bulbasaur, charmander)", stats.Families)
	}
}

func TestComputeStats_TierCounts(t *testing.T) {
	c := []model.Pokemon{
		{SpeciesID: "a", FamilyID: "a", RaidTier: "S"},
		{SpeciesID: "b", FamilyID: "b", GLTier: "A+"},
		{SpeciesID: "c", FamilyID: "c", ULTier: "A"},
		{SpeciesID: "d", FamilyID: "d", MLTier: "B+"},
		{SpeciesID: "e", FamilyID: "e", RaidTier: "B"},
		{SpeciesID: "f", FamilyID: "f", GLTier: "C"},
		{SpeciesID: "g", FamilyID: "g"},
		{SpeciesID: "h", FamilyID: "h", RaidTier: "D"},
	}

	stats := ComputeStats(c)

	if stats.TierS != 1 {
		t.Errorf("TierS = %d, want 1", stats.TierS)
	}
	if stats.TierA != 2 {
		t.Errorf("TierA = %d, want 2 (A+ and A)", stats.TierA)
	}
	if stats.TierB != 2 {
		t.Errorf("TierB = %d, want 2 (B+ and B)", stats.TierB)
	}
	// C and D and unranked all go to TierC (default)
	if stats.TierC != 3 {
		t.Errorf("TierC = %d, want 3 (C, D, unranked)", stats.TierC)
	}
}

func TestComputeStats_PurposeCounts(t *testing.T) {
	c := []model.Pokemon{
		{PokemonRaw: model.PokemonRaw{Purpose: "rajdy"}, SpeciesID: "a", FamilyID: "a"},
		{PokemonRaw: model.PokemonRaw{Purpose: "raid"}, SpeciesID: "b", FamilyID: "b"},
		{PokemonRaw: model.PokemonRaw{Purpose: "pvp"}, SpeciesID: "c", FamilyID: "c"},
		{PokemonRaw: model.PokemonRaw{Purpose: "GL"}, SpeciesID: "d", FamilyID: "d"},
		{PokemonRaw: model.PokemonRaw{Purpose: "ul"}, SpeciesID: "e", FamilyID: "e"},
		{PokemonRaw: model.PokemonRaw{Purpose: "ML"}, SpeciesID: "f", FamilyID: "f"},
		{PokemonRaw: model.PokemonRaw{Purpose: "kolekcja"}, SpeciesID: "g", FamilyID: "g"},
		{PokemonRaw: model.PokemonRaw{Purpose: "trade"}, SpeciesID: "h", FamilyID: "h"},
		{PokemonRaw: model.PokemonRaw{Purpose: ""}, SpeciesID: "i", FamilyID: "i"},
	}

	stats := ComputeStats(c)

	if stats.PurposeRaid != 2 {
		t.Errorf("PurposeRaid = %d, want 2 (rajdy, raid)", stats.PurposeRaid)
	}
	if stats.PurposePvP != 4 {
		t.Errorf("PurposePvP = %d, want 4 (pvp, GL, ul, ML)", stats.PurposePvP)
	}
	if stats.PurposeColl != 1 {
		t.Errorf("PurposeColl = %d, want 1", stats.PurposeColl)
	}
	if stats.PurposeOther != 2 {
		t.Errorf("PurposeOther = %d, want 2 (trade, empty)", stats.PurposeOther)
	}
}

func TestComputeStats_Empty(t *testing.T) {
	stats := ComputeStats([]model.Pokemon{})

	if stats.Total != 0 {
		t.Errorf("Total = %d, want 0", stats.Total)
	}
	if stats.Families != 0 {
		t.Errorf("Families = %d, want 0", stats.Families)
	}
}

func TestTopN_ReturnsCopy(t *testing.T) {
	original := []model.Pokemon{
		{PokemonRaw: model.PokemonRaw{UID: "P-1"}, IVPct: 90},
		{PokemonRaw: model.PokemonRaw{UID: "P-2"}, IVPct: 80},
	}
	top := TopN(original, 1)

	top[0].IVPct = 999

	if original[0].IVPct == 999 {
		t.Error("TopN returned shared slice — modifying result changed original")
	}
}

func TestTopN_NLargerThanLen(t *testing.T) {
	c := []model.Pokemon{
		{PokemonRaw: model.PokemonRaw{UID: "P-1"}},
		{PokemonRaw: model.PokemonRaw{UID: "P-2"}},
	}
	result := TopN(c, 10)

	if len(result) != 2 {
		t.Errorf("len = %d, want 2", len(result))
	}
}

func TestTopN_NSmaller(t *testing.T) {
	c := []model.Pokemon{
		{PokemonRaw: model.PokemonRaw{UID: "P-1"}},
		{PokemonRaw: model.PokemonRaw{UID: "P-2"}},
		{PokemonRaw: model.PokemonRaw{UID: "P-3"}},
	}
	result := TopN(c, 2)

	if len(result) != 2 {
		t.Errorf("len = %d, want 2", len(result))
	}
	if result[0].UID != "P-1" || result[1].UID != "P-2" {
		t.Errorf("wrong elements: %v", result)
	}
}

func TestSortByCP(t *testing.T) {
	c := []model.Pokemon{
		{PokemonRaw: model.PokemonRaw{UID: "P-1", CP: 100}},
		{PokemonRaw: model.PokemonRaw{UID: "P-2", CP: 300}},
		{PokemonRaw: model.PokemonRaw{UID: "P-3", CP: 200}},
	}

	sorted := SortByCP(c)

	if sorted[0].CP != 300 || sorted[1].CP != 200 || sorted[2].CP != 100 {
		t.Errorf("not sorted: %v", []int{sorted[0].CP, sorted[1].CP, sorted[2].CP})
	}
}

func TestSortByCP_DoesNotMutateOriginal(t *testing.T) {
	c := []model.Pokemon{
		{PokemonRaw: model.PokemonRaw{UID: "P-1", CP: 100}},
		{PokemonRaw: model.PokemonRaw{UID: "P-2", CP: 300}},
	}
	originalUID := c[0].UID
	originalCP := c[0].CP

	SortByCP(c)

	if c[0].UID != originalUID || c[0].CP != originalCP {
		t.Error("SortByCP mutated original slice")
	}
}
