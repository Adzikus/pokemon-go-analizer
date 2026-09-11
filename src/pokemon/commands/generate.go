package commands

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/adzik/pokemon/model"
	"github.com/adzik/pokemon/ods"
)

func RunGenerate(args []string, projectRoot string) {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	konto := fs.String("konto", "", "Nazwa konta (np. Test)")
	fs.Parse(args)

	if *konto == "" {
		fmt.Fprintf(os.Stderr, "Uzycie: pokemon generate --konto <nazwa>\n")
		os.Exit(1)
	}

	gmPath := filepath.Join(projectRoot, "data", "gamemaster.json")

	gmData, err := LoadGamemaster(gmPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: Gamemaster: %v\n", err)
		os.Exit(1)
	}
	gmMap := BuildGamemasterMap(gmData)

	// Prepare species entries
	entries := prepareEntries(gmData, gmMap)

	// Prepare move lookup maps
	fastMap, chargedMap := buildMoveMaps(gmData)

	// Build sheets
	kolekcjaCols := buildKolekcjaColumns()
	pokemonSheet := buildPokemonSheet(entries)
	ruchySheet := buildRuchySheet(gmData)
	atakiSheet := buildAtakiSheet(fastMap, chargedMap, entries)
	howtoSheet := buildHowtoSheet()
	listySheet := buildListySheet()
	pokedexSheet := buildPokedexSheet(entries)

	// Validations for Kolekcja
	kolekcjaValidations := buildKolekcjaValidations(len(entries), len(atakiSheet.Rows))

	// Kolekcja sheet (empty rows with formulas/validations)
	kolekcjaSheet := ods.SheetDef{
		Name:         "Kolekcja",
		Columns:      kolekcjaCols,
		Header:       kolekcjaHeader(),
		Rows:         buildKolekcjaRows(len(entries), len(atakiSheet.Rows)),
		Validations:  kolekcjaValidations,
	}

	// Pokedex validations
	pokedexValidations := []ods.Validation{
		{
			Name:        "valPokedexTick",
			Condition:   `of:cell-content-is-in-list($'Listy'.$A$2:$A$2)`,
			BaseCell:    "Pokedex.F2",
			DisplayList: "unsorted",
			AllowEmpty:  true,
			HelpMsg:     "Odhacz X jesli Reczyt juz zlapany.",
		},
	}
	pokedexSheet.Validations = pokedexValidations

	// Freeze config
	freezeConfig := map[string]ods.FreezePane{
		"Kolekcja": {Rows: 1, Columns: 2},
		"Pokemon":  {Rows: 1, Columns: 0},
		"Ruchy":    {Rows: 1, Columns: 0},
		"Ataki":    {Rows: 1, Columns: 0},
		"Pokedex":  {Rows: 1, Columns: 2},
	}

	writer := &ods.ODSWriter{
		Sheets: []ods.SheetDef{
			kolekcjaSheet,
			pokemonSheet,
			ruchySheet,
			atakiSheet,
			howtoSheet,
			listySheet,
			pokedexSheet,
		},
		FreezeConfig: freezeConfig,
	}

	outPath := filepath.Join(projectRoot, fmt.Sprintf("Pokemon-GO-%s.ods", *konto))
	if err := ods.WriteODS(outPath, writer); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: Write ODS: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("SUCCESS: Wygenerowano %s (%d gatunkow, %d ruchow)\n", outPath, len(entries), len(gmData.Moves))
}

// --- Data Preparation ---

type speciesEntry struct {
	Dex        int
	Species    string // normalized name for ODS (VLOOKUP compatible)
	SpeciesID  string // gamemaster speciesId
	Type1      string
	Type2      string
	BaseAtk    int
	BaseDef    int
	BaseHP     int
	FastMoves  []string
	ChargedMoves []string
	IsShadow   bool
}

