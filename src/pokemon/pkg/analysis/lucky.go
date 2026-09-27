package analysis

import (
	"slices"
	"sort"

	"github.com/adzik/pokemon/model"
	"github.com/adzik/pokemon/pkg/gamemaster"
)

type LuckyTarget struct {
	SpeciesID string
	Species   string
	Score     float64
	IsLegend  bool
	HasLucky  bool
	Reason    string
}

func BuildLuckyTargets(c []model.Pokemon, gmData *gamemaster.GamemasterData, metaData *gamemaster.MetaData, pokedex map[string]*gamemaster.PokedexEntry) []LuckyTarget {
	// Step 1: build map of lucky species from Pokedex (master list)
	ownedLucky := map[string]bool{}
	for speciesID, entry := range pokedex {
		if entry.Lucky {
			ownedLucky[speciesID] = true
		}
	}

	// Step 2: build meta tier maps from Raid rankings
	// Rank 0-9 → S (20pts), 10-19 → A (14pts)
	metaScore := map[string]float64{}
	for i, sid := range metaData.Rankings.Raid {
		switch {
		case i < 10:
			metaScore[sid] = 20
		case i < 20:
			metaScore[sid] = 14
		}
	}

	// Step 3: collect candidates from gamemaster
	var targets []LuckyTarget
	for _, gm := range gmData.Pokemon {
		tags := gm.Tags

		// Exclude shadow variants
		if hasTag(tags, "shadow") {
			continue
		}
		// Exclude untradeable
		if hasTag(tags, "untradeable") {
			continue
		}

		isLegend := hasTag(tags, "legendary") || hasTag(tags, "mythical")

		// Skip if already owned as Lucky
		if ownedLucky[gm.SpeciesID] {
			continue
		}

		// Calculate score
		base := metaScore[gm.SpeciesID] // 0 if unranked
		bonus := 0.0
		if isLegend {
			bonus = 5
		}
		score := base + bonus

		if score == 0 {
			continue
		}

		reason := buildReason(gm.SpeciesName, isLegend, base)

		targets = append(targets, LuckyTarget{
			SpeciesID: gm.SpeciesID,
			Species:   gm.SpeciesName,
			Score:     score,
			IsLegend:  isLegend,
			HasLucky:  false,
			Reason:    reason,
		})
	}

	// Step 4: sort by score descending
	sort.Slice(targets, func(i, j int) bool {
		return targets[i].Score > targets[j].Score
	})

	return targets
}

func hasTag(tags []string, tag string) bool {
	return slices.Contains(tags, tag)
}

func buildReason(species string, isLegend bool, metaPts float64) string {
	if isLegend {
		switch metaPts {
		case 20:
			return species + " — Legendary, Meta S"
		case 14:
			return species + " — Legendary, Meta A"
		default:
			return species + " — Legendary"
		}
	}
	switch metaPts {
	case 20:
		return species + " — Meta S"
	case 14:
		return species + " — Meta A"
	default:
		return species
	}
}
