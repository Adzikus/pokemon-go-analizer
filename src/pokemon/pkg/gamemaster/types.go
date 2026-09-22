package gamemaster

import "github.com/adzik/pokemon/model"

// GamemasterData represents the gamemaster.json structure (subset).
type GamemasterData struct {
	Pokemon []model.GamemasterPokemon `json:"pokemon"`
	Moves   []MoveEntry               `json:"moves"`
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

// MetaTiersByCategory holds per-context tier maps and rank maps.
type MetaTiersByCategory struct {
	Raid     map[string]string
	GL       map[string]string
	UL       map[string]string
	ML       map[string]string
	RaidRank map[string]int
	GLRank   map[string]int
	ULRank   map[string]int
	MLRank   map[string]int
}
