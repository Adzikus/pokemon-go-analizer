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

	"github.com/adzik/pokemon/config"
	"github.com/adzik/pokemon/model"
	"github.com/adzik/pokemon/ods"
	"github.com/adzik/pokemon/pkg/analysis"
	"github.com/adzik/pokemon/pkg/gamemaster"
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

	gmData, err := gamemaster.LoadGamemaster(gmPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: Gamemaster: %v\n", err)
		os.Exit(1)
	}
	gmMap := gamemaster.BuildGamemasterMap(gmData)
	moveMap := gamemaster.BuildMoveMap(gmData)

	metaData, err := gamemaster.LoadMeta(metaPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: Meta rankings: %v\n", err)
		os.Exit(1)
	}
	metaTiers := gamemaster.BuildMetaTiers(metaData)
	metaCats := gamemaster.BuildMetaTiersByCategory(metaData)

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

	collection := gamemaster.ParseCollection(sheet, gmMap, metaCats)

	// Enrich with per-context scores and raid DPS
	analysis.EnrichWithScores(collection, gmMap, moveMap, metaCats)

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
	mdPath := filepath.Join(analizyDir, fmt.Sprintf("Analiza-%s-%s-%s.md", dateStr, config.Global.Report.ModelName, kontoName))
	writeMarkdownReport(mdPath, collection, kontoName, gmData, gmMap, moveMap, metaData, metaTiers, metaCats)

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

func writeMarkdownReport(
	path string,
	collection []model.Pokemon,
	konto string,
	gm *gamemaster.GamemasterData,
	gmMap map[string]*model.GamemasterPokemon,
	moveMap map[string]*gamemaster.MoveEntry,
	metaData *gamemaster.MetaData,
	metaTiers map[string]string,
	metaCats gamemaster.MetaTiersByCategory,
) {
	var sb strings.Builder
	dateStr := time.Now().Format("2006-01-02")

	sb.WriteString(fmt.Sprintf("# Analiza Pokemon -- %s -- %s\n\n", konto, dateStr))
	sb.WriteString("---\n\n")

	stats := analysis.ComputeStats(collection)

	// === SECTION 1 ===
	writeSection1(&sb, collection, stats, metaData, gmMap)

	// === SECTION 2 ===
	writeSection2(&sb, collection, gmMap, moveMap)

	// === SECTION 3 ===
	writeSection3(&sb, collection)

	// === SECTION 4 ===
	writeSection4(&sb, collection, gmMap)

	// === SECTION 5 ===
	writeSection5(&sb, collection, gm, metaData)

	// === SECTION 6 ===
	writeSection6(&sb, collection, gmMap, metaCats)

	// === SECTION 7 ===
	writeSection7(&sb, collection)

	if err := os.WriteFile(path, []byte(sb.String()), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: Zapisz Markdown: %v\n", err)
	}
}

