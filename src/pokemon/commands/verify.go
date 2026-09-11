package commands

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adzik/pokemon/ods"
)

func RunVerify(args []string, projectRoot string) {
	fs := flag.NewFlagSet("verify", flag.ExitOnError)
	konto := fs.String("konto", "", "Nazwa konta (np. Test)")
	odsPath := fs.String("ods", "", "Sciezka do ODS (opcjonalnie)")
	fs.Parse(args)

	path := *odsPath
	if path == "" && *konto != "" {
		path = filepath.Join(projectRoot, fmt.Sprintf("Pokemon-GO-%s.ods", *konto))
	}
	if path == "" {
		fmt.Fprintf(os.Stderr, "Uzycie: pokemon verify --konto <nazwa> [--ods <sciezka>]\n")
		os.Exit(1)
	}

	fmt.Printf("Weryfikacja: %s\n", path)

	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "FAIL: Plik nie istnieje: %s\n", path)
		os.Exit(1)
	}
	fmt.Printf("  Rozmiar: %d bytes\n", info.Size())

	odsDoc, err := ods.Read(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
		os.Exit(1)
	}

	expectedSheets := []string{"Kolekcja", "Pokemon", "Ruchy", "Ataki", "Pokedex"}
	found := odsDoc.SheetNames()

	fmt.Printf("  Arkusze: %d\n", len(found))
	for _, name := range found {
		sheet := odsDoc.GetSheet(name)
		rows := len(sheet.Rows)
		cells := 0
		for _, row := range sheet.Rows {
			cells += len(row)
		}
		fmt.Printf("    %-15s — %4d wierszy, %6d komorek\n", name, rows, cells)
	}

	missing := []string{}
	for _, expected := range expectedSheets {
		foundSheet := false
		for _, name := range found {
			if name == expected {
				foundSheet = true
				break
			}
		}
		if !foundSheet {
			missing = append(missing, expected)
		}
	}

	kolekcja := odsDoc.GetSheet("Kolekcja")
	if kolekcja == nil {
		fmt.Fprintf(os.Stderr, "FAIL: Brak arkusza 'Kolekcja'\n")
		os.Exit(1)
	}

	if len(kolekcja.Rows) < 10 {
		fmt.Fprintf(os.Stderr, "FAIL: Za malo wierszy w Kolekcja: %d (min 10)\n", len(kolekcja.Rows))
		os.Exit(1)
	}

	if len(kolekcja.Rows) < 2 || len(kolekcja.Rows[1]) == 0 || !strings.HasPrefix(kolekcja.Rows[1][0], "P-") {
		fmt.Fprintf(os.Stderr, "FAIL: Drugi wiersz nie zaczyna sie od P- (UID)\n")
		os.Exit(1)
	}

	fmt.Println()
	if len(missing) > 0 {
		fmt.Printf("UWAGA: Brakujace arkusze: %s\n", strings.Join(missing, ", "))
	}

	kontoName := *konto
	if kontoName == "" {
		kontoName = filepath.Base(path)
	}
	fmt.Printf("SUCCESS: ODS poprawny (%s)\n", kontoName)
}
