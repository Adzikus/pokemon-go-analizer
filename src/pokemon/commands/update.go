package commands

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/adzik/pokemon/ods"
)

func RunUpdate(args []string, projectRoot string) {
	fs := flag.NewFlagSet("update", flag.ExitOnError)
	konto := fs.String("konto", "", "Nazwa konta (np. Test)")
	odsPath := fs.String("ods", "", "Sciezka do ODS (opcjonalnie)")
	fs.Parse(args)

	path := *odsPath
	if path == "" && *konto != "" {
		path = filepath.Join(projectRoot, fmt.Sprintf("Pokemon-GO-%s.ods", *konto))
	}
	if path == "" {
		fmt.Fprintf(os.Stderr, "Uzycie: pokemon update --konto <nazwa> [--ods <sciezka>]\n")
		os.Exit(1)
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(projectRoot, path)
	}

	// 1. Verify source ODS exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "FAIL: ODS nie istnieje: %s\n", path)
		os.Exit(1)
	}

	// 2. Backup
	fmt.Println("1/4 Backup...")
	backupDir := filepath.Join(projectRoot, "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: Backup dir: %v\n", err)
		os.Exit(1)
	}
	timestamp := time.Now().Format("20060102_150405")
	backupPath := filepath.Join(backupDir, fmt.Sprintf("%s.bak.%s", filepath.Base(path), timestamp))

	srcData, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: Read ODS: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(backupPath, srcData, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: Write backup: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  Backup: %s\n", backupPath)

	// 3. Read collection data from old ODS
	fmt.Println("2/4 Czytam Kolekcje...")
	oldDoc, err := ods.Read(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: Read old ODS: %v\n", err)
		os.Exit(1)
	}
	oldSheet := oldDoc.GetSheet("Kolekcja")
	if oldSheet == nil {
		fmt.Fprintf(os.Stderr, "FAIL: Brak arkusza Kolekcja w %s\n", path)
		os.Exit(1)
	}
	collectionRows := extractCollectionRows(oldSheet)
	fmt.Printf("  Wczytano %d wierszy Kolekcji\n", len(collectionRows))

	// 4. Extract konto name
	kontoName := *konto
	if kontoName == "" {
		kontoName = extractKontoFromFilename(filepath.Base(path))
	}
	if kontoName == "" {
		kontoName = "unknown"
	}

	// 5. Generate new ODS with collection data
	fmt.Println("3/4 Generuje nowy ODS...")
	newPath := filepath.Join(projectRoot, fmt.Sprintf("Pokemon-GO-%s.ods", kontoName))
	if err := injectCollection(newPath, collectionRows); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: Inject collection: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("SUCCESS: Zaktualizowano %s (%d wierszy Kolekcji)\n", newPath, len(collectionRows))
}

// extractCollectionRows extracts all data rows from Kolekcja sheet.
// Returns raw cell values (strings) preserving column positions.
func extractCollectionRows(sheet *ods.Sheet) [][]string {
	var rows [][]string
	// Skip header row (index 0)
	for i := 1; i < len(sheet.Rows); i++ {
		row := sheet.Rows[i]
		// Skip empty rows (all cells empty)
		allEmpty := true
		for _, cell := range row {
			if cell != "" {
				allEmpty = false
				break
			}
		}
		if allEmpty {
			continue
		}
		rows = append(rows, row)
	}
	return rows
}