func writeSection1(sb *strings.Builder, c []model.Pokemon, stats analysis.CollectionStats, metaData *gamemaster.MetaData, gmMap map[string]*model.GamemasterPokemon) {
	sb.WriteString("## 1. Ogolne podsumowanie\n\n")
	sb.WriteString("### Podsumowanie kolekcji\n\n")

	sb.WriteString(fmt.Sprintf("- **Laczna ilosc Pokemon:** %d\n", stats.Total))
	sb.WriteString(fmt.Sprintf("- **Shiny:** %d (%.1f%%)\n", stats.Shiny, analysis.Pct(stats.Shiny, stats.Total)))
	sb.WriteString(fmt.Sprintf("- **Lucky:** %d (%.1f%%)\n", stats.Lucky, analysis.Pct(stats.Lucky, stats.Total)))
	sb.WriteString(fmt.Sprintf("- **Shadow:** %d (%.1f%%)\n", stats.Shadow, analysis.Pct(stats.Shadow, stats.Total)))
	sb.WriteString(fmt.Sprintf("- **Purified:** %d (%.1f%%)\n", stats.Purified, analysis.Pct(stats.Purified, stats.Total)))
	sb.WriteString(fmt.Sprintf("- **Dynamax:** %d (%.1f%%)\n", stats.Dynamax, analysis.Pct(stats.Dynamax, stats.Total)))
	sb.WriteString(fmt.Sprintf("- **Trade'owane:** %d (%.1f%%)\n", stats.Traded, analysis.Pct(stats.Traded, stats.Total)))
	sb.WriteString(fmt.Sprintf("- **HUNDO (100%% IV):** %d (%.1f%%)\n", stats.Hundo, analysis.Pct(stats.Hundo, stats.Total)))
	sb.WriteString(fmt.Sprintf("- **High IV (>=90%%):** %d (%.1f%%)\n", stats.HighIV, analysis.Pct(stats.HighIV, stats.Total)))
	sb.WriteString(fmt.Sprintf("- **Rodziny ewolucyjne:** %d\n", stats.Families))
	sb.WriteString("\n")

	// TOP 10 HUNDO
	sb.WriteString("### TOP 10 HUNDO\n\n")
	hundo := analysis.FilterBool(c, func(p model.Pokemon) bool { return p.IVPct == 100.0 })
	sortedHundo := analysis.SortByCP(hundo)
	sb.WriteString("| # | Pokemon | CP | Lucky | Purpose |\n")
	sb.WriteString("|--:|---------|---:|:-----:|---------|\n")
	for i, p := range analysis.TopN(sortedHundo, 10) {
		sb.WriteString(fmt.Sprintf("| %d | **%s** | %d | %s | %s |\n",
			i+1, p.Species, p.CP, analysis.BoolStr(p.IsLucky), analysis.PurposeLabel(p)))
	}
	sb.WriteString("\n")

	// TOP 10 CP
	sb.WriteString("### TOP 10 CP\n\n")
	sortedCP := analysis.SortByCP(c)
	sb.WriteString("| # | Pokemon | CP | IV% | Lucky | Purpose |\n")
	sb.WriteString("|--:|---------|---:|----:|:-----:|---------|\n")
	for i, p := range analysis.TopN(sortedCP, 10) {
		sb.WriteString(fmt.Sprintf("| %d | **%s** | %d | %.1f%% | %s | %s |\n",
			i+1, p.Species, p.CP, p.IVPct, analysis.BoolStr(p.IsLucky), analysis.PurposeLabel(p)))
	}
	sb.WriteString("\n")

	// TOP 10 Dynamax
	sb.WriteString("### TOP 10 Dynamax\n\n")
	dynamax := analysis.FilterBool(c, func(p model.Pokemon) bool { return p.IsDynamax })
	sortedDyn := analysis.SortByCP(dynamax)
	sb.WriteString("| # | Pokemon | CP | IV% | Tier |\n")
	sb.WriteString("|--:|---------|---:|----:|:----:|\n")
	for i, p := range analysis.TopN(sortedDyn, 10) {
		sb.WriteString(fmt.Sprintf("| %d | **%s** | %d | %.1f%% | %s |\n",
			i+1, p.Species, p.CP, p.IVPct, analysis.TierOrDash(p)))
	}
	sb.WriteString("\n")

	// TOP 10 Shadow
	sb.WriteString("### TOP 10 Shadow\n\n")
	shadow := analysis.FilterBool(c, func(p model.Pokemon) bool { return p.IsShadow })
	sortedShadow := analysis.SortByScore(shadow)
	sb.WriteString("| # | Pokemon | CP | IV% | Tier | Score |\n")
	sb.WriteString("|--:|---------|---:|----:|:----:|------:|\n")
	for i, p := range analysis.TopN(sortedShadow, 10) {
		sb.WriteString(fmt.Sprintf("| %d | **%s** | %d | %.1f%% | %s | %.1f |\n",
			i+1, p.Species, p.CP, p.IVPct, analysis.TierOrDash(p), p.DevScore))
	}
	sb.WriteString("\n")

	// TOP 10 Shiny
	sb.WriteString("### TOP 10 Shiny\n\n")
	shiny := analysis.FilterBool(c, func(p model.Pokemon) bool { return p.IsShiny })
	sortedShiny := analysis.SortByScore(shiny)
	sb.WriteString("| # | Pokemon | CP | IV% | Tier | Lucky |\n")
	sb.WriteString("|--:|---------|---:|----:|:----:|:-----:|\n")
	for i, p := range analysis.TopN(sortedShiny, 10) {
		sb.WriteString(fmt.Sprintf("| %d | **%s** | %d | %.1f%% | %s | %s |\n",
			i+1, p.Species, p.CP, p.IVPct, analysis.TierOrDash(p), analysis.BoolStr(p.IsLucky)))
	}
	sb.WriteString("\n")

	// TOP 10 Power-Up (IV>=90%, CP<2500, nie na max CP)
	// Wykluczamy: PvP na GL cap, Pokemon na max CP (nie ma sensu podciągać)
	sb.WriteString("### TOP 10 Power-Up\n\n")
	powerUp := analysis.FilterBool(c, func(p model.Pokemon) bool {
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
	sortedPowerUp := analysis.SortByScore(powerUp)
	sb.WriteString("| # | Pokemon | CP | IV% | Raid Tier | Score |\n")
	sb.WriteString("|--:|---------|---:|----:|:---------:|------:|\n")
	for i, p := range analysis.TopN(sortedPowerUp, 10) {
		sb.WriteString(fmt.Sprintf("| %d | **%s** | %d | %.1f%% | %s | %.1f |\n",
			i+1, p.Species, p.CP, p.IVPct, analysis.TierOrDash(p), p.DevScore))
	}
	sb.WriteString("\n")
	sb.WriteString("_Wykluczone: Pokemon z Purpose=PvP na GL cap (CP≥1450) oraz Pokemon na max CP. Sortowanie wg DevScore._\n\n")

	// TOP 10 Dev Score
	sb.WriteString("### TOP 10 Dev Score\n\n")
	sortedDev := analysis.SortByScore(c)
	sb.WriteString("| # | Pokemon | CP | IV% | Raid Tier | Score |\n")
	sb.WriteString("|--:|---------|---:|----:|:---------:|------:|\n")
	for i, p := range analysis.TopN(sortedDev, 10) {
		sb.WriteString(fmt.Sprintf("| %d | **%s** | %d | %.1f%% | %s | %.1f |\n",
			i+1, p.Species, p.CP, p.IVPct, analysis.TierOrDash(p), p.DevScore))
	}
	sb.WriteString("\n")

	// Per-Context TOP 10
	sb.WriteString("### TOP 10 per kontekst\n\n")
	type ctxDef struct {
		name     string
		getTier  func(model.Pokemon) string
		getRank  func(model.Pokemon) int
		getScore func(model.Pokemon) float64
	}
	contexts := []ctxDef{
		{"Raid", func(p model.Pokemon) string { return p.RaidTier }, func(p model.Pokemon) int { return p.RaidRank }, func(p model.Pokemon) float64 { return p.RaidScore }},
		{"Master League", func(p model.Pokemon) string { return p.MLTier }, func(p model.Pokemon) int { return p.MLRank }, func(p model.Pokemon) float64 { return p.MLScore }},
		{"Ultra League", func(p model.Pokemon) string { return p.ULTier }, func(p model.Pokemon) int { return p.ULRank }, func(p model.Pokemon) float64 { return p.ULScore }},
		{"Great League", func(p model.Pokemon) string { return p.GLTier }, func(p model.Pokemon) int { return p.GLRank }, func(p model.Pokemon) float64 { return p.GLScore }},
	}
	for _, ctx := range contexts {
		scored := analysis.FilterBool(c, func(p model.Pokemon) bool { return ctx.getTier(p) != "" })
		sort.Slice(scored, func(i, j int) bool {
			ri, rj := ctx.getRank(scored[i]), ctx.getRank(scored[j])
			if ri != rj {
				return ri < rj
			}
			return ctx.getScore(scored[i]) > ctx.getScore(scored[j])
		})
		sb.WriteString(fmt.Sprintf("#### %s (Tier S/A+)\n\n", ctx.name))
		sb.WriteString("| # | Pokemon | CP | IV% | Tier | Rank | Score |\n")
		sb.WriteString("|--:|---------|---:|----:|:----:|-----:|------:|\n")
		count := 0
		for _, p := range scored {
			tier := ctx.getTier(p)
			if tier != "S" && tier != "A+" && tier != "A" {
				continue
			}
			sb.WriteString(fmt.Sprintf("| %d | **%s** | %d | %.1f%% | %s | #%d | %.1f |\n",
				count+1, p.Species, p.CP, p.IVPct, tier, ctx.getRank(p)+1, ctx.getScore(p)))
			count++
			if count >= 10 {
				break
			}
		}
		sb.WriteString("\n")
	}

	// Meta vs Collection
	if metaData != nil && len(metaData.Rankings.Raid) > 0 {
		rows := analysis.BuildMetaComparison(c, metaData.Rankings.Raid, gmMap)
		sb.WriteString(analysis.FormatMetaComparison(rows))
	}

	// Uwagi ogolne
	sb.WriteString("### Uwagi ogolne\n\n")
	sb.WriteString(fmt.Sprintf("- Lucky rate: %.1f%% (%d/%d)\n", 	analysis.Pct(stats.Lucky, stats.Total), stats.Lucky, stats.Total))
	sb.WriteString(fmt.Sprintf("- Shadow: %d, Purified: %d\n", stats.Shadow, stats.Purified))
	sb.WriteString(fmt.Sprintf("- Dynamax: %d\n", stats.Dynamax))
	sb.WriteString("\n---\n\n")
}

func writeSection2(sb *strings.Builder, c []model.Pokemon, gmMap map[string]*model.GamemasterPokemon, moveMap map[string]*gamemaster.MoveEntry) {
	sb.WriteString("## 2. Rajdy -- Zalecane Zespoly\n\n")

	rankings := analysis.BuildRaidRankings(c, gmMap, moveMap, 10)

	for _, bossType := range analysis.BossTypes {
		counters := rankings.ByType[bossType]
		if len(counters) == 0 {
			continue
		}
		typeName := strings.Title(bossType)
		sb.WriteString(fmt.Sprintf("### vs %s-type\n\n", typeName))
		sb.WriteString("| # | UID | Pokemon | ATK | DPS | Optimal | MoveDPS (fast/charged) | Moves (actual) | Moves (best) | IV | CP |\n")
		sb.WriteString("|--:|-----|---------|----:|----:|--------:|-----------------------|----------------|--------------|---:|---:|\n")
		for i, r := range counters {
			p := r.Pokemon
			movesActual := fmt.Sprintf("%s / %s", analysis.StrOrDash(r.FastMove), analysis.StrOrDash(r.ChargedMove))
			movesBest := fmt.Sprintf("%s / %s", analysis.StrOrDash(r.BestFast), analysis.StrOrDash(r.BestCharged))
			if len(movesActual) > config.Global.Report.MaxMoveDisplay {
				movesActual = movesActual[:config.Global.Report.MaxMoveDisplay-3] + "..."
			}
			if len(movesBest) > config.Global.Report.MaxMoveDisplay {
				movesBest = movesBest[:config.Global.Report.MaxMoveDisplay-3] + "..."
			}
			moveDPS := fmt.Sprintf("%.1f / %.1f", p.MoveDPSFast, p.MoveDPSCharged)
			sb.WriteString(fmt.Sprintf("| %d | %s | **%s** | %.0f | %.1f | %.1f | %s | %s | %s | %.1f%% | %d |\n",
				i+1, p.UID, p.Species, p.EffectiveATK, r.DPS, r.OptimalDPS,
				moveDPS, movesActual, movesBest,
				p.IVPct, p.CP))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("### Uniwersalny (wszystkie typy)\n\n")
	if len(rankings.Universal) > 0 {
		sb.WriteString("| # | UID | Pokemon | ATK | DPS | Optimal | MoveDPS (fast/charged) | Moves (actual) | Moves (best) | IV | CP |\n")
		sb.WriteString("|--:|-----|---------|----:|----:|--------:|-----------------------|----------------|--------------|---:|---:|\n")
		for i, r := range rankings.Universal {
			p := r.Pokemon
			movesActual := fmt.Sprintf("%s / %s", analysis.StrOrDash(r.FastMove), analysis.StrOrDash(r.ChargedMove))
			movesBest := fmt.Sprintf("%s / %s", analysis.StrOrDash(r.BestFast), analysis.StrOrDash(r.BestCharged))
			if len(movesActual) > config.Global.Report.MaxMoveDisplay {
				movesActual = movesActual[:config.Global.Report.MaxMoveDisplay-3] + "..."
			}
			if len(movesBest) > config.Global.Report.MaxMoveDisplay {
				movesBest = movesBest[:config.Global.Report.MaxMoveDisplay-3] + "..."
			}
			moveDPS := fmt.Sprintf("%.1f / %.1f", p.MoveDPSFast, p.MoveDPSCharged)
			sb.WriteString(fmt.Sprintf("| %d | %s | **%s** | %.0f | %.1f | %.1f | %s | %s | %s | %.1f%% | %d |\n",
				i+1, p.UID, p.Species, p.EffectiveATK, r.DPS, r.OptimalDPS,
				moveDPS, movesActual, movesBest,
				p.IVPct, p.CP))
		}
	} else {
		sb.WriteString("Brak Pokemon z CP >= 1000.\n")
	}
	sb.WriteString("\n---\n\n")
}

func writeSection3(sb *strings.Builder, c []model.Pokemon) {
	sb.WriteString("## 3. Zestawy do PvP -- Gotowe druzyny 3v3\n\n")

	for _, league := range []string{"GL", "UL", "ML"} {
		cap := analysis.GetLeagueCaps()[league]
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

		teams := analysis.BuildPvPTeams(c, league)
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
				tier := analysis.LeagueTier(member, league)
				sb.WriteString(fmt.Sprintf("- **%s:** %s -- IV %.1f%%, CP %d, Tier %s, Score %.1f\n",
					roles[i], member.Species, member.IVPct, cp, analysis.StrOrDash(tier), leagueScore(member, league)))
			}
			sb.WriteString("\n")
		}
	}

	// PvP Summary table
	sb.WriteString("### Podsumowanie PvP — wszystkie ligi\n\n")
	pvp := analysis.FilterBool(c, func(p model.Pokemon) bool {
		purpose := strings.ToLower(p.Purpose)
		return purpose == "pvp" || purpose == "gl" || purpose == "ul" || purpose == "ml"
	})
	if len(pvp) > 0 {
		sb.WriteString("| Pokemon | CP | IV% | GL Tier | UL Tier | ML Tier | DevScore |\n")
		sb.WriteString("|---------|---:|----:|:-------:|:-------:|:-------:|---------:|\n")
		for _, p := range analysis.TopN(analysis.SortByScore(pvp), 20) {
			sb.WriteString(fmt.Sprintf("| **%s** | %d | %.1f%% | %s | %s | %s | %.1f |\n",
				p.Species, p.CP, p.IVPct, analysis.StrOrDash(p.GLTier), analysis.StrOrDash(p.ULTier), analysis.StrOrDash(p.MLTier), p.DevScore))
		}
	} else {
		sb.WriteString("Brak Pokemon z Purpose=PvP.\n")
	}
	sb.WriteString("\n---\n\n")
}

func leagueScore(p model.Pokemon, league string) float64 {
	switch league {
	case "GL":
		return p.GLScore
	case "UL":
		return p.ULScore
	case "ML":
		return p.MLScore
	}
	return 0
}

func writeSection4(sb *strings.Builder, c []model.Pokemon, gmMap map[string]*model.GamemasterPokemon) {
	sb.WriteString("## 4. Kandydaci do Transferu\n\n")

	results := analysis.BuildTransferCandidates(c, gmMap)

	safeList := []analysis.TransferResult{}
	tradeList := []analysis.TransferResult{}
	keepList := []analysis.TransferResult{}
	softKeepList := []analysis.TransferResult{}

	for _, r := range results {
		switch r.Status {
		case analysis.Safe:
			safeList = append(safeList, r)
		case analysis.Trade:
			tradeList = append(tradeList, r)
		case analysis.Keep:
			keepList = append(keepList, r)
		case analysis.SoftKeep:
			softKeepList = append(softKeepList, r)
		}
	}

	sb.WriteString(fmt.Sprintf("### Bezpieczne do transferu (%d szt.)\n\n", len(safeList)))
	limit := analysis.SafeMin(len(safeList), 30)
	for i := 0; i < limit; i++ {
		p := safeList[i].Pokemon
		sb.WriteString(fmt.Sprintf("- **%s** -- IV %.1f%%, CP %d, Tier %s\n",
			p.Species, p.IVPct, p.CP, analysis.TierOrDash(p)))
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
	sb.WriteString("- **Kolekcja / analysis.Trade / Lucky analysis.Trade** -- oznaczone przez uzytkownika\n")
	sb.WriteString("- **Dynamax** -- nowa mechanika, Max Moves\n")
	sb.WriteString("- **Purified + Return** -- przydatne w PvP\n")
	sb.WriteString("\n---\n\n")
}

func writeSection5(sb *strings.Builder, c []model.Pokemon, gm *gamemaster.GamemasterData, metaData *gamemaster.MetaData) {
	sb.WriteString("## 5. Poszukiwani do statusu Lucky -- Priorytety Lucky Trades\n\n")

	targets := analysis.BuildLuckyTargets(c, gm, metaData)

	if len(targets) == 0 {
		sb.WriteString("Brak kandydatow -- wszystkie gatunki maja juz status Lucky.\n\n")
	} else {
		for _, t := range targets {
			status := "BRAK Lucky"
			if t.HasLucky {
				status = "MASZ Lucky"
			}
			sb.WriteString(fmt.Sprintf("- **%s** (score %.0f) -- %s\n", t.Reason, t.Score, status))
		}
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

func writeSection6(sb *strings.Builder, c []model.Pokemon, gmMap map[string]*model.GamemasterPokemon, metaCats gamemaster.MetaTiersByCategory) {
	sb.WriteString("## 6. Szczegolowa analiza -- Ocena 1-10\n\n")

	// Legendary / Mythical
	sb.WriteString("### Legendary / Mythical\n\n")
	legendary := analysis.FilterBool(c, func(p model.Pokemon) bool {
		bestTier := metaCats.BestMetaTier(p.SpeciesID)
		return gamemaster.TierRank(bestTier) >= gamemaster.TierRank("A") && (p.BaseStats.Atk+p.BaseStats.Def+p.BaseStats.HP) >= 600
	})
	if len(legendary) == 0 {
		sb.WriteString("Brak legendarnych w kolekcji.\n\n")
	} else {
		sorted := analysis.SortByScore(legendary)
		for _, p := range sorted {
			rating := analysis.RatingFromScore(p.DevScore)
			sb.WriteString(fmt.Sprintf("- **%s** -- %s | Ocena: %d/10, IV %.1f%%, Lucky: %s, Rekomendacja: %s\n",
				p.Species, tierContext(p), rating, p.IVPct, analysis.BoolStr(p.IsLucky), analysis.Recommendation(p, rating)))
		}
	}

	// HUNDO
	sb.WriteString("\n### HUNDO\n\n")
	hundo := analysis.FilterBool(c, func(p model.Pokemon) bool { return p.IVPct == 100.0 })
	if len(hundo) == 0 {
		sb.WriteString("Brak HUNDO.\n\n")
	} else {
		sorted := analysis.SortByCP(hundo)
		for _, p := range sorted {
			sb.WriteString(fmt.Sprintf("- **%s** -- CP %d, Lucky: %s, Shadow: %s, RaidTier: %s, Rekomendacja: MAX OUT (L50)\n",
				p.Species, p.CP, analysis.BoolStr(p.IsLucky), analysis.BoolStr(p.IsShadow), analysis.StrOrDash(p.RaidTier)))
		}
	}
	sb.WriteString("\n")

	// Shadow
	sb.WriteString("### Shadow -- Frustration -> Return\n\n")
	shadow := analysis.FilterBool(c, func(p model.Pokemon) bool { return p.IsShadow })
	if len(shadow) == 0 {
		sb.WriteString("Brak Shadow.\n\n")
	} else {
		sorted := analysis.SortByScore(shadow)
		for _, p := range sorted {
			rating := analysis.RatingFromScore(p.DevScore)
			sb.WriteString(fmt.Sprintf("- **%s** -- IV %.1f%%, Lucky: %s, %s, Ocena: %d/10, Rekomendacja: %s\n",
				p.Species, p.IVPct, analysis.BoolStr(p.IsLucky), tierContext(p), rating, analysis.Recommendation(p, rating)))
		}
	}
	sb.WriteString("\n")

	// PvP Relevant
	sb.WriteString("### PvP Relevant\n\n")
	pvp := analysis.FilterBool(c, func(p model.Pokemon) bool {
		purpose := strings.ToLower(p.Purpose)
		return purpose == "pvp" || purpose == "gl" || purpose == "ul" || purpose == "ml" ||
			gamemaster.TierRank(p.GLTier) >= gamemaster.TierRank("A") || gamemaster.TierRank(p.ULTier) >= gamemaster.TierRank("A") || gamemaster.TierRank(p.MLTier) >= gamemaster.TierRank("A")
	})
	if len(pvp) == 0 {
		sb.WriteString("Brak PvP-relevantnych.\n\n")
	} else {
		sorted := analysis.SortByScore(pvp)
		for _, p := range analysis.TopN(sorted, 30) {
			rating := analysis.RatingFromScore(p.DevScore)
			sb.WriteString(fmt.Sprintf("- **%s** -- IV %.1f%%, CP %d, Lucky: %s, %s, Moves: %s/%s/%s, Ocena: %d/10\n",
				p.Species, p.IVPct, p.CP, analysis.BoolStr(p.IsLucky), tierContext(p),
				analysis.StrOrDash(p.FastMove), analysis.StrOrDash(p.Charged1), analysis.StrOrDash(p.Charged2), rating))
		}
	}
	sb.WriteString("\n")

	// Shiny
	sb.WriteString("### Shiny\n\n")
	shiny := analysis.FilterBool(c, func(p model.Pokemon) bool { return p.IsShiny })
	if len(shiny) == 0 {
		sb.WriteString("Brak Shiny.\n\n")
	} else {
		sorted := analysis.SortByScore(shiny)
		for _, p := range sorted {
			rating := analysis.RatingFromScore(p.DevScore)
			starMeta := ""
			bestTier := metaCats.BestMetaTier(p.SpeciesID)
			if gamemaster.TierRank(bestTier) >= gamemaster.TierRank("A") {
				starMeta = " [META]"
			}
			sb.WriteString(fmt.Sprintf("- **%s** -- IV %.1f%%, Lucky: %s%s, %s, Ocena: %d/10, Rekomendacja: %s\n",
				p.Species, p.IVPct, analysis.BoolStr(p.IsLucky), starMeta, tierContext(p), rating, analysis.Recommendation(p, rating)))
		}
	}
	sb.WriteString("\n")

	// Inne
	sb.WriteString("### Inne (Dynamax / Purified / Traded)\n\n")
	other := analysis.FilterBool(c, func(p model.Pokemon) bool {
		return p.IsDynamax || (p.IsPurified && p.IVPct >= 80) || (p.IsTraded && p.IVPct >= 90)
	})
	if len(other) == 0 {
		sb.WriteString("Brak.\n\n")
	} else {
		sorted := analysis.SortByScore(other)
		for _, p := range sorted {
			rating := analysis.RatingFromScore(p.DevScore)
			sb.WriteString(fmt.Sprintf("- **%s** -- IV %.1f%%, Tags: %s, %s, Ocena: %d/10, Rekomendacja: %s\n",
				p.Species, p.IVPct, analysis.FormatTags(p), tierContext(p), rating, analysis.Recommendation(p, rating)))
		}
	}
	sb.WriteString("\n---\n\n")
}

func writeSection7(sb *strings.Builder, c []model.Pokemon) {
	sb.WriteString(fmt.Sprintf("## 7. Wszystkie Pokemon-y -- Ranking Rozwoju (%d sztuk)\n\n", len(c)))

	ranking := analysis.BuildFullRanking(c)

	sb.WriteString("| # | UID | Pokemon | Score | IV% | Lucky | Shadow | Tier | CP | Liga | Moves | Przeznaczenie |\n")
	sb.WriteString("|--:|-----|---------|------:|----:|:-----:|:------:|:----:|---:|:-----|:------|:-------------|\n")

	for _, row := range ranking {
		p := row.Pokemon
		lucky := analysis.FormatLucky(p)
		shadow := analysis.FormatShadow(p)
		tier := analysis.TierOrDash(p)
		moveStr := fmt.Sprintf("%s/%s/%s", analysis.StrOrDash(p.FastMove), analysis.StrOrDash(p.Charged1), analysis.StrOrDash(p.Charged2))
		if len(moveStr) > 30 {
			moveStr = moveStr[:27] + "..."
		}
		sb.WriteString(fmt.Sprintf("| %d | %s | **%s** %s | %.1f | %.1f%% | %s | %s | %s | %4d | %s | %s | %s |\n",
			row.Rank, p.UID, p.Species, analysis.FormatEmojis(p),
			p.DevScore, p.IVPct, lucky, shadow, tier,
			p.CP, "-", moveStr, analysis.PurposeLabel(p)))
	}
	sb.WriteString("\n")
}

// --- Helpers ---

func tierContext(p model.Pokemon) string {
	return fmt.Sprintf("R:%s GL:%s UL:%s ML:%s", analysis.StrOrDash(p.RaidTier), analysis.StrOrDash(p.GLTier), analysis.StrOrDash(p.ULTier), analysis.StrOrDash(p.MLTier))
}
