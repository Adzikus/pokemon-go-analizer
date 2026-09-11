package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/adzik/pokemon/model"
	"github.com/adzik/pokemon/ods"
)

func RunAnalyze(args []string, projectRoot string) {
	fs := flag.NewFlagSet("analyze", flag.ExitOnError)
	konto := fs.String("konto", "", "Nazwa konta (np. Test)")
	odsPath := fs.String("ods", "", "Sciezka do ODS (opcjonalnie)")
	fs.Parse(args)

	path := *odsPath
	if path == "" && *konto != "" {
		path = filepath.Join(projectRoot, fmt.Sprintf("Pokemon-GO-%s.ods", *konto))
	}
	if path == "" {
		fmt.Fprintf(os.Stderr, "Uzycie: pokemon analyze --konto <nazwa> [--ods <sciezka>]\n")
		os.Exit(1)
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(projectRoot, path)
	}

	gmPath := filepath.Join(projectRoot, "data", "gamemaster.json")
	metaPath := filepath.Join(projectRoot, "data", "meta_rankings.json")

	gmData, err := LoadGamemaster(gmPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: Gamemaster: %v\n", err)
		os.Exit(1)
	}
	gmMap := BuildGamemasterMap(gmData)

	metaData, err := LoadMeta(metaPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: Meta rankings: %v\n", err)
		os.Exit(1)
	}
	metaTiers := BuildMetaTiers(metaData)

	doc, err := ods.Read(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: ODS: %v\n", err)
		os.Exit(1)
	}

	sheet := doc.GetSheet("Kolekcja")
	if sheet == nil {
		fmt.Fprintf(os.Stderr, "FAIL: Brak arkusza Kolekcja\n")
		os.Exit(1)
	}

	collection := ParseCollection(sheet, gmMap, metaTiers)

	// Enrich with dev_score
	EnrichWithScores(collection, gmMap, metaTiers)

	now := time.Now()
	dateStr := now.Format("20060102")

	kontoName := *konto
	if kontoName == "" {
		// Extract from ODS filename: Pokemon-GO-{konto}.ods
		base := filepath.Base(path)
		kontoName = extractKontoFromFilename(base)
	}
	if kontoName == "" {
		kontoName = "unknown"
	}

	// Ensure analizy directory exists
	analizyDir := filepath.Join(projectRoot, "analizy")
	if err := os.MkdirAll(analizyDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: Cannot create analizy dir: %v\n", err)
		os.Exit(1)
	}

	// JSON snapshot
	jsonPath := filepath.Join(analizyDir, fmt.Sprintf("collection-%s-%s.json", kontoName, dateStr))
	writeJSONSnapshot(jsonPath, collection, kontoName)

	// Markdown report
	mdPath := filepath.Join(analizyDir, fmt.Sprintf("Analiza-%s-opencode-%s.md", dateStr, kontoName))
	writeMarkdownReport(mdPath, collection, kontoName, gmData, gmMap, metaData, metaTiers)

	fmt.Printf("SUCCESS: Wygenerowano dla %s (%d Pokemon)\n", kontoName, len(collection))
	fmt.Printf("  JSON: %s\n", jsonPath)
	fmt.Printf("  Markdown: %s\n", mdPath)
}

func writeJSONSnapshot(path string, collection []model.Pokemon, konto string) {
	data := map[string]interface{}{
		"konto":      konto,
		"generated":  time.Now().Format(time.RFC3339),
		"count":      len(collection),
		"collection": collection,
	}
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: JSON marshal: %v\n", err)
		return
	}
	if err := os.WriteFile(path, b, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: Zapisz JSON: %v\n", err)
	}
}

