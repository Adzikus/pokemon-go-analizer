package analysis

import (
	"fmt"

	"github.com/adzik/pokemon/model"
	"github.com/adzik/pokemon/pkg/gamemaster"
)

type LuckyTarget struct {
	Pokemon  model.Pokemon
	Priority int
	HasLucky bool
	Reason   string
}

func BuildLuckyTargets(c []model.Pokemon, gmData *gamemaster.GamemasterData, metaData *gamemaster.MetaData) []LuckyTarget {
	luckySpecies := map[string]bool{}
	for _, p := range c {
		if p.IsLucky {
			luckySpecies[p.SpeciesID] = true
		}
	}

	metaS := map[string]bool{}
	metaA := map[string]bool{}
	for i, sid := range metaData.Rankings.Raid {
		if i < 10 {
			metaS[sid] = true
		} else if i < 20 {
			metaA[sid] = true
		}
	}

	var targets []LuckyTarget

	for _, gm := range gmData.Pokemon {
		isLegend := false
		for _, tag := range gm.Tags {
			if tag == "legendary" || tag == "mythical" {
				isLegend = true
				break
			}
		}
		if !isLegend {
			continue
		}
		if !metaS[gm.SpeciesID] {
			continue
		}
		reason := fmt.Sprintf("%s -- Meta S raid", gm.SpeciesName)
		targets = append(targets, LuckyTarget{
			Priority: 1,
			HasLucky: luckySpecies[gm.SpeciesID],
			Reason:   reason,
		})
	}

	for _, gm := range gmData.Pokemon {
		isShadowElig := false
		for _, tag := range gm.Tags {
			if tag == "shadoweligible" {
				isShadowElig = true
				break
			}
		}
		if !isShadowElig {
			continue
		}
		if !metaS[gm.SpeciesID] && !metaA[gm.SpeciesID] {
			continue
		}
		tier := "A"
		if metaS[gm.SpeciesID] {
			tier = "S"
		}
		reason := fmt.Sprintf("%s — Shadow-eligible, Meta %s", gm.SpeciesName, tier)
		targets = append(targets, LuckyTarget{
			Priority: 2,
			HasLucky: luckySpecies[gm.SpeciesID],
			Reason:   reason,
		})
	}

	return targets
}
