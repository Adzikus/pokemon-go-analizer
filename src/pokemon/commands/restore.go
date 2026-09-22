package commands

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/adzik/pokemon/ods"
)

func RunRestore(args []string, projectRoot string) {
	fs := flag.NewFlagSet("restore", flag.ExitOnError)
	konto := fs.String("konto", "", "Nazwa konta (np. Test)")
	odsPath := fs.String("ods", "", "Sciezka docelowa ODS (opcjonalnie)")
	fs.Parse(args)

	if fs.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "Uzycie: pokemon restore --konto <nazwa> <backup_path>\n")
		fmt.Fprintf(os.Stderr, "  backup_path: nazwa pliku backupu (np. Pokemon-GO-Test.ods.bak.20260910_102018)\n")
		os.Exit(1)
	}

	backupInput := fs.Arg(0)

	path := *odsPath
	if path == "" && *konto != "" {
		path = filepath.Join(projectRoot, fmt.Sprintf("Pokemon-GO-%s.ods", *konto))
	}
	if path == "" {
		fmt.Fprintf(os.Stderr, "Uzycie: pokemon restore --konto <nazwa> <backup_path>\n")
		os.Exit(1)
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(projectRoot, path)
	}

	var backupPath string
	if filepath.IsAbs(backupInput) {
		backupPath = backupInput
	} else {
		candidate := filepath.Join(projectRoot, "backups", backupInput)
		if _, err := os.Stat(candidate); err == nil {
			backupPath = candidate
		} else {
			backupPath = backupInput
			if !filepath.IsAbs(backupPath) {
				backupPath = filepath.Join(projectRoot, backupPath)
			}
		}
	}

	if _, err := os.Stat(backupPath); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: Backup nie istnieje: %s\n", backupPath)
		os.Exit(1)
	}

	if err := ods.VerifyBackupZIP(backupPath); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: Backup jest uszkodzony: %v\n", err)
		os.Exit(1)
	}

	if _, err := os.Stat(path); err == nil {
		preRestoreBackup := path + ".pre-restore"
		src, err := os.Open(path)
		if err == nil {
			dst, err := os.Create(preRestoreBackup)
			if err == nil {
				io.Copy(dst, src)
				dst.Close()
			}
			src.Close()
		}
	}

	srcFile, err := os.Open(backupPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: Otworz backup: %v\n", err)
		os.Exit(1)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: Utworz ODS: %v\n", err)
		os.Exit(1)
	}
	defer dstFile.Close()

	written, err := io.Copy(dstFile, srcFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: Kopiuj: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("SUCCESS: Przywrócono backup (%d bytes)\n  Z: %s\n  Do: %s\n", written, backupPath, path)
}