func writeMarkdownReport(path string, collection []model.Pokemon, konto string, gm *GamemasterData, gmMap map[string]*model.GamemasterPokemon, metaData *MetaData, metaTiers map[string]string) {
	var sb strings.Builder
	dateStr := time.Now().Format("2006-01-02")

	sb.WriteString(fmt.Sprintf("# Analiza Pokemon -- %s -- %s\n\n", konto, dateStr))
	sb.WriteString("---\n\n")

	stats := ComputeStats(collection)

	// === SECTION 1 ===
	writeSection1(&sb, collection, stats, metaData)

	// === SECTION 2 ===
	writeSection2(&sb, collection, gmMap)

	// === SECTION 3 ===
	writeSection3(&sb, collection)

	// === SECTION 4 ===
	writeSection4(&sb, collection, gmMap)

	// === SECTION 5 ===
	writeSection5(&sb, collection, gm, metaData)

	// === SECTION 6 ===
	writeSection6(&sb, collection, gmMap, metaTiers)

	// === SECTION 7 ===
	writeSection7(&sb, collection)

	if err := os.WriteFile(path, []byte(sb.String()), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: Zapisz Markdown: %v\n", err)
	}
}

func writeSection1(sb *strings.Builder, c []model.Pokemon, stats CollectionStats, metaData *MetaData) {
	sb.WriteString("## 1. Ogolne podsumowanie\n\n")
	sb.WriteString("### Podsumowanie kolekcji\n\n")

	sb.WriteString(fmt.Sprintf("- **Laczna ilosc Pokemon:** %d\n", stats.Total))
	sb.WriteString(fmt.Sprintf("- **Shiny:** %d (%.1f%%)\n", stats.Shiny, pct(stats.Shiny, stats.Total)))
	sb.WriteString(fmt.Sprintf("- **Lucky:** %d (%.1f%%)\n", stats.Lucky, pct(stats.Lucky, stats.Total)))
	sb.WriteString(fmt.Sprintf("- **Shadow:** %d (%.1f%%)\n", stats.Shadow, pct(stats.Shadow, stats.Total)))
	sb.WriteString(fmt.Sprintf("- **Purified:** %d (%.1f%%)\n", stats.Purified, pct(stats.Purified, stats.Total)))
	sb.WriteString(fmt.Sprintf("- **Dynamax:** %d (%.1f%%)\n", stats.Dynamax, pct(stats.Dynamax, stats.Total)))
	sb.WriteString(fmt.Sprintf("- **Trade'owane:** %d (%.1f%%)\n", stats.Traded, pct(stats.Traded, stats.Total)))
	sb.WriteString(fmt.Sprintf("- **HUNDO (100%% IV):** %d (%.1f%%)\n", stats.Hundo, pct(stats.Hundo, stats.Total)))
	sb.WriteString(fmt.Sprintf("- **High IV (>=90%%):** %d (%.1f%%)\n", stats.HighIV, pct(stats.HighIV, stats.Total)))
	sb.WriteString(fmt.Sprintf("- **Rodziny ewolucyjne:** %d\n", stats.Families))
	sb.WriteString("\n")

	// TOP 10 HUNDO
	sb.WriteString("### TOP 10 HUNDO\n\n")
	hundo := filterBool(c, func(p model.Pokemon) bool { return p.IVPct == 100.0 })
	sortedHundo := SortByCP(hundo)
	for i, p := range TopN(sortedHundo, 10) {
		sb.WriteString(fmt.Sprintf("%d. **%s** -- %d CP, Lucky: %s, Purpose: %s\n",
			i+1, p.Species, p.CP, boolStr(p.IsLucky), purposeLabel(p)))
	}
	sb.WriteString("\n")

	// TOP 10 CP
	sb.WriteString("### TOP 10 CP\n\n")
	sortedCP := SortByCP(c)
	for i, p := range TopN(sortedCP, 10) {
		sb.WriteString(fmt.Sprintf("%d. **%s** -- %d CP, IV: %.1f%%, Lucky: %s\n",
			i+1, p.Species, p.CP, p.IVPct, boolStr(p.IsLucky)))
	}
	sb.WriteString("\n")

	// TOP 10 Dynamax
	sb.WriteString("### TOP 10 Dynamax\n\n")
	dynamax := filterBool(c, func(p model.Pokemon) bool { return p.IsDynamax })
	sortedDyn := SortByCP(dynamax)
	for i, p := range TopN(sortedDyn, 10) {
		sb.WriteString(fmt.Sprintf("%d. **%s** -- %d CP, IV: %.1f%%, Tier: %s\n",
			i+1, p.Species, p.CP, p.IVPct, tierOrDash(p)))
	}
	sb.WriteString("\n")

	// TOP 10 Shadow
	sb.WriteString("### TOP 10 Shadow\n\n")
	shadow := filterBool(c, func(p model.Pokemon) bool { return p.IsShadow })
	sortedShadow := SortByScore(shadow)
	for i, p := range TopN(sortedShadow, 10) {
		sb.WriteString(fmt.Sprintf("%d. **%s** -- %d CP, IV: %.1f%%, Tier: %s, Score: %.1f\n",
			i+1, p.Species, p.CP, p.IVPct, tierOrDash(p), p.DevScore))
	}
	sb.WriteString("\n")

	// TOP 10 Shiny
	sb.WriteString("### TOP 10 Shiny\n\n")
	shiny := filterBool(c, func(p model.Pokemon) bool { return p.IsShiny })
	sortedShiny := SortByScore(shiny)
	for i, p := range TopN(sortedShiny, 10) {
		sb.WriteString(fmt.Sprintf("%d. **%s** -- %d CP, IV: %.1f%%, Tier: %s, Lucky: %s\n",
			i+1, p.Species, p.CP, p.IVPct, tierOrDash(p), boolStr(p.IsLucky)))
	}
	sb.WriteString("\n")

	// TOP 10 Power-Up (IV>=90%, CP<2500, nie na max CP)
	// Wykluczamy: PvP na GL cap, Pokemon na max CP (nie ma sensu podciągać)
	sb.WriteString("### TOP 10 Power-Up\n\n")
	powerUp := filterBool(c, func(p model.Pokemon) bool {
		if p.IVPct < 90 || p.CP >= 2500 {
			return false
		}
		if p.MaxCP > 0 && p.CP >= p.MaxCP {
			return false
		}
		purpose := strings.ToUpper(p.Purpose)
		if purpose == "PVP" && p.CP >= 1450 {
			return false
		}
		return true
	})
	sortedPowerUp := SortByScore(powerUp)
	for i, p := range TopN(sortedPowerUp, 10) {
		sb.WriteString(fmt.Sprintf("%d. **%s** -- %d CP, IV: %.1f%%, Tier: %s\n",
			i+1, p.Species, p.CP, p.IVPct, tierOrDash(p)))
	}
	sb.WriteString("\n")
	sb.WriteString("_Wykluczone: Pokemon z Purpose=PvP na GL cap (CP≥1450) oraz Pokemon na max CP. Sortowanie wg DevScore._\n\n")

	// TOP 10 Dev Score
	sb.WriteString("### TOP 10 Dev Score\n\n")
	sortedDev := SortByScore(c)
	for i, p := range TopN(sortedDev, 10) {
		sb.WriteString(fmt.Sprintf("%d. **%s** -- %d CP, IV: %.1f%%, Tier: %s, Score: %.1f\n",
			i+1, p.Species, p.CP, p.IVPct, tierOrDash(p), p.DevScore))
	}
	sb.WriteString("\n")

	// Meta vs Collection
	if metaData != nil && len(metaData.Rankings.Raid) > 0 {
		rows := BuildMetaComparison(c, metaData.Rankings.Raid)
		sb.WriteString(FormatMetaComparison(rows))
	}

	// Uwagi ogolne
	sb.WriteString("### Uwagi ogolne\n\n")
	sb.WriteString(fmt.Sprintf("- Lucky rate: %.1f%% (%d/%d)\n", pct(stats.Lucky, stats.Total), stats.Lucky, stats.Total))
	sb.WriteString(fmt.Sprintf("- Shadow: %d, Purified: %d\n", stats.Shadow, stats.Purified))
	sb.WriteString(fmt.Sprintf("- Dynamax: %d\n", stats.Dynamax))
	sb.WriteString("\n---\n\n")
}

