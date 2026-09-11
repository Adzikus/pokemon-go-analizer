package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var rankingURLs = map[string]string{
	"great_league_overall":  "https://raw.githubusercontent.com/pvpoke/pvpoke/master/src/data/rankings/all/overall/rankings-1500.json",
	"ultra_league_overall":  "https://raw.githubusercontent.com/pvpoke/pvpoke/master/src/data/rankings/all/overall/rankings-2500.json",
	"master_league_overall": "https://raw.githubusercontent.com/pvpoke/pvpoke/master/src/data/rankings/all/overall/rankings-10000.json",
	"raid_attackers":        "https://raw.githubusercontent.com/pvpoke/pvpoke/master/src/data/rankings/all/attackers/rankings-10000.json",
}

var fallbackRankings = map[string][]string{
	"great_league_overall": {
		"medicham", "azumarill", "galarian_stunfisk", "trevenant", "talonflame", "swampert", "walrein", "nidoqueen", "jellicent", "scrafty", "obstagoon", "sableye", "haunter", "vigoroth", "deoxys_defense", "skarmory", "altaria", "victreebel", "shiftry", "lapras", "whiscash", "quagsire", "lurantis", "mantine", "pelipper", "tangrowth", "ludicolo", "kingdra", "slowbro", "slowbro_galarian", "cresselia", "registeel", "bastiodon", "probopass", "articuno_galarian", "zapdos_galarian", "moltres_galarian",
	},
	"ultra_league_overall": {
		"giratina", "giratina_origin", "cresselia", "swampert", "talonflame", "registeel", "scrafty", "togekiss", "charizard", "victini", "mandibuzz", "obstagoon", "nidoqueen", "walrein", "abomasnow", "lapras", "empoleon", "gyarados", "dragonite", "swampert_shadow", "metagross", "metagross_shadow", "muk_alolan", "muk_alolan_shadow", "ninetales_alolan", "ninetales_alolan_shadow",
	},
	"master_league_overall": {
		"dialga", "groudon", "kyogre", "mewtwo", "garchomp", "metagross", "tyranitar", "rayquaza", "palkia", "giratina_origin", "lugia", "ho_oh", "zacian", "zamazenta", "calyrex_shadow", "calyrex_ice", "necrozma_dusk_mane", "necrozma_dawn_wings", "lunala", "solgaleo", "zarude", "landorus_therian", "kartana", "rillaboom", "melmetal",
	},
	"raid_attackers": {
		"mewtwo", "rayquaza", "kyogre", "groudon", "dialga", "palkia", "giratina_origin", "zacian", "zamazenta", "eternatus", "calyrex_shadow", "calyrex_ice", "necrozma_dusk_mane", "necrozma_dawn_wings", "lunala", "solgaleo", "rampardos", "rhyperior", "tyranitar", "metagross", "garchomp", "kyurem", "reshiram", "zekrom", "xerneas", "yalvelt", "zarude", "kartana", "landorus_therian",
	},
}

type metaRankingsData struct {
	Rankings  map[string][]string `json:"rankings"`
	FetchedAt string              `json:"fetched_at"`
}

func RunUpdateMeta(args []string, projectRoot string) {
	fs := flag.NewFlagSet("update-meta", flag.ExitOnError)
	force := fs.Bool("force", false, "Wymuszenie aktualizacji")
	fs.Parse(args)

	cachePath := filepath.Join(projectRoot, "data", "meta_rankings.json")

	// 1. Check cache
	if !*force {
		if cached, ok := loadMetaCache(cachePath); ok {
			fmt.Printf("Cache OK (wiek: %d dni) — pomijam pobieranie\n", cached.age)
			return
		}
	}

	// 2. Fetch rankings from PvPoke GitHub
	fmt.Println("Pobieranie rankingów z PvPoke GitHub...")
	rankings := make(map[string][]string)
	fetchFailed := false

	for name, url := range rankingURLs {
		speciesIDs, err := fetchRankingJSON(url)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  BŁĄD %s: %v\n", name, err)
			fetchFailed = true
		} else {
			rankings[name] = speciesIDs
			fmt.Printf("  %s: %d gatunków\n", name, len(speciesIDs))
		}
	}

	// 3. Fallback
	if fetchFailed || len(rankings) == 0 {
		fmt.Println("Pobieranie nie powiodło się — używam fallback rankings")
		rankings = fallbackRankings
	}

	// 4. Save
	data := metaRankingsData{
		Rankings:  rankings,
		FetchedAt: time.Now().Format(time.RFC3339),
	}

	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: JSON marshal: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(cachePath, b, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: Zapisz cache: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Zapisano: %s\n", cachePath)
}

type cachedMetaResult struct {
	age int
}

func loadMetaCache(path string) (cachedMetaResult, bool) {
	f, err := os.Open(path)
	if err != nil {
		return cachedMetaResult{}, false
	}
	defer f.Close()

	var data metaRankingsData
	if err := json.NewDecoder(f).Decode(&data); err != nil {
		return cachedMetaResult{}, false
	}

	if data.FetchedAt == "" {
		return cachedMetaResult{}, false
	}

	cached, err := time.Parse(time.RFC3339, data.FetchedAt)
	if err != nil {
		return cachedMetaResult{}, false
	}

	age := int(time.Since(cached).Hours() / 24)
	if age < 30 {
		return cachedMetaResult{age: age}, true
	}

	return cachedMetaResult{}, false
}

func fetchRankingJSON(url string) ([]string, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// PvPoke format: [{"speciesId": "mewtwo", "rating": 950, ...}, ...]
	var entries []struct {
		SpeciesID string `json:"speciesId"`
	}

	if err := json.Unmarshal(body, &entries); err != nil {
		return nil, err
	}

	var speciesIDs []string
	for _, r := range entries {
		if r.SpeciesID != "" {
			speciesIDs = append(speciesIDs, r.SpeciesID)
		}
	}

	return speciesIDs, nil
}