func injectCollection(path string, collectionRows [][]string) error {
	// We need to re-generate the ODS with the user's collection data
	// by replacing the empty Kolekcja rows with the user's data

	// Determine project root from ODS path
	projectRoot := filepath.Dir(path) // e.g., /home/adzik/Repozytoria/Pokemon-GO
	gmPath := filepath.Join(projectRoot, "data", "gamemaster.json")

	gmData, err := LoadGamemaster(gmPath)
	if err != nil {
		return fmt.Errorf("load gamemaster: %w", err)
	}
	gmMap := BuildGamemasterMap(gmData)

	fastMap, chargedMap, moveLookup := buildMoveMaps(gmData)
	entries := prepareEntries(gmData, gmMap, moveLookup)

	pokemonSheet := buildPokemonSheet(entries)
	ruchySheet := buildRuchySheet(gmData)
	atakiSheet := buildAtakiSheet(fastMap, chargedMap, entries)
	howtoSheet := buildHowtoSheet()
	listySheet := buildListySheet()
	pokedexSheet := buildPokedexSheet(entries)
	kolekcjaValidations := buildKolekcjaValidations(len(entries), len(atakiSheet.Rows))

	// Convert string rows to Cell rows
	nCols := 27
	var cellRows [][]ods.Cell
	for _, row := range collectionRows {
		cellRow := make([]ods.Cell, nCols)
		for i := 0; i < nCols; i++ {
			val := ""
			if i < len(row) {
				val = row[i]
			}
			cellRow[i] = ods.Cell{Type: "text", Value: val}
		}
		cellRows = append(cellRows, cellRow)
	}

	// Build Kolekcja sheet with user data + formulas
	kolekcjaCellRows := buildKolekcjaRows(len(entries), len(atakiSheet.Rows), cellRows)

	freezeConfig := map[string]ods.FreezePane{
		"Kolekcja": {Rows: 1, Columns: 2},
		"Pokemon":  {Rows: 1, Columns: 0},
		"Ruchy":    {Rows: 1, Columns: 0},
		"Ataki":    {Rows: 1, Columns: 0},
		"Pokedex":  {Rows: 1, Columns: 2},
	}

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

	writer := &ods.ODSWriter{
		Sheets: []ods.SheetDef{
			{
				Name:        "Kolekcja",
				Columns:     buildKolekcjaColumns(),
				Header:      kolekcjaHeader(),
				Rows:        kolekcjaCellRows,
				Validations: kolekcjaValidations,
			},
			pokemonSheet,
			ruchySheet,
			atakiSheet,
			howtoSheet,
			listySheet,
			pokedexSheet,
		},
		FreezeConfig: freezeConfig,
	}

	return ods.WriteODS(path, writer)
}