func writeSection2(sb *strings.Builder, c []model.Pokemon, gmMap map[string]*model.GamemasterPokemon) {
	sb.WriteString("## 2. Zestawy do Rajdow -- Tier List\n\n")

	tiers := BuildRaidTierList(c, gmMap)
	for _, tier := range []string{"S", "A", "B", "C"} {
		entries := tiers[tier]
		if len(entries) == 0 {
			continue
		}
		sb.WriteString(fmt.Sprintf("### Tier %s\n\n", tier))
		limit := safeMin(len(entries), 20)
		for i := 0; i < limit; i++ {
			p := entries[i].Pokemon
			lucky := ""
			if p.IsLucky {
				lucky = " [Lucky]"
			}
			sb.WriteString(fmt.Sprintf("- **%s** -- IV %.1f%%, CP %d, Moves: %s/%s/%s%s\n",
				p.Species, p.IVPct, p.CP,
				strOrDash(p.FastMove), strOrDash(p.Charged1), strOrDash(p.Charged2), lucky))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("### Priorytet inwestycji\n\n")
	sorted := SortByScore(c)
	count := 0
	for _, p := range sorted {
		if p.MetaTier == "S" || p.MetaTier == "A" {
			sb.WriteString(fmt.Sprintf("%d. %s (Score %.1f, Tier %s)\n", count+1, p.Species, p.DevScore, p.MetaTier))
			count++
			if count >= 10 {
				break
			}
		}
	}
	sb.WriteString("\n---\n\n")
}

func writeSection3(sb *strings.Builder, c []model.Pokemon) {
	sb.WriteString("## 3. Zestawy do PvP -- Gotowe druzyny 3v3\n\n")

	for _, league := range []string{"GL", "UL", "ML"} {
		cap := LeagueCaps[league]
		leagueName := league
		switch league {
		case "GL":
			leagueName = "Great League (1500 CP)"
		case "UL":
			leagueName = "Ultra League (2500 CP)"
		case "ML":
			leagueName = "Master League (bez limitu CP)"
		}
		sb.WriteString(fmt.Sprintf("### %s\n\n", leagueName))

		teams := BuildPvPTeams(c, league)
		if len(teams) == 0 {
			sb.WriteString("Brak wystarczajacych kandydatow.\n\n")
			continue
		}
		for _, team := range teams {
			sb.WriteString(fmt.Sprintf("**%s:**\n", team.Name))
			roles := []string{"Lead", "Switch", "Closer"}
			for i, member := range team.Members {
				cp := member.CP
				if cap > 0 && cp > cap {
					cp = cap
				}
				sb.WriteString(fmt.Sprintf("- **%s:** %s -- IV %.1f%%, CP %d, Tier %s\n",
					roles[i], member.Species, member.IVPct, cp, tierOrDash(member)))
			}
			sb.WriteString("\n")
		}
	}
	sb.WriteString("---\n\n")
}

func writeSection4(sb *strings.Builder, c []model.Pokemon, gmMap map[string]*model.GamemasterPokemon) {
	sb.WriteString("## 4. Kandydaci do Transferu\n\n")

	results := BuildTransferCandidates(c, gmMap)

	safeList := []TransferResult{}
	tradeList := []TransferResult{}
	keepList := []TransferResult{}
	softKeepList := []TransferResult{}

	for _, r := range results {
		switch r.Status {
		case Safe:
			safeList = append(safeList, r)
		case Trade:
			tradeList = append(tradeList, r)
		case Keep:
			keepList = append(keepList, r)
		case SoftKeep:
			softKeepList = append(softKeepList, r)
		}
	}

	sb.WriteString(fmt.Sprintf("### Bezpieczne do transferu (%d szt.)\n\n", len(safeList)))
	limit := safeMin(len(safeList), 30)
	for i := 0; i < limit; i++ {
		p := safeList[i].Pokemon
		sb.WriteString(fmt.Sprintf("- **%s** -- IV %.1f%%, CP %d, Tier %s\n",
			p.Species, p.IVPct, p.CP, tierOrDash(p)))
	}
	if len(safeList) > 30 {
		sb.WriteString(fmt.Sprintf("... i %d wiecej\n", len(safeList)-30))
	}
	sb.WriteString("\n")

	if len(tradeList) > 0 {
		sb.WriteString(fmt.Sprintf("### Do Handlu (TRADE) (%d szt.)\n\n", len(tradeList)))
		for _, r := range tradeList {
			p := r.Pokemon
			sb.WriteString(fmt.Sprintf("- **%s** -- IV %.1f%%, CP %d, %s\n",
				p.Species, p.IVPct, p.CP, r.Reason))
		}
		sb.WriteString("\n")
	}

	if len(softKeepList) > 0 {
		sb.WriteString("### Ostrzezenia (SOFT KEEP)\n\n")
		for _, r := range softKeepList {
			p := r.Pokemon
			sb.WriteString(fmt.Sprintf("- **%s** -- %s\n", p.Species, r.Reason))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("### NIE TRANSFERUJ (zachowaj)\n\n")
	sb.WriteString("- **Shiny** -- nawet slabych IV, unikalne\n")
	sb.WriteString("- **Lucky** -- 50% Stardust discount, wartosc handlowa\n")
	sb.WriteString("- **Shadow** -- potencjal Frustration -> Return po eventach\n")
	sb.WriteString("- **HUNDO (100% IV)** -- perfekcyjne IV\n")
	sb.WriteString("- **High IV (>=90%) + meta S/A** -- meta-relevantne\n")
	sb.WriteString("- **Kolekcja / Trade / Lucky Trade** -- oznaczone przez uzytkownika\n")
	sb.WriteString("- **Dynamax** -- nowa mechanika, Max Moves\n")
	sb.WriteString("- **Purified + Return** -- przydatne w PvP\n")
	sb.WriteString("\n---\n\n")
}

func writeSection5(sb *strings.Builder, c []model.Pokemon, gm *GamemasterData, metaData *MetaData) {
	sb.WriteString("## 5. Poszukiwani do statusu Lucky -- Priorytety Lucky Trades\n\n")

	targets := BuildLuckyTargets(c, gm, metaData)

	sb.WriteString("### Priorytet 1: Legendary / Mythical (Meta S-tier)\n\n")
	for _, t := range targets {
		if t.Priority != 1 {
			continue
		}
		status := "NO BRAK Lucky"
		if t.HasLucky {
			status = "YES MASZ Lucky"
		}
		sb.WriteString(fmt.Sprintf("- **%s** -- %s\n", t.Reason, status))
	}
	sb.WriteString("\n")

	sb.WriteString("### Priorytet 2: Shadow-eligible Meta\n\n")
	for _, t := range targets {
		if t.Priority != 2 {
			continue
		}
		status := "NO BRAK Lucky"
		if t.HasLucky {
			status = "YES MASZ Lucky"
		}
		sb.WriteString(fmt.Sprintf("- **%s** -- %s\n", t.Reason, status))
	}
	sb.WriteString("\n")

	sb.WriteString("### Strategia Lucky Trade\n\n")
	sb.WriteString("1. Handel to jedyny sposob na uzyskanie Lucky -- im wiecej handlujesz, tym wyzszy rate\n")
	sb.WriteString("2. Gwarantowany Lucky Trade (Lucky Friend) przeznaczaj na legendy, ktorych nie masz\n")
	sb.WriteString("3. Nie handluj HUNDO -- juz sa idealne\n")
	sb.WriteString("4. Im starszy Pokemon (wiek w kolekcji), tym wyzsza szansa na Lucky\n")
	sb.WriteString("5. Candy XL z handlu zalezy od odleglosci łapania (100km+ = 100% szansa XL)\n")
	sb.WriteString("\n---\n\n")
}

func writeSection6(sb *strings.Builder, c []model.Pokemon, gmMap map[string]*model.GamemasterPokemon, metaTiers map[string]string) {
	sb.WriteString("## 6. Szczegolowa analiza -- Ocena 1-10\n\n")

	// Legendary / Mythical
	sb.WriteString("### Legendary / Mythical\n\n")
	legendary := filterBool(c, func(p model.Pokemon) bool {
		return (p.MetaTier == "S" || p.MetaTier == "A") && (p.BaseStats.Atk+p.BaseStats.Def+p.BaseStats.HP) >= 600
	})
	if len(legendary) == 0 {
		sb.WriteString("Brak legendarnych w kolekcji.\n\n")
	} else {
		sorted := SortByScore(legendary)
		for _, p := range sorted {
			rating := RatingFromScore(p.DevScore)
			sb.WriteString(formatDetailWithRating(p, rating))
		}
	}

	// HUNDO
	sb.WriteString("### HUNDO\n\n")
	hundo := filterBool(c, func(p model.Pokemon) bool { return p.IVPct == 100.0 })
	if len(hundo) == 0 {
		sb.WriteString("Brak HUNDO.\n\n")
	} else {
		sorted := SortByCP(hundo)
		for _, p := range sorted {
			sb.WriteString(fmt.Sprintf("- **%s** -- CP %d, Lucky: %s, Shadow: %s, Rekomendacja: MAX OUT (L50)\n",
				p.Species, p.CP, boolStr(p.IsLucky), boolStr(p.IsShadow)))
		}
	}
	sb.WriteString("\n")

	// Shadow
	sb.WriteString("### Shadow -- Frustration -> Return\n\n")
	shadow := filterBool(c, func(p model.Pokemon) bool { return p.IsShadow })
	if len(shadow) == 0 {
		sb.WriteString("Brak Shadow.\n\n")
	} else {
		sorted := SortByScore(shadow)
		for _, p := range sorted {
			rating := RatingFromScore(p.DevScore)
			sb.WriteString(fmt.Sprintf("- **%s** -- IV %.1f%%, Lucky: %s, Ocena: %d/10, Rekomendacja: %s\n",
				p.Species, p.IVPct, boolStr(p.IsLucky), rating, Recommendation(p, rating)))
		}
	}
	sb.WriteString("\n")

	// PvP Relevant
	sb.WriteString("### PvP Relevant\n\n")
	pvp := filterBool(c, func(p model.Pokemon) bool {
		purpose := strings.ToLower(p.Purpose)
		return purpose == "pvp" || purpose == "gl" || purpose == "ul" || purpose == "ml" ||
			p.MetaTier == "S" || p.MetaTier == "A"
	})
	if len(pvp) == 0 {
		sb.WriteString("Brak PvP-relevantnych.\n\n")
	} else {
		sorted := SortByScore(pvp)
		for _, p := range TopN(sorted, 30) {
			rating := RatingFromScore(p.DevScore)
			sb.WriteString(fmt.Sprintf("- **%s** -- IV %.1f%%, CP %d, Lucky: %s, Moves: %s/%s/%s, Ocena: %d/10\n",
				p.Species, p.IVPct, p.CP, boolStr(p.IsLucky),
				strOrDash(p.FastMove), strOrDash(p.Charged1), strOrDash(p.Charged2), rating))
		}
	}
	sb.WriteString("\n")

	// Shiny
	sb.WriteString("### Shiny\n\n")
	shiny := filterBool(c, func(p model.Pokemon) bool { return p.IsShiny })
	if len(shiny) == 0 {
		sb.WriteString("Brak Shiny.\n\n")
	} else {
		sorted := SortByScore(shiny)
		for _, p := range sorted {
			rating := RatingFromScore(p.DevScore)
			starMeta := ""
			if p.MetaTier == "S" || p.MetaTier == "A" {
				starMeta = " [META]"
			}
			sb.WriteString(fmt.Sprintf("- **%s** -- IV %.1f%%, Lucky: %s%s, Ocena: %d/10, Rekomendacja: %s\n",
				p.Species, p.IVPct, boolStr(p.IsLucky), starMeta, rating, Recommendation(p, rating)))
		}
	}
	sb.WriteString("\n")

	// Inne
	sb.WriteString("### Inne (Dynamax / Purified / Traded)\n\n")
	other := filterBool(c, func(p model.Pokemon) bool {
		return p.IsDynamax || (p.IsPurified && p.IVPct >= 80) || (p.IsTraded && p.IVPct >= 90)
	})
	if len(other) == 0 {
		sb.WriteString("Brak.\n\n")
	} else {
		sorted := SortByScore(other)
		for _, p := range sorted {
			rating := RatingFromScore(p.DevScore)
			sb.WriteString(fmt.Sprintf("- **%s** -- IV %.1f%%, Tags: %s, Ocena: %d/10, Rekomendacja: %s\n",
				p.Species, p.IVPct, formatTags(p), rating, Recommendation(p, rating)))
		}
	}
	sb.WriteString("\n---\n\n")
}

func writeSection7(sb *strings.Builder, c []model.Pokemon) {
	sb.WriteString(fmt.Sprintf("## 7. Wszystkie Pokemon-y -- Ranking Rozwoju (%d sztuk)\n\n", len(c)))

	ranking := BuildFullRanking(c)

	sb.WriteString("| # | UID | Pokemon | Score | IV% | Lucky | Shadow | Tier | CP | Liga | Moves | Przeznaczenie |\n")
	sb.WriteString("|--:|-----|---------|------:|----:|:-----:|:------:|:----:|---:|:-----|:------|:-------------|\n")

	for _, row := range ranking {
		p := row.Pokemon
		lucky := FormatLucky(p)
		shadow := FormatShadow(p)
		tier := tierOrDash(p)
		moveStr := fmt.Sprintf("%s/%s/%s", strOrDash(p.FastMove), strOrDash(p.Charged1), strOrDash(p.Charged2))
		if len(moveStr) > 30 {
			moveStr = moveStr[:27] + "..."
		}
		sb.WriteString(fmt.Sprintf("| %d | %s | **%s** %s | %.1f | %.1f%% | %s | %s | %s | %4d | %s | %s | %s |\n",
			row.Rank, p.UID, p.Species, FormatEmojis(p),
			p.DevScore, p.IVPct, lucky, shadow, tier,
			p.CP, "-", moveStr, purposeLabel(p)))
	}
	sb.WriteString("\n")
}

// --- Helpers ---

func boolStr(b bool) string {
	if b {
		return "TAK"
	}
	return "NIE"
}

func strOrDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func tierOrDash(p model.Pokemon) string {
	if p.MetaTier == "" {
		return "-"
	}
	return p.MetaTier
}

func formatTags(p model.Pokemon) string {
	var tags []string
	if p.IsDynamax {
		tags = append(tags, "Dynamax")
	}
	if p.IsPurified {
		tags = append(tags, "Purified")
	}
	if p.IsTraded {
		tags = append(tags, "Traded")
	}
	if len(tags) == 0 {
		return "-"
	}
	return strings.Join(tags, ", ")
}

func formatDetailWithRating(p model.Pokemon, rating int) string {
	return fmt.Sprintf("- **%s** -- Ocena: %d/10, IV %.1f%%, Lucky: %s, CP %d, Zastosowanie: %s, Rekomendacja: %s\n",
		p.Species, rating, p.IVPct, boolStr(p.IsLucky), p.CP, purposeLabel(p), Recommendation(p, rating))
}

func filterBool(c []model.Pokemon, fn func(model.Pokemon) bool) []model.Pokemon {
	var result []model.Pokemon
	for _, p := range c {
		if fn(p) {
			result = append(result, p)
		}
	}
	return result
}

func pct(a, b int) float64 {
	if b == 0 {
		return 0
	}
	return float64(a) / float64(b) * 100
}

func sortByScoreLocal(c []model.Pokemon) []model.Pokemon {
	sorted := make([]model.Pokemon, len(c))
	copy(sorted, c)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].DevScore != sorted[j].DevScore {
			return sorted[i].DevScore > sorted[j].DevScore
		}
		if sorted[i].IVPct != sorted[j].IVPct {
			return sorted[i].IVPct > sorted[j].IVPct
		}
		return sorted[i].CP > sorted[j].CP
	})
	return sorted
}
