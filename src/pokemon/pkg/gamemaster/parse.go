package gamemaster

import (
	"strconv"
	"strings"

	"github.com/adzik/pokemon/model"
	"github.com/adzik/pokemon/ods"
)

func ParseCollection(sheet *ods.Sheet, gmMap map[string]*model.GamemasterPokemon, metaCats MetaTiersByCategory) []model.Pokemon {
	var collection []model.Pokemon

	for _, row := range sheet.Rows {
		if len(row) < 11 {
			continue
		}

		uid := strings.TrimSpace(row[0])
		if !strings.HasPrefix(uid, "P-") {
			continue
		}

		species := strings.TrimSpace(row[1])
		if species == "" {
			continue
		}

		nick := strings.TrimSpace(row[2])
		type1 := strings.TrimSpace(row[3])
		type2 := strings.TrimSpace(row[4])

		cp := ParseInt(row[5])
		ivAtk := ParseInt(row[6])
		ivDef := ParseInt(row[7])
		ivHP := ParseInt(row[8])
		ivPct := 0.0
		if ivAtk+ivDef+ivHP > 0 {
			ivPct = float64(ivAtk+ivDef+ivHP) / 45.0 * 100.0
		}

		fast := strings.TrimSpace(row[10])
		ch1 := ""
		ch2 := ""
		if len(row) > 11 {
			ch1 = strings.TrimSpace(row[11])
		}
		if len(row) > 12 {
			ch2 = strings.TrimSpace(row[12])
		}

		tags := map[string]bool{}
		tagCols := map[int]string{
			13: "Shiny", 14: "Lucky", 15: "Shadow",
			16: "Purified", 17: "Dynamax", 18: "Traded",
		}
		for col, name := range tagCols {
			if col < len(row) && strings.TrimSpace(row[col]) == "X" {
				tags[name] = true
			}
		}

		extraTag := ""
		if len(row) > 19 {
			extraTag = strings.TrimSpace(row[19])
		}

		date := ""
		if len(row) > 20 {
			date = strings.TrimSpace(row[20])
		}
		purpose := ""
		if len(row) > 21 {
			purpose = strings.TrimSpace(row[21])
		}
		decision := ""
		if len(row) > 23 {
			decision = strings.TrimSpace(row[23])
		}
		notes := ""
		if len(row) > 26 {
			notes = strings.TrimSpace(row[26])
		}

		speciesNorm := ResolveSpecies(species, gmMap)
		gmPokemon, _ := gmMap[speciesNorm]

	familyID := ""
	var baseStats model.BaseStats
	var pokemonTypes []string
	if gmPokemon != nil {
		baseStats = model.BaseStats{
			Atk: gmPokemon.BaseStats.Atk,
			Def: gmPokemon.BaseStats.Def,
			HP:  gmPokemon.BaseStats.HP,
		}
		if gmPokemon.Family != nil {
			familyID = gmPokemon.Family.ID
		}
		pokemonTypes = gmPokemon.Types
	}

		p := model.Pokemon{
			PokemonRaw: model.PokemonRaw{
				UID:      uid,
				Species:  species,
				Nick:     nick,
				Type1:    type1,
				Type2:    type2,
				CP:       cp,
				IVAtk:    ivAtk,
				IVDef:    ivDef,
				IVHP:     ivHP,
				FastMove: fast,
				Charged1: ch1,
				Charged2: ch2,
				Date:     date,
				Purpose:  purpose,
				Decision: decision,
				Notes:    notes,
			},
			IVPct:      ivPct,
			IsShiny:    tags["Shiny"],
			IsLucky:    tags["Lucky"],
			IsShadow:   tags["Shadow"],
			IsPurified: tags["Purified"],
			IsDynamax:  tags["Dynamax"],
			IsTraded:   tags["Traded"],
			SpeciesID:  speciesNorm,
			BaseStats:  baseStats,
			Types:      pokemonTypes,
			FamilyID:   familyID,
			RaidTier:   metaCats.Raid[speciesNorm],
			GLTier:     metaCats.GL[speciesNorm],
			ULTier:     metaCats.UL[speciesNorm],
			MLTier:     metaCats.ML[speciesNorm],
			RaidRank:   metaCats.RaidRank[speciesNorm],
			GLRank:     metaCats.GLRank[speciesNorm],
			ULRank:     metaCats.ULRank[speciesNorm],
			MLRank:     metaCats.MLRank[speciesNorm],
		}

		if extraTag != "" {
			p.Tags = append(p.Tags, extraTag)
		}

		collection = append(collection, p)
	}

	return collection
}

type PokedexEntry struct {
	SpeciesID string
	Caught    bool
	Lucky     bool
	Shiny     bool
	Hundo     bool
	Shadow    bool
	Purified  bool
	Dynamax   bool
	Mega      bool
}

func ParsePokedex(sheet *ods.Sheet, gmMap map[string]*model.GamemasterPokemon) map[string]*PokedexEntry {
	entries := map[string]*PokedexEntry{}

	for _, row := range sheet.Rows {
		if len(row) < 6 {
			continue
		}

		species := strings.TrimSpace(row[1])
		if species == "" {
			continue
		}

		speciesID := ResolveSpecies(species, gmMap)
		if speciesID == "" {
			continue
		}
		if _, ok := gmMap[speciesID]; !ok {
			continue
		}

		e := &PokedexEntry{SpeciesID: speciesID}
		e.Caught = isTick(row, 5)
		e.Lucky = isTick(row, 6)
		e.Shiny = isTick(row, 7)
		e.Hundo = isTick(row, 8)
		e.Shadow = isTick(row, 9)
		e.Purified = isTick(row, 10)
		e.Dynamax = isTick(row, 11)
		e.Mega = isTick(row, 12)

		entries[speciesID] = e
	}

	return entries
}

func isTick(row []string, col int) bool {
	if col >= len(row) {
		return false
	}
	return strings.TrimSpace(row[col]) == "X"
}

func ParseInt(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return v
}
