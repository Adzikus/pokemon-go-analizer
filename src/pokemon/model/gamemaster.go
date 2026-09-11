package model

// GamemasterData represents the full gamemaster.json structure.
type GamemasterData struct {
	Timestamp    string              `json:"timestamp"`
	ID           string              `json:"id"`
	Title        string              `json:"title"`
	Settings     interface{}         `json:"settings"`
	Pokemon      []GamemasterPokemon `json:"pokemon"`
	FastMoves    []FastMove          `json:"fastMoves"`
	ChargedMoves []ChargedMove       `json:"chargedMoves"`
	Formats      []Format            `json:"formats"`
	PokemonTags  []string            `json:"pokemonTags"`
}

// GamemasterPokemon represents a single Pokemon entry in gamemaster.
type GamemasterPokemon struct {
	Dex            int                   `json:"dex"`
	SpeciesName    string                `json:"speciesName"`
	SpeciesID      string                `json:"speciesId"`
	BaseStats      BaseStats             `json:"baseStats"`
	Types          []string              `json:"types"`
	FastMoves      []string              `json:"fastMoves"`
	ChargedMoves   []string              `json:"chargedMoves"`
	Tags           []string              `json:"tags"`
	DefaultIVs     map[string][4]float64 `json:"defaultIVs"`
	Level25CP      int                   `json:"level25CP"`
	BuddyDistance  int                   `json:"buddyDistance"`
	ThirdMoveCost  interface{}           `json:"thirdMoveCost"` // can be int or bool
	Released       bool                  `json:"released"`
	Family         *Family               `json:"family,omitempty"`
	EliteMoves     []string              `json:"eliteMoves,omitempty"`
	SearchPriority *int                  `json:"searchPriority,omitempty"`
}

// Family represents Pokemon evolution family.
type Family struct {
	ID         string   `json:"id"`
	Evolutions []string `json:"evolutions,omitempty"`
	Parent     string   `json:"parent,omitempty"`
}

// FastMove represents a fast move entry.
type FastMove struct {
	MoveID       string `json:"moveId"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	Power        int    `json:"power"`
	Duration     int    `json:"duration"`
	Energy       int    `json:"energy"`
	DamageWindow int    `json:"damageWindow"`
}

// ChargedMove represents a charged move entry.
type ChargedMove struct {
	MoveID         string  `json:"moveId"`
	Name           string  `json:"name"`
	Type           string  `json:"type"`
	Power          int     `json:"power"`
	Duration       int     `json:"duration"`
	Energy         int     `json:"energy"`
	CriticalChance float64 `json:"criticalChance"`
}

// Format represents a league/cup format.
type Format struct {
	ID   string   `json:"id"`
	Name string   `json:"name"`
	CP   int      `json:"cp,omitempty"`
	Meta string   `json:"meta,omitempty"`
	Tags []string `json:"tags,omitempty"`
}
