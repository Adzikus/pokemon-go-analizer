package commands

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func RunBackup(args []string, projectRoot string) {
	fs := flag.NewFlagSet("backup", flag.ExitOnError)
	konto := fs.String("konto", "", "Nazwa konta (np. Test)")
	odsPath := fs.String("ods", "", "Sciezka do ODS (opcjonalnie)")
	fs.Parse(args)

	path := *odsPath
	if path == "" && *konto != "" {
		path = filepath.Join(projectRoot, fmt.Sprintf("Pokemon-GO-%s.ods", *konto))
	}
	if path == "" {
		fmt.Fprintf(os.Stderr, "Uzycie: pokemon backup --konto <nazwa> [--ods <sciezka>]\n")
		os.Exit(1)
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(projectRoot, path)
	}

	if _, err := os.Stat(path); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: Nie mozna odczytac %s: %v\n", path, err)
		os.Exit(1)
	}

	backupDir := filepath.Join(projectRoot, "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: Nie moge utworzyc backups/: %v\n", err)
		os.Exit(1)
	}

	baseName := filepath.Base(path)
	timestamp := time.Now().Format("20060102_150405")
	backupName := fmt.Sprintf("%s.bak.%s", baseName, timestamp)
	backupPath := filepath.Join(backupDir, backupName)

	srcFile, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: Otworz zrodlo: %v\n", err)
		os.Exit(1)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(backupPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: Utworz backup: %v\n", err)
		os.Exit(1)
	}
	defer dstFile.Close()

	written, err := io.Copy(dstFile, srcFile)
	if err != nil {
		os.Remove(backupPath)
		fmt.Fprintf(os.Stderr, "FAIL: Kopiuj: %v\n", err)
		os.Exit(1)
	}

	if err := VerifyBackupZIP(backupPath); err != nil {
		os.Remove(backupPath)
		fmt.Fprintf(os.Stderr, "FAIL: Weryfikacja backupu: %v\n", err)
		os.Exit(1)
	}

	cleanupOldBackups(backupDir, baseName)

	fmt.Printf("SUCCESS: Backup utworzony (%d bytes)\n  %s\n", written, backupPath)
}

func cleanupOldBackups(dir, baseName string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	var backups []string
	prefix := baseName + ".bak."
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), prefix) {
			backups = append(backups, e.Name())
		}
	}

	sort.Sort(sort.Reverse(sort.StringSlice(backups)))

	for i := 5; i < len(backups); i++ {
		os.Remove(filepath.Join(dir, backups[i]))
	}
}
