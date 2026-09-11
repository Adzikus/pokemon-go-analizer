package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/adzik/pokemon/model"
	"github.com/adzik/pokemon/ods"
)

// GamemasterData represents the gamemaster.json structure (subset).
type GamemasterData struct {
	Pokemon []model.GamemasterPokemon `json:"pokemon"`
	Moves   []MoveEntry                   `json:"moves"`
}

// MoveEntry represents a single move from gamemaster.json.
type MoveEntry struct {
	MoveID     string `json:"moveId"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Power      int    `json:"power"`
	Energy     int    `json:"energy"`
	EnergyGain int    `json:"energyGain"`
	Cooldown   int    `json:"cooldown"`
	Turns      int    `json:"turns"`
}

// MetaData represents meta_rankings.json structure.
type MetaData struct {
	Rankings struct {
		GreatLeague  []string `json:"great_league_overall"`
		UltraLeague  []string `json:"ultra_league_overall"`
		MasterLeague []string `json:"master_league_overall"`
		Raid         []string `json:"raid_attackers"`
	} `json:"rankings"`
}

func LoadGamemaster(path string) (*GamemasterData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var gm GamemasterData
	if err := json.Unmarshal(data, &gm); err != nil {
		return nil, err
	}
	return &gm, nil
}

func LoadMeta(path string) (*MetaData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m MetaData
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func BuildGamemasterMap(gm *GamemasterData) map[string]*model.GamemasterPokemon {
	m := make(map[string]*model.GamemasterPokemon, len(gm.Pokemon))
	for i := range gm.Pokemon {
		m[gm.Pokemon[i].SpeciesID] = &gm.Pokemon[i]
	}
	return m
}

func BuildMetaTiers(m *MetaData) map[string]string {
	tiers := make(map[string]string)

	for _, list := range [][]string{
		m.Rankings.GreatLeague,
		m.Rankings.UltraLeague,
		m.Rankings.MasterLeague,
	} {
		for pos, sid := range list {
			if existing, ok := tiers[sid]; ok && existing == "S" {
				continue
			}
			switch {
			case pos < 10:
				tiers[sid] = "S"
			case pos < 20:
				if tiers[sid] != "S" {
					tiers[sid] = "A"
				}
			case pos < 30:
				if tiers[sid] == "" {
					tiers[sid] = "B"
				}
			default:
				if tiers[sid] == "" {
					tiers[sid] = "C"
				}
			}
		}
	}

	for pos, sid := range m.Rankings.Raid {
		if existing, ok := tiers[sid]; ok && existing == "S" {
			continue
		}
		switch {
		case pos < 10:
			tiers[sid] = "S"
		case pos < 20:
			if tiers[sid] != "S" {
				tiers[sid] = "A"
			}
		}
	}

	suffixes := []string{"", "_shadow", "_mega", "_mega_x", "_mega_y"}
	baseTiers := make(map[string]string)
	for sid, tier := range tiers {
		base := sid
		for _, sfx := range suffixes[1:] {
			base = strings.TrimSuffix(base, sfx)
		}
		if existing, ok := baseTiers[base]; !ok || tier == "S" || (tier == "A" && existing != "S") {
			baseTiers[base] = tier
		}
	}
	for base, tier := range baseTiers {
		for _, sfx := range suffixes {
			variant := base + sfx
			if _, ok := tiers[variant]; !ok {
				tiers[variant] = tier
			}
		}
	}

	return tiers
}

func NormalizeName(name string) string {
	name = strings.TrimSpace(name)
	if idx := strings.Index(name, " - "); idx > 0 {
		form := name[idx+3:]
		base := name[:idx]
		return fmt.Sprintf("%s_%s", strings.ToLower(strings.ReplaceAll(base, " ", "_")),
			strings.ToLower(strings.ReplaceAll(form, " ", "_")))
	}
	return strings.ToLower(strings.ReplaceAll(name, " ", "_"))
}

func ResolveSpecies(odsName string, gmMap map[string]*model.GamemasterPokemon) string {
	normalized := NormalizeName(odsName)

	if _, ok := gmMap[normalized]; ok {
		return normalized
	}

	nameToID := make(map[string]string, len(gmMap))
	for sid, gm := range gmMap {
		nameToID[strings.ToLower(gm.SpeciesName)] = sid
	}

	if id, ok := nameToID[strings.ToLower(odsName)]; ok {
		return id
	}

	words := strings.Fields(odsName)
	for i := len(words) - 1; i > 0; i-- {
		candidate := strings.Join(words[:i], " ")
		candidateNorm := NormalizeName(candidate)
		if _, ok := gmMap[candidateNorm]; ok {
			return candidateNorm
		}
		if id, ok := nameToID[strings.ToLower(candidate)]; ok {
			return id
		}
	}

	return normalized
}

func ParseCollection(sheet *ods.Sheet, gmMap map[string]*model.GamemasterPokemon, metaTiers map[string]string) []model.Pokemon {
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
		if gmPokemon != nil {
			baseStats = model.BaseStats{
				Atk: gmPokemon.BaseStats.Atk,
				Def: gmPokemon.BaseStats.Def,
				HP:  gmPokemon.BaseStats.HP,
			}
			if gmPokemon.Family != nil {
				familyID = gmPokemon.Family.ID
			}
		}

		metaTier := metaTiers[speciesNorm]

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
			FamilyID:   familyID,
			MetaTier:   metaTier,
		}

		if extraTag != "" {
			p.Tags = append(p.Tags, extraTag)
		}

		collection = append(collection, p)
	}

	return collection
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

func ParsePercent(s string) float64 {
	s = strings.TrimSpace(s)
	s = strings.Replace(s, "%", "", 1)
	s = strings.Replace(s, ",", ".", 1)
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

func SortByCP(c []model.Pokemon) []model.Pokemon {
	sorted := make([]model.Pokemon, len(c))
	copy(sorted, c)
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].CP > sorted[i].CP {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	return sorted
}

func VerifyBackupZIP(path string) error {
	r, err := ods.Read(path)
	if err != nil {
		return fmt.Errorf("otworz ODS: %w", err)
	}
	_ = r
	return nil
}

// extractKontoFromFilename extracts account name from ODS filename.
// Example: "Pokemon-GO-Adzik.ods" -> "Adzik"
func extractKontoFromFilename(filename string) string {
	// Remove .ods extension
	name := strings.TrimSuffix(filename, ".ods")
	name = strings.TrimSuffix(name, ".ODS")
	// Remove prefix "Pokemon-GO-"
	const prefix = "Pokemon-GO-"
	if strings.HasPrefix(name, prefix) {
		return name[len(prefix):]
	}
	return ""
}