func prepareEntries(gm *GamemasterData, gmMap map[string]*model.GamemasterPokemon) []speciesEntry {
	forceReleased := map[string]bool{
		"ditto": true, "spewpa": true, "arrokuda": true, "cramorant": true,
	}

	var entries []speciesEntry
	for _, p := range gm.Pokemon {
		if !p.Released && !forceReleased[p.SpeciesID] {
			continue
		}

		species := normalizeName(p.SpeciesName)

		type1 := ""
		type2 := ""
		if len(p.Types) > 0 {
			type1 = capitalize(p.Types[0])
		}
		if len(p.Types) > 1 {
			type2 = capitalize(p.Types[1])
		}

		// Resolve move names
		var fastMoves []string
		for _, id := range p.FastMoves {
			if name, ok := moveLookup[id]; ok {
				fastMoves = append(fastMoves, name)
			} else {
				fastMoves = append(fastMoves, id)
			}
		}
		var chargedMoves []string
		for _, id := range p.ChargedMoves {
			if name, ok := moveLookup[id]; ok {
				chargedMoves = append(chargedMoves, name)
			} else {
				chargedMoves = append(chargedMoves, id)
			}
		}

		entries = append(entries, speciesEntry{
			Dex:          p.Dex,
			Species:      species,
			SpeciesID:    p.SpeciesID,
			Type1:        type1,
			Type2:        type2,
			BaseAtk:      p.BaseStats.Atk,
			BaseDef:      p.BaseStats.Def,
			BaseHP:       p.BaseStats.HP,
			FastMoves:    fastMoves,
			ChargedMoves: chargedMoves,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Dex != entries[j].Dex {
			return entries[i].Dex < entries[j].Dex
		}
		return entries[i].Species < entries[j].Species
	})

	return entries
}

func normalizeName(name string) string {
	// "Alolan Vulpix" -> "Vulpix - Alolan"
	// "Charizard Mega Y" -> "Charizard - Mega Y"
	parts := strings.SplitN(name, " ", 2)
	if len(parts) == 2 && strings.Contains(parts[1], "Mega") {
		return parts[0] + " - " + parts[1]
	}
	// Handle regional forms
	if strings.HasPrefix(name, "Alolan ") {
		return strings.TrimPrefix(name, "Alolan ") + " - Alolan"
	}
	if strings.HasPrefix(name, "Galarian ") {
		return strings.TrimPrefix(name, "Galarian ") + " - Galarian"
	}
	if strings.HasPrefix(name, "Hisuian ") {
		return strings.TrimPrefix(name, "Hisuian ") + " - Hisuian"
	}
	if strings.HasPrefix(name, "Paldean ") {
		return strings.TrimPrefix(name, "Paldean ") + " - Paldean"
	}
	return name
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// Move lookup (populated at init)
var moveLookup map[string]string

func buildMoveMaps(gm *GamemasterData) (fastMap map[string][]string, chargedMap map[string][]string) {
	moveLookup = make(map[string]string)
	fastMap = make(map[string][]string)
	chargedMap = make(map[string][]string)

	for _, m := range gm.Moves {
		moveLookup[m.MoveID] = m.Name
	}

	for _, p := range gm.Pokemon {
		for _, id := range p.FastMoves {
			if name, ok := moveLookup[id]; ok {
				fastMap[p.SpeciesID] = append(fastMap[p.SpeciesID], name)
			}
		}
		for _, id := range p.ChargedMoves {
			if name, ok := moveLookup[id]; ok {
				chargedMap[p.SpeciesID] = append(chargedMap[p.SpeciesID], name)
			}
		}
	}

	return fastMap, chargedMap
}

// --- Kolekcja Sheet ---

func buildKolekcjaColumns() []ods.ColDef {
	return []ods.ColDef{
		{Width: 2.2}, // A: UID
		{Width: 3.6}, // B: Gatunek
		{Width: 3.2}, // C: Wlasna nazwa
		{Width: 2.4}, // D: Typ 1
		{Width: 2.4}, // E: Typ 2
		{Width: 1.9}, // F: CP
		{Width: 1.4}, // G: IV Atk
		{Width: 1.4}, // H: IV Def
		{Width: 1.4}, // I: IV HP
		{Width: 1.9}, // J: IV %
		{Width: 3.2}, // K: Fast Attack
		{Width: 3.4}, // L: Charged 1
		{Width: 3.4}, // M: Charged 2
		{Width: 1.7}, // N: Shiny
		{Width: 1.7}, // O: Lucky
		{Width: 1.7}, // P: Shadow
		{Width: 1.7}, // Q: Purified
		{Width: 2.8}, // R: Dynamax
		{Width: 1.7}, // S: Traded
		{Width: 3.6}, // T: Inne cechy
		{Width: 2.8}, // U: Data
		{Width: 2.8}, // V: Przeznaczenie
		{Width: 3.0}, // W: Priorytet
		{Width: 3.0}, // X: Decyzja
		{Width: 3.5}, // Y: Cel rozwoju
		{Width: 5.0}, // Z: Powod
		{Width: 4.5}, // AA: Uwagi
	}
}

func kolekcjaHeader() []string {
	return []string{
		"UID", "Gatunek", "Wlasna nazwa", "Typ 1", "Typ 2",
		"CP", "IV Atk", "IV Def", "IV HP", "IV %",
		"Fast Attack", "Charged Attack 1", "Charged Attack 2",
		"Shiny", "Lucky", "Shadow", "Purified", "Dynamax", "Traded",
		"Inne cechy", "Data zlapania", "Przeznaczenie", "Priorytet Rozwoju",
		"Decyzja", "Cel rozwoju", "Powod decyzji", "Uwagi",
	}
}

func buildKolekcjaRows(nSpecies, nMoves int) [][]ods.Cell {
	var rows [][]ods.Cell
	lastRow := 1001 // 1000 data rows (rows 2-1001)

	for r := 2; r <= lastRow; r++ {
		row := make([]ods.Cell, 27)
		// A: UID (empty)
		row[0] = ods.Cell{Type: "empty"}
		// B: Gatunek (with validation)
		row[1] = ods.Cell{Type: "empty", Validation: "valGat"}
		// C: Wlasna nazwa (empty)
		row[2] = ods.Cell{Type: "empty"}
		// D: Typ 1 (formula)
		dFormula := fmt.Sprintf(`IF([.B%d]="";"";VLOOKUP([.B%d];$'Pokemon'.$B$2:$D$%d;2;0))`, r, r, nSpecies+1)
		row[3] = ods.Cell{Type: "formula", Formula: dFormula, Style: "ce_ro"}
		// E: Typ 2 (formula)
		eFormula := fmt.Sprintf(`IF([.B%d]="";"";VLOOKUP([.B%d];$'Pokemon'.$B$2:$D$%d;3;0))`, r, r, nSpecies+1)
		row[4] = ods.Cell{Type: "formula", Formula: eFormula, Style: "ce_ro"}
		// F: CP (empty)
		row[5] = ods.Cell{Type: "empty"}
		// G-I: IV (empty)
		row[6] = ods.Cell{Type: "empty"}
		row[7] = ods.Cell{Type: "empty"}
		row[8] = ods.Cell{Type: "empty"}
		// J: IV % (formula)
		jFormula := fmt.Sprintf(`IF(COUNT([.G%d]:[.I%d])=3;([.G%d]+[.H%d]+[.I%d])/45*100;"")`, r, r, r, r, r)
		row[9] = ods.Cell{Type: "formula", Formula: jFormula, Style: "ce_iv"}
		// K: Fast Attack (with dynamic validation)
		kVal := fmt.Sprintf("valF%d", r)
		row[10] = ods.Cell{Type: "empty", Validation: kVal}
		// L: Charged 1 (with dynamic validation)
		lVal := fmt.Sprintf("valC1%d", r)
		row[11] = ods.Cell{Type: "empty", Validation: lVal}
		// M: Charged 2 (with dynamic validation)
		mVal := fmt.Sprintf("valC2%d", r)
		row[12] = ods.Cell{Type: "empty", Validation: mVal}
		// N-S: Tick columns (with valX)
		for i := 13; i <= 18; i++ {
			row[i] = ods.Cell{Type: "empty", Style: "ce_center", Validation: "valX"}
		}
		// T: Inne cechy (empty)
		row[19] = ods.Cell{Type: "empty"}
		// U: Data (with date validation)
		row[20] = ods.Cell{Type: "empty", Style: "ce_date", Validation: "valData"}
		// V: Przeznaczenie (with validation)
		row[21] = ods.Cell{Type: "empty", Validation: "valPrze"}
		// W: Priorytet (with validation)
		row[22] = ods.Cell{Type: "empty", Validation: "valPrio"}
		// X: Decyzja (with validation)
		row[23] = ods.Cell{Type: "empty", Validation: "valDec"}
		// Y-AA: Freeform (empty)
		row[24] = ods.Cell{Type: "empty"}
		row[25] = ods.Cell{Type: "empty"}
		row[26] = ods.Cell{Type: "empty"}

		rows = append(rows, row)
	}
	return rows
}

func buildKolekcjaValidations(nSpecies, nMoves int) []ods.Validation {
	var vals []ods.Validation

	// Static validations
	vals = append(vals,
		ods.Validation{
			Name: "valGat", Condition: `of:cell-content-is-in-list($'Pokemon'.$B$2:$B$` + fmt.Sprintf("%d", nSpecies+1) + `)`,
			BaseCell: "Kolekcja.B2", DisplayList: "unsorted", AllowEmpty: true,
			ErrorMsg: "Wybierz wartosc z listy rozwijanej.",
			HelpMsg:  "Wpisz poczatek nazwy lub wybierz z listy.",
		},
		ods.Validation{
			Name: "valX", Condition: `of:cell-content-is-in-list($'Listy'.$A$2:$A$2)`,
			BaseCell: "Kolekcja.N2", DisplayList: "unsorted", AllowEmpty: true,
			HelpMsg: "Odhacz X jesli cecha dotyczy tego pokemona.",
		},
		ods.Validation{
			Name: "valPrze", Condition: `of:cell-content-is-in-list($'Listy'.$B$3:$B$8)`,
			BaseCell: "Kolekcja.V2", DisplayList: "unsorted", AllowEmpty: true,
			HelpMsg: "Wybierz przeznaczenie.",
		},
		ods.Validation{
			Name: "valPrio", Condition: `of:cell-content-is-in-list($'Listy'.$C$9:$C$13)`,
			BaseCell: "Kolekcja.W2", DisplayList: "unsorted", AllowEmpty: true,
			HelpMsg: "Priorytet rozwoju 1-5.",
		},
		ods.Validation{
			Name: "valDec", Condition: `of:cell-content-is-in-list($'Listy'.$D$14:$D$18)`,
			BaseCell: "Kolekcja.X2", DisplayList: "unsorted", AllowEmpty: true,
			HelpMsg: "Wybierz decyzje.",
		},
		ods.Validation{
			Name: "valData", Condition: `of:cell-content-is-date()`,
			BaseCell: "Kolekcja.U2", DisplayList: "none", AllowEmpty: true,
			HelpMsg: "Data w formacie RRRR-MM-DD.",
		},
	)

	// Dynamic move validations (per row)
	lastRow := 1001
	for r := 2; r <= lastRow; r++ {
		fastCond := fmt.Sprintf(`of:cell-content-is-in-list(IF(COUNTIF($Ataki.$A$2:$A$%d;[.B%d]);OFFSET($Ataki.$B$2;MATCH([.B%d];$Ataki.$A$2:$A$%d;0)-1;0;COUNTIF($Ataki.$A$2:$A$%d;[.B%d]);1);""))`,
			nMoves+1, r, r, nMoves+1, nMoves+1, r)
		vals = append(vals, ods.Validation{
			Name: fmt.Sprintf("valF%d", r), Condition: fastCond,
			BaseCell: fmt.Sprintf("Kolekcja.K%d", r), DisplayList: "unsorted", AllowEmpty: true,
			HelpMsg: "Atak szybki - lista zawazona do ruchow gatunku z kolumny B.",
		})

		chargedCond := fmt.Sprintf(`of:cell-content-is-in-list(IF(COUNTIF($Ataki.$C$2:$C$%d;[.B%d]);OFFSET($Ataki.$D$2;MATCH([.B%d];$Ataki.$C$2:$C$%d;0)-1;0;COUNTIF($Ataki.$C$2:$C$%d;[.B%d]);1);""))`,
			nMoves+1, r, r, nMoves+1, nMoves+1, r)
		vals = append(vals, ods.Validation{
			Name: fmt.Sprintf("valC1%d", r), Condition: chargedCond,
			BaseCell: fmt.Sprintf("Kolekcja.L%d", r), DisplayList: "unsorted", AllowEmpty: true,
			HelpMsg: "Atak/specjalny 1 - lista zawazona do ruchow gatunku z kolumny B.",
		})
		vals = append(vals, ods.Validation{
			Name: fmt.Sprintf("valC2%d", r), Condition: chargedCond,
			BaseCell: fmt.Sprintf("Kolekcja.M%d", r), DisplayList: "unsorted", AllowEmpty: true,
			HelpMsg: "Atak/specjalny 2 - lista zawazona do ruchow gatunku z kolumny B.",
		})
	}

	return vals
}

// --- Pokemon Sheet ---

func buildPokemonSheet(entries []speciesEntry) ods.SheetDef {
	cols := []ods.ColDef{
		{Width: 1.2}, // A: Dex
		{Width: 3.6}, // B: Gatunek
		{Width: 2.4}, // C: Typ 1
		{Width: 2.4}, // D: Typ 2
		{Width: 1.4}, // E: Atk
		{Width: 1.4}, // F: Def
		{Width: 1.4}, // G: HP
		{Width: 1.9}, // H: Max CP
		{Width: 5.0}, // I: Fast Moves
		{Width: 6.0}, // J: Charged Moves
		{Width: 2.0}, // K: Stars
		{Width: 2.0}, // L: Buddy
	}
	header := []string{"Dex", "Gatunek", "Typ 1", "Typ 2", "Atk", "Def", "HP", "Max CP", "Fast Moves", "Charged Moves", "Stars", "Buddy"}

	var rows [][]ods.Cell
	for _, e := range entries {
		maxCP := calcMaxCP(e.BaseAtk, e.BaseDef, e.BaseHP)
		row := []ods.Cell{
			{Type: "float", FloatVal: float64(e.Dex), Style: "ce_center"},
			{Type: "text", Value: e.Species},
			{Type: "text", Value: e.Type1, Style: "ce_center"},
			{Type: "text", Value: e.Type2, Style: "ce_center"},
			{Type: "float", FloatVal: float64(e.BaseAtk), Style: "ce_center"},
			{Type: "float", FloatVal: float64(e.BaseDef), Style: "ce_center"},
			{Type: "float", FloatVal: float64(e.BaseHP), Style: "ce_center"},
			{Type: "float", FloatVal: float64(maxCP), Style: "ce_center"},
			{Type: "text", Value: strings.Join(e.FastMoves, ", ")},
			{Type: "text", Value: strings.Join(e.ChargedMoves, ", ")},
			{Type: "text", Value: "", Style: "ce_center"},
			{Type: "text", Value: "", Style: "ce_center"},
		}
		rows = append(rows, row)
	}

	return ods.SheetDef{Name: "Pokemon", Columns: cols, Header: header, Rows: rows}
}

func calcMaxCP(atk, def, hp int) int {
	// CP = floor((Atk * sqrt(Def) * sqrt(HP) * CPM^2) / 10)
	// At L50 (CPM ~0.8403)
	a := float64(atk + 15)
	d := float64(def + 15)
	h := float64(hp + 15)
	cpm := 0.8403000003293185
	cp := (a * sqrt(d) * sqrt(h) * cpm * cpm) / 10.0
	return int(cp)
}

func sqrt(x float64) float64 {
	// Newton's method
	if x <= 0 {
		return 0
	}
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}

// --- Ruchy Sheet ---

func buildRuchySheet(gm *GamemasterData) ods.SheetDef {
	cols := []ods.ColDef{
		{Width: 3.6}, // A: Name
		{Width: 2.4}, // B: Type
		{Width: 2.0}, // C: Kind
		{Width: 1.4}, // D: Power
		{Width: 1.4}, // E: Energy
		{Width: 2.0}, // F: Cooldown
		{Width: 1.4}, // G: Turns
		{Width: 1.4}, // H: DPT
	}
	header := []string{"Name", "Type", "Kind", "Power", "Energy", "Cooldown", "Turns", "DPT"}

	var rows [][]ods.Cell
	for _, m := range gm.Moves {
		kind := "Fast"
		if m.EnergyGain == 0 {
			kind = "Charged"
		}
		dpt := 0.0
		if m.Cooldown > 0 {
			dpt = float64(m.Power) / float64(m.Cooldown) * 1000
		}
		row := []ods.Cell{
			{Type: "text", Value: m.Name},
			{Type: "text", Value: capitalize(m.Type), Style: "ce_center"},
			{Type: "text", Value: kind, Style: "ce_center"},
			{Type: "float", FloatVal: float64(m.Power), Style: "ce_center"},
			{Type: "float", FloatVal: float64(m.Energy), Style: "ce_center"},
			{Type: "float", FloatVal: float64(m.Cooldown), Style: "ce_center"},
			{Type: "float", FloatVal: float64(m.Turns), Style: "ce_center"},
			{Type: "float", FloatVal: dpt, Style: "ce_center"},
		}
		rows = append(rows, row)
	}

	return ods.SheetDef{Name: "Ruchy", Columns: cols, Header: header, Rows: rows}
}

// --- Ataki Sheet (hidden) ---

func buildAtakiSheet(fastMap, chargedMap map[string][]string, entries []speciesEntry) ods.SheetDef {
	cols := []ods.ColDef{
		{Width: 3.6}, // A: Species (fast)
		{Width: 3.2}, // B: Fast Move
		{Width: 3.6}, // C: Species (charged)
		{Width: 3.4}, // D: Charged Move
	}
	header := []string{"Species", "Fast Move", "Species", "Charged Move"}

	// Build vertical lookup tables
	type atakiRow struct {
		speciesA string
		fastA    string
		speciesC string
		chargedC string
	}

	var rowsData []atakiRow

	// Fast moves: each species repeated per fast move
	for _, e := range entries {
		moves := fastMap[e.SpeciesID]
		for i, move := range moves {
			species := ""
			if i == 0 {
				species = e.Species
			}
			rowsData = append(rowsData, atakiRow{speciesA: species, fastA: move})
		}
	}

	// Charged moves: each species repeated per charged move
	cIdx := 0
	for _, e := range entries {
		moves := chargedMap[e.SpeciesID]
		for i, move := range moves {
			species := ""
			if i == 0 {
				species = e.Species
			}
			if cIdx < len(rowsData) {
				rowsData[cIdx].speciesC = species
				rowsData[cIdx].chargedC = move
			} else {
				rowsData = append(rowsData, atakiRow{speciesC: species, chargedC: move})
			}
			cIdx++
		}
	}

	var rows [][]ods.Cell
	for _, r := range rowsData {
		sA := r.speciesA
		sC := r.speciesC
		if sA == "" && sC == "" {
			continue
		}
		row := []ods.Cell{
			{Type: "text", Value: sA},
			{Type: "text", Value: r.fastA},
			{Type: "text", Value: sC},
			{Type: "text", Value: r.chargedC},
		}
		rows = append(rows, row)
	}

	return ods.SheetDef{Name: "Ataki", Hidden: true, Columns: cols, Header: header, Rows: rows}
}

// --- Listy Sheet (hidden) ---

func buildListySheet() ods.SheetDef {
	cols := []ods.ColDef{
		{Width: 2.0}, // A: X
		{Width: 3.0}, // B: Przeznaczenie
		{Width: 2.0}, // C: Priorytet
		{Width: 3.0}, // D: Decyzja
	}
	header := []string{"X", "Przeznaczenie", "Priorytet", "Decyzja"}

	purposes := []string{"PvP", "Rajdy", "Gym", "Trade", "Transfer", "Kolekcja"}
	priorities := []string{"1", "2", "3", "4", "5"}
	decisions := []string{"ROZWIN", "PRETENDENT", "TRADE", "PROFESOR", "ZACHOWAJ"}

	maxLen := len(purposes)
	if len(priorities) > maxLen {
		maxLen = len(priorities)
	}
	if len(decisions) > maxLen {
		maxLen = len(decisions)
	}

	var rows [][]ods.Cell
	// Row 1: "X" only
	rows = append(rows, []ods.Cell{
		{Type: "text", Value: "X"},
		{Type: "empty"},
		{Type: "empty"},
		{Type: "empty"},
	})
	// Rows 2+: purposes, priorities, decisions
	for i := 0; i < maxLen; i++ {
		row := []ods.Cell{
			{Type: "empty"},
			{Type: "empty"},
			{Type: "empty"},
			{Type: "empty"},
		}
		if i < len(purposes) {
			row[1] = ods.Cell{Type: "text", Value: purposes[i]}
		}
		if i < len(priorities) {
			row[2] = ods.Cell{Type: "text", Value: priorities[i]}
		}
		if i < len(decisions) {
			row[3] = ods.Cell{Type: "text", Value: decisions[i]}
		}
		rows = append(rows, row)
	}

	return ods.SheetDef{Name: "Listy", Hidden: true, Columns: cols, Header: header, Rows: rows}
}

// --- Jak Uzywac Sheet ---

func buildHowtoSheet() ods.SheetDef {
	cols := []ods.ColDef{
		{Width: 26.0},
	}
	header := []string{"Jak uzyc - Pokemon GO Collection Tracker"}

	lines := []struct{ style, text string }{
		{"howto_h1", "Pokemon GO Collection Tracker"},
		{"howto_text", ""},
		{"howto_h2", "1. Kolekcja - Twoja kolekcja Pokemon"},
		{"howto_text", "Kazdy wiersz = jeden Pokemon. Wpisz dane recznie lub skopiuj z gry."},
		{"howto_text", "Kolumny D, E, J sa automatycznie obliczane (NIE EDYTUJ)."},
		{"howto_text", "Kolumny N-S (Shiny, Lucky, itp.) - wpisz X jesli cecha dotyczy."},
		{"howto_text", "Kolumna V (Przeznaczenie) - wybierz z listy: PvP, Rajdy, Gym, Trade, Transfer, Kolekcja."},
		{"howto_text", ""},
		{"howto_h2", "2. Pokemon - Slownik gatunkow"},
		{"howto_text", "Wszystkie gatunki z Game Mastera. Uzywane przez VLOOKUP w kolumnach D, E."},
		{"howto_text", ""},
		{"howto_h2", "3. Ruchy - Lista wszystkich ruchow"},
		{"howto_text", "Referencyjna lista ruchow z moca, energia, czasem odnowienia."},
		{"howto_text", ""},
		{"howto_h2", "4. Ataki - Ukryty arkusz pomocniczy"},
		{"howto_text", "Listy ruchow per gatunek. Uzywane przez dynamiczne listy rozwijane w kolumnach K, L, M."},
		{"howto_text", ""},
		{"howto_h2", "5. Pokedex - Checklist"},
		{"howto_text", "Wszystkie gatunki z checkboxami. Odhacz X gdy lapiesz nowego Pokemon."},
		{"howto_text", ""},
		{"howto_h2", "6. Listy - Listy walidacji"},
		{"howto_text", "Ukryty arkusz z listami do walidacji danych. Nie edytuj."},
	}

	var rows [][]ods.Cell
	for _, l := range lines {
		rows = append(rows, []ods.Cell{
			{Type: "text", Value: l.text, Style: l.style},
		})
	}

	return ods.SheetDef{Name: "Jak używać", Columns: cols, Header: header, Rows: rows}
}

// --- Pokedex Sheet ---

func buildPokedexSheet(entries []speciesEntry) ods.SheetDef {
	cols := []ods.ColDef{
		{Width: 1.2}, // A: Dex
		{Width: 3.6}, // B: Gatunek
		{Width: 2.4}, // C: Typ 1
		{Width: 2.4}, // D: Typ 2
		{Width: 1.4}, // E: Atk
		{Width: 1.7}, // F: Caught
		{Width: 1.7}, // G: Lucky
		{Width: 1.7}, // H: Shiny
		{Width: 1.7}, // I: Hundo
		{Width: 1.7}, // J: Shadow
		{Width: 1.7}, // K: Purified
		{Width: 2.8}, // L: Dynamax
		{Width: 1.7}, // M: Mega
		{Width: 3.0}, // N: Tagi
		{Width: 5.0}, // O: Notatki
	}
	header := []string{"Dex", "Gatunek", "Typ 1", "Typ 2", "Atk", "Caught", "Lucky", "Shiny", "Hundo", "Shadow", "Purified", "Dynamax", "Mega", "Tagi", "Notatki"}

	var rows [][]ods.Cell
	for _, e := range entries {
		row := []ods.Cell{
			{Type: "float", FloatVal: float64(e.Dex), Style: "ce_center"},
			{Type: "text", Value: e.Species},
			{Type: "text", Value: e.Type1, Style: "ce_center"},
			{Type: "text", Value: e.Type2, Style: "ce_center"},
			{Type: "float", FloatVal: float64(e.BaseAtk), Style: "ce_center"},
		}
		// Checkbox columns (F-M): empty with valPokedexTick
		for i := 0; i < 8; i++ {
			row = append(row, ods.Cell{Type: "empty", Style: "ce_center", Validation: "valPokedexTick"})
		}
		// N-O: freeform
		row = append(row, ods.Cell{Type: "empty"}, ods.Cell{Type: "empty"})
		rows = append(rows, row)
	}

	return ods.SheetDef{Name: "Pokedex", Columns: cols, Header: header, Rows: rows}
}
