package gamemaster

import (
	"fmt"
	"strings"

	"github.com/adzik/pokemon/model"
)

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