// buildKolekcjaRowsWithData builds Kolekcja rows with user data + formulas/validations.
func buildKolekcjaRowsWithData(userData [][]ods.Cell, nSpecies, nMoves int) [][]ods.Cell {
	var rows [][]ods.Cell
	lastRow := 1001

	for r := 2; r <= lastRow; r++ {
		rowIdx := r - 2 // 0-based index into userData

		if rowIdx < len(userData) {
			// User row - add formulas and validations
			userRow := userData[rowIdx]
			row := make([]ods.Cell, 27)

			// A: UID (from user)
			row[0] = userRow[0]
			// B: Gatunek (from user + validation)
			row[1] = userRow[1]
			row[1].Validation = "valGat"
			// C: Wlasna nazwa (from user)
			row[2] = userRow[2]
			// D: Typ 1 (formula)
			dFormula := fmt.Sprintf(`IF([.B%d]="";"";VLOOKUP([.B%d];$'Pokemon'.$B$2:$D$%d;2;0))`, r, r, nSpecies+1)
			row[3] = ods.Cell{Type: "formula", Formula: dFormula, Style: "ce_ro"}
			// E: Typ 2 (formula)
			eFormula := fmt.Sprintf(`IF([.B%d]="";"";VLOOKUP([.B%d];$'Pokemon'.$B$2:$D$%d;3;0))`, r, r, nSpecies+1)
			row[4] = ods.Cell{Type: "formula", Formula: eFormula, Style: "ce_ro"}
			// F: CP (from user)
			if len(userRow) > 5 {
				row[5] = userRow[5]
			} else {
				row[5] = ods.Cell{Type: "empty"}
			}
			// G-I: IV (from user)
			for i := 6; i <= 8; i++ {
				if len(userRow) > i {
					row[i] = userRow[i]
				} else {
					row[i] = ods.Cell{Type: "empty"}
				}
			}
			// J: IV % (formula)
			jFormula := fmt.Sprintf(`IF(COUNT([.G%d]:[.I%d])=3;([.G%d]+[.H%d]+[.I%d])/45*100;"")`, r, r, r, r, r)
			row[9] = ods.Cell{Type: "formula", Formula: jFormula, Style: "ce_iv"}
			// K-M: Moves (from user + validation)
			kVal := fmt.Sprintf("valF%d", r)
			lVal := fmt.Sprintf("valC1%d", r)
			mVal := fmt.Sprintf("valC2%d", r)
			if len(userRow) > 10 {
				row[10] = userRow[10]
			} else {
				row[10] = ods.Cell{Type: "empty"}
			}
			row[10].Validation = kVal
			if len(userRow) > 11 {
				row[11] = userRow[11]
			} else {
				row[11] = ods.Cell{Type: "empty"}
			}
			row[11].Validation = lVal
			if len(userRow) > 12 {
				row[12] = userRow[12]
			} else {
				row[12] = ods.Cell{Type: "empty"}
			}
			row[12].Validation = mVal
			// N-S: Ticks (from user + valX)
			for i := 13; i <= 18; i++ {
				if len(userRow) > i {
					row[i] = userRow[i]
				} else {
					row[i] = ods.Cell{Type: "empty"}
				}
				row[i].Style = "ce_center"
				row[i].Validation = "valX"
			}
			// T: Inne cechy (from user)
			if len(userRow) > 19 {
				row[19] = userRow[19]
			} else {
				row[19] = ods.Cell{Type: "empty"}
			}
			// U: Date (from user + validation)
			if len(userRow) > 20 {
				row[20] = userRow[20]
			} else {
				row[20] = ods.Cell{Type: "empty"}
			}
			row[20].Style = "ce_date"
			row[20].Validation = "valData"
			// V: Purpose (from user + validation)
			if len(userRow) > 21 {
				row[21] = userRow[21]
			} else {
				row[21] = ods.Cell{Type: "empty"}
			}
			row[21].Validation = "valPrze"
			// W: Priority (from user + validation)
			if len(userRow) > 22 {
				row[22] = userRow[22]
			} else {
				row[22] = ods.Cell{Type: "empty"}
			}
			row[22].Validation = "valPrio"
			// X: Decision (from user + validation)
			if len(userRow) > 23 {
				row[23] = userRow[23]
			} else {
				row[23] = ods.Cell{Type: "empty"}
			}
			row[23].Validation = "valDec"
			// Y-AA: Freeform (from user)
			for i := 24; i <= 26; i++ {
				if len(userRow) > i {
					row[i] = userRow[i]
				} else {
					row[i] = ods.Cell{Type: "empty"}
				}
			}

			rows = append(rows, row)
		} else {
			// Empty row (padding)
			row := make([]ods.Cell, 27)
			// A: empty
			row[0] = ods.Cell{Type: "empty"}
			// B: validation
			row[1] = ods.Cell{Type: "empty", Validation: "valGat"}
			// C: empty
			row[2] = ods.Cell{Type: "empty"}
			// D: formula
			dFormula := fmt.Sprintf(`IF([.B%d]="";"";VLOOKUP([.B%d];$'Pokemon'.$B$2:$D$%d;2;0))`, r, r, nSpecies+1)
			row[3] = ods.Cell{Type: "formula", Formula: dFormula, Style: "ce_ro"}
			// E: formula
			eFormula := fmt.Sprintf(`IF([.B%d]="";"";VLOOKUP([.B%d];$'Pokemon'.$B$2:$D$%d;3;0))`, r, r, nSpecies+1)
			row[4] = ods.Cell{Type: "formula", Formula: eFormula, Style: "ce_ro"}
			// F-I: empty
			row[5] = ods.Cell{Type: "empty"}
			row[6] = ods.Cell{Type: "empty"}
			row[7] = ods.Cell{Type: "empty"}
			row[8] = ods.Cell{Type: "empty"}
			// J: formula
			jFormula := fmt.Sprintf(`IF(COUNT([.G%d]:[.I%d])=3;([.G%d]+[.H%d]+[.I%d])/45*100;"")`, r, r, r, r, r)
			row[9] = ods.Cell{Type: "formula", Formula: jFormula, Style: "ce_iv"}
			// K-M: validations
			row[10] = ods.Cell{Type: "empty", Validation: fmt.Sprintf("valF%d", r)}
			row[11] = ods.Cell{Type: "empty", Validation: fmt.Sprintf("valC1%d", r)}
			row[12] = ods.Cell{Type: "empty", Validation: fmt.Sprintf("valC2%d", r)}
			// N-S: valX
			for i := 13; i <= 18; i++ {
				row[i] = ods.Cell{Type: "empty", Style: "ce_center", Validation: "valX"}
			}
			// T: empty
			row[19] = ods.Cell{Type: "empty"}
			// U: date validation
			row[20] = ods.Cell{Type: "empty", Style: "ce_date", Validation: "valData"}
			// V: validation
			row[21] = ods.Cell{Type: "empty", Validation: "valPrze"}
			// W: validation
			row[22] = ods.Cell{Type: "empty", Validation: "valPrio"}
			// X: validation
			row[23] = ods.Cell{Type: "empty", Validation: "valDec"}
			// Y-AA: empty
			row[24] = ods.Cell{Type: "empty"}
			row[25] = ods.Cell{Type: "empty"}
			row[26] = ods.Cell{Type: "empty"}

			rows = append(rows, row)
		}
	}
	return rows
}
