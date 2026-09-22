package gamemaster

import "github.com/adzik/pokemon/model"

func BuildGamemasterMap(gm *GamemasterData) map[string]*model.GamemasterPokemon {
	m := make(map[string]*model.GamemasterPokemon, len(gm.Pokemon))
	for i := range gm.Pokemon {
		m[gm.Pokemon[i].SpeciesID] = &gm.Pokemon[i]
	}
	return m
}

func BuildMoveMap(gm *GamemasterData) map[string]*MoveEntry {
	m := make(map[string]*MoveEntry, len(gm.Moves))
	for i := range gm.Moves {
		m[gm.Moves[i].MoveID] = &gm.Moves[i]
	}
	return m
}
