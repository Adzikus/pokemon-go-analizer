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
	Date       string // Kol U: Data wpisu
	Purpose    string // Kol V: Przeznaczenie
	Priority   string // Kol W: Priorytet Rozwoju
	Decision   string // Kol X: Decyzja
	CelRozwoju string // Kol Y: Cel rozwoju
	Notes      string // Kol AA: Notatki
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
	FamilyID     string    // Familie ID z gamemaster
	Tags         []string  // Dodatkowe tagi (z kol T)
	MaxCP        int       // Max CP na lvl 50 (15/15/15)

	// Per-context meta tiers (S/A+/A/B+/B/C/D)
	RaidTier string `json:"RaidTier"`
	GLTier   string `json:"GLTier"`
	ULTier   string `json:"ULTier"`
	MLTier   string `json:"MLTier"`

	// Per-context ranking positions (0-indexed, -1 if not ranked)
	RaidRank int `json:"RaidRank"`
	GLRank   int `json:"GLRank"`
	ULRank   int `json:"ULRank"`
	MLRank   int `json:"MLRank"`

	// Per-context scores (0-100)
	RaidScore float64 `json:"RaidScore"`
	GLScore   float64 `json:"GLScore"`
	ULScore   float64 `json:"ULScore"`
	MLScore   float64 `json:"MLScore"`

	// DevScore — Purpose-aware (best-of if no Purpose)
	DevScore float64 `json:"DevScore"`

	// Attack stats
	EffectiveATK float64 `json:"EffectiveATK"` // (BaseAtk + IVAtk) * CPM * ShadowBonus

	// Raid DPS (against Tier 5 boss with DEF ~190)
	RaidDPS       float64 `json:"RaidDPS"`       // DPS with current moves
	RaidDPSOptimal float64 `json:"RaidDPSOptimal"` // DPS with best possible moves

	// Move DPS (pure, no Pokemon stats)
	MoveDPSFast    float64 `json:"MoveDPSFast"`    // Fast move DPS (Power / Duration)
	MoveDPSCharged float64 `json:"MoveDPSCharged"` // Charged move DPS (Power / Duration)
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
