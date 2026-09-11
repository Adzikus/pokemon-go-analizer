package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adzik/pokemon/commands"
	"github.com/rs/zerolog"
)

var log = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	binDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	projectRoot := filepath.Join(binDir, "..")

	switch cmd {
	case "verify":
		commands.RunVerify(args, projectRoot)
	case "backup":
		commands.RunBackup(args, projectRoot)
	case "restore":
		commands.RunRestore(args, projectRoot)
	case "analyze":
		commands.RunAnalyze(args, projectRoot)
	case "update-meta":
		commands.RunUpdateMeta(args, projectRoot)
	case "generate":
		commands.RunGenerate(args, projectRoot)
	case "update":
		commands.RunUpdate(args, projectRoot)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Nieznane polecenie: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Pokemon GO — narzedzie do zarzadzania kolekcja

Uzycie:
  pokemon <polecenie> [flagi]

Polecenia:
  verify      Weryfikuje strukture pliku ODS
  backup      Tworzy backup ODS z weryfikacja
  restore     Przywraca ODS z wybranego backupu
  analyze     Wyswietla podsumowanie kolekcji
  update-meta Aktualizuje meta rankingi z PvPoke
  generate    Generuje JSON snapshot i raport Markdown
  update      Regeneruje arkusze pochodne ODS
  help        Wyswietla te pomoc

Przyklady:
  pokemon verify --konto Test
  pokemon backup --konto Test
  pokemon analyze --konto Test
  pokemon generate --konto Test
  pokemon restore --konto Test -- Pokemon-GO-Test.ods.bak.20260910_102018
  pokemon update-meta --force`)
}
