package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/adzik/pokemon/config"
)

type gamemasterCacheData struct {
	FetchedAt string `json:"fetched_at"`
}

func RunUpdateGamemaster(args []string, projectRoot string) {
	fs := flag.NewFlagSet("update-gamemaster", flag.ExitOnError)
	force := fs.Bool("force", false, "Wymuszenie pobrania (ignoruj cache)")
	fs.Parse(args)

	gmPath := filepath.Join(projectRoot, config.Global.Paths.Data, "gamemaster.json")

	// 1. Check cache
	if !*force {
		if info, err := os.Stat(gmPath); err == nil {
			age := int(time.Since(info.ModTime()).Hours() / 24)
			if age < config.Global.Cache.MaxAgeDays {
				fmt.Printf("Cache OK (wiek: %d dni) — pomijam pobieranie\n", age)
				return
			}
		}
	}

	// 2. Backup old file
	if _, err := os.Stat(gmPath); err == nil {
		if err := backupGamemaster(gmPath); err != nil {
			fmt.Fprintf(os.Stderr, "WARN: Backup: %v\n", err)
		}
	}

	// 3. Fetch
	fmt.Println("Pobieranie gamemaster.json z PvPoke GitHub...")
	client := &http.Client{Timeout: 120 * time.Second}
	req, err := http.NewRequest("GET", config.Global.URLs.Gamemaster, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: Request: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: Fetch: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "FAIL: HTTP %d\n", resp.StatusCode)
		os.Exit(1)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: Read: %v\n", err)
		os.Exit(1)
	}

	// 4. Validate JSON
	var raw json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: Invalid JSON: %v\n", err)
		os.Exit(1)
	}

	// 5. Save
	if err := os.MkdirAll(filepath.Dir(gmPath), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: Mkdir: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(gmPath, body, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: Write: %v\n", err)
		os.Exit(1)
	}

	// 6. Print stats
	var gm GamemasterData
	if err := json.Unmarshal(body, &gm); err == nil {
		fmt.Printf("Zapisano: %s\n", gmPath)
		fmt.Printf("  Gatunki: %d\n", len(gm.Pokemon))
		fmt.Printf("  Ruchy: %d\n", len(gm.Moves))
	} else {
		fmt.Printf("Zapisano: %s (nie moge odczytac statystyk)\n", gmPath)
	}
}

func backupGamemaster(gmPath string) error {
	dir := filepath.Dir(gmPath)
	ts := time.Now().Format("20060102_150405")
	backupName := fmt.Sprintf("gamemaster.json.bak.%s", ts)
	backupPath := filepath.Join(dir, backupName)

	src, err := os.ReadFile(gmPath)
	if err != nil {
		return err
	}
	if err := os.WriteFile(backupPath, src, 0644); err != nil {
		return err
	}

	fmt.Printf("Backup: %s\n", backupName)

	// Cleanup old backups
	cleanOldGamemasterBackups(dir)

	return nil
}

func cleanOldGamemasterBackups(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	var backups []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), "gamemaster.json.bak.") {
			backups = append(backups, e.Name())
		}
	}

	maxBackups := config.Global.Cache.MaxBackups
	if maxBackups <= 0 {
		maxBackups = 5
	}

	if len(backups) <= maxBackups {
		return
	}

	sort.Sort(sort.Reverse(sort.StringSlice(backups)))

	for _, name := range backups[maxBackups:] {
		os.Remove(filepath.Join(dir, name))
	}
}
