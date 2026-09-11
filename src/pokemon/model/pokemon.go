package model

// PokemonRaw represents raw data from ODS "Kolekcja" sheet.
type PokemonRaw struct {
	UID      string // Kol A: P-XXXXX
	Species  string // Kol B: Nazwa gatunku
	Nick     string // Kol C: Przydzielona nazwa
	Type1    string // Kol D: Typ 1
	Type2    string // Kol E: Typ 2
	CP       int    // Kol F: Combat Power
	IVAtk    int    // Kol G: IV Attack
	IVDef    int    // Kol H: IV Defense
	IVHP     int    // Kol I: IV HP
	FastMove string // Kol K: Szybki ruch
	Charged1 string // Kol L: Charged 1
	Charged2 string // Kol M: Charged 2
	Date     string // Kol U: Data wpisu
	Purpose  string // Kol V: Przeznaczenie
	Decision string // Kol X: Decyzja
	Notes    string // Kol AA: Notatki
}

// Pokemon is enriched PokemonRaw with computed fields.
type Pokemon struct {
	PokemonRaw

	// Enriched fields
	IVPct        float64   // Obliczony IV%
	IsLucky      bool      // Lucky z tags
	IsShiny      bool      // Shiny z tags
	IsShadow     bool      // Shadow z tags
	IsPurified   bool      // Purified z tags
	IsDynamax    bool      // Dynamax z tags
	IsTraded     bool      // Traded z tags
	IsHundo      bool      // IV% == 100.0
	SpeciesID    string    // ID z gamemaster (np. "mantine")
	BaseStats    BaseStats // atk, def, hp z gamemaster
	Types        []string  // Typy z gamemaster
	FastMoves    []string  // Dostępne szybkie ruchy
	ChargedMoves []string  // Dostępne charged ruchy
	MetaTier     string    // S/A/B/C
	FamilyID     string    // Familie ID z gamemaster
	Tags         []string  // Dodatkowe tagi (z kol T)
	DevScore     float64   // Score rozwoju 0-100
}

// BaseStats represents Pokemon base statistics.
type BaseStats struct {
	Atk int `json:"atk"`
	Def int `json:"def"`
	HP  int `json:"hp"`
}

// IsHundo checks if IV% is 100.
func (p *Pokemon) CheckHundo() {
	p.IsHundo = p.IVPct == 100.0
}
