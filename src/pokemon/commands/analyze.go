package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/adzik/pokemon/config"
	"github.com/adzik/pokemon/model"
	"github.com/adzik/pokemon/ods"
	"github.com/adzik/pokemon/pkg/analysis"
	"github.com/adzik/pokemon/pkg/gamemaster"
)

type ReportData struct {
	Konto string
	Date  string

	Collection []model.Pokemon
	Stats      analysis.CollectionStats

	TopHundo    []model.Pokemon
	TopCP       []model.Pokemon
	TopDynamax  []model.Pokemon
	TopShadow   []model.Pokemon
	TopShiny    []model.Pokemon
	TopPowerUp  []model.Pokemon
	TopDevScore []model.Pokemon

	Contexts []ContextData

	MetaRows    []analysis.MetaCompareRow
	MetaOwned   int
	MetaTotal   int
	MetaMissing []string

	RaidTypes         []RaidTypeData
	UniversalCounters []analysis.RaidDPSResult

	Leagues []LeagueData

	PvPSummary []model.Pokemon

	SafeTransfer     []analysis.TransferResult
	TradeTransfer    []analysis.TransferResult
	SoftKeepTransfer []analysis.TransferResult

	LuckyTargets []analysis.LuckyTarget

	Legendary    []model.Pokemon
	HundoDetail  []model.Pokemon
	ShadowDetail []model.Pokemon
	PvPDetail    []model.Pokemon
	ShinyDetail  []model.Pokemon
	MetaStars    map[string]bool
	OtherDetail  []model.Pokemon

	FullRanking []analysis.RankingRow
}

type ContextData struct {
	Name   string
	Top10  []ContextPokemon
}

type ContextPokemon struct {
	model.Pokemon
	Tier  string
	Rank  int
	Score float64
}

type RaidTypeData struct {
	Type     string
	Counters []analysis.RaidDPSResult
}

type LeagueData struct {
	Key      string
	FullName string
	Teams    []analysis.PvPTeam
	Roles    []string
}

type teamRoles struct{}

func (teamRoles) Index(i int) string {
	roles := []string{"Lead", "Switch", "Closer"}
	if i < len(roles) {
		return roles[i]
	}
	return ""
}

func RunAnalyze(args []string, projectRoot string) {
	fs := flag.NewFlagSet("analyze", flag.ExitOnError)
	konto := fs.String("konto", "", "Nazwa konta (np. Test)")
	all := fs.Bool("all", false, "Analiza dla wszystkich kont (Pokemon-GO-*.ods)")
	odsPath := fs.String("ods", "", "Sciezka do ODS (opcjonalnie)")
	fs.Parse(args)

	if *all {
		runAnalyzeAll(projectRoot)
		return
	}

	path := *odsPath
	if path == "" && *konto != "" {
		path = filepath.Join(projectRoot, fmt.Sprintf("Pokemon-GO-%s.ods", *konto))
	}
	if path == "" {
		fmt.Fprintf(os.Stderr, "Uzycie: pokemon analyze --konto <nazwa> [--ods <sciezka>]\n")
		fmt.Fprintf(os.Stderr, "       pokemon analyze --all\n")
		os.Exit(1)
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(projectRoot, path)
	}

	runAnalyzeOne(path, *konto, projectRoot)
}

func runAnalyzeAll(projectRoot string) {
	matches, err := filepath.Glob(filepath.Join(projectRoot, "Pokemon-GO-*.ods"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: Glob: %v\n", err)
		os.Exit(1)
	}
	if len(matches) == 0 {
		fmt.Fprintf(os.Stderr, "Brak plikow Pokemon-GO-*.ods w %s\n", projectRoot)
		os.Exit(1)
	}

	for _, path := range matches {
		base := filepath.Base(path)
		kontoName := extractKontoFromFilename(base)
		if kontoName == "" || kontoName == "Test" {
			continue
		}
		fmt.Printf("\n=== Analiza: %s ===\n", kontoName)
		runAnalyzeOne(path, kontoName, projectRoot)
	}
}

func runAnalyzeOne(path, kontoName, projectRoot string) {

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

	analysis.EnrichWithScores(collection, gmMap, moveMap, metaCats)

	now := time.Now()
	dateStr := now.Format("20060102")

	analizyDir := filepath.Join(projectRoot, "analizy")
	if err := os.MkdirAll(analizyDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: Cannot create analizy dir: %v\n", err)
		os.Exit(1)
	}

	jsonPath := filepath.Join(analizyDir, fmt.Sprintf("collection-%s-%s.json", kontoName, dateStr))
	writeJSONSnapshot(jsonPath, collection, kontoName)

	mdPath := filepath.Join(analizyDir, fmt.Sprintf("Analiza-%s-%s-%s.md", dateStr, config.Global.Report.ModelName, kontoName))
	writeMarkdownReport(mdPath, collection, kontoName, gmData, gmMap, moveMap, metaData, metaTiers, metaCats, projectRoot)

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

func buildReportData(
	collection []model.Pokemon,
	konto string,
	gm *gamemaster.GamemasterData,
	gmMap map[string]*model.GamemasterPokemon,
	moveMap map[string]*gamemaster.MoveEntry,
	metaData *gamemaster.MetaData,
	metaTiers map[string]string,
	metaCats gamemaster.MetaTiersByCategory,
) ReportData {
	dateStr := time.Now().Format("2006-01-02")
	stats := analysis.ComputeStats(collection)
	data := ReportData{
		Konto:      konto,
		Date:       dateStr,
		Collection: collection,
		Stats:      stats,
	}

	data.TopHundo = analysis.TopN(analysis.SortByCP(analysis.FilterBool(collection, func(p model.Pokemon) bool { return p.IVPct == 100.0 })), 10)
	data.TopCP = analysis.TopN(analysis.SortByCP(collection), 10)
	data.TopDynamax = analysis.TopN(analysis.SortByCP(analysis.FilterBool(collection, func(p model.Pokemon) bool { return p.IsDynamax })), 10)
	data.TopShadow = analysis.TopN(analysis.SortByScore(analysis.FilterBool(collection, func(p model.Pokemon) bool { return p.IsShadow })), 10)
	data.TopShiny = analysis.TopN(analysis.SortByScore(analysis.FilterBool(collection, func(p model.Pokemon) bool { return p.IsShiny })), 10)
	data.TopPowerUp = analysis.TopN(analysis.SortByScore(analysis.FilterBool(collection, func(p model.Pokemon) bool {
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
	})), 10)
	data.TopDevScore = analysis.TopN(analysis.SortByScore(collection), 10)

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
		scored := analysis.FilterBool(collection, func(p model.Pokemon) bool { return ctx.getTier(p) != "" })
		sort.Slice(scored, func(i, j int) bool {
			ri, rj := ctx.getRank(scored[i]), ctx.getRank(scored[j])
			if ri != rj {
				return ri < rj
			}
			return ctx.getScore(scored[i]) > ctx.getScore(scored[j])
		})
		cd := ContextData{Name: ctx.name}
		count := 0
		for _, p := range scored {
			tier := ctx.getTier(p)
			if tier != "S" && tier != "A+" && tier != "A" {
				continue
			}
			cd.Top10 = append(cd.Top10, ContextPokemon{Pokemon: p, Tier: tier, Rank: ctx.getRank(p), Score: ctx.getScore(p)})
			count++
			if count >= 10 {
				break
			}
		}
		data.Contexts = append(data.Contexts, cd)
	}

	if metaData != nil && len(metaData.Rankings.Raid) > 0 {
		data.MetaRows = analysis.BuildMetaComparison(collection, metaData.Rankings.Raid, gmMap)
		for _, r := range data.MetaRows {
			if r.Owned {
				data.MetaOwned++
			}
			data.MetaTotal++
			if !r.Owned {
				data.MetaMissing = append(data.MetaMissing, r.Name)
			}
		}
	}

	rankings := analysis.BuildRaidRankings(collection, gmMap, moveMap, 10)
	for _, bossType := range analysis.BossTypes {
		counters := rankings.ByType[bossType]
		if len(counters) > 0 {
			data.RaidTypes = append(data.RaidTypes, RaidTypeData{Type: bossType, Counters: counters})
		}
	}
	data.UniversalCounters = rankings.Universal

	leagueDefs := []struct{ key, name string }{
		{"GL", "Great League (1500 CP)"},
		{"UL", "Ultra League (2500 CP)"},
		{"ML", "Master League (bez limitu CP)"},
	}
	for _, ld := range leagueDefs {
		teams := analysis.BuildPvPTeams(collection, ld.key, moveMap)
		data.Leagues = append(data.Leagues, LeagueData{Key: ld.key, FullName: ld.name, Teams: teams, Roles: []string{"Lead", "Switch", "Closer"}})
	}

	pvp := analysis.FilterBool(collection, func(p model.Pokemon) bool {
		purpose := strings.ToLower(p.Purpose)
		return purpose == "pvp" || purpose == "gl" || purpose == "ul" || purpose == "ml"
	})
	if len(pvp) > 0 {
		data.PvPSummary = analysis.TopN(analysis.SortByScore(pvp), 20)
	}

	results := analysis.BuildTransferCandidates(collection, gmMap)
	for _, r := range results {
		switch r.Status {
		case analysis.Safe:
			data.SafeTransfer = append(data.SafeTransfer, r)
		case analysis.Trade:
			data.TradeTransfer = append(data.TradeTransfer, r)
		case analysis.Keep:
		case analysis.SoftKeep:
			data.SoftKeepTransfer = append(data.SoftKeepTransfer, r)
		}
	}

	data.LuckyTargets = analysis.BuildLuckyTargets(collection, gm, metaData)

	data.Legendary = analysis.SortByScore(analysis.FilterBool(collection, func(p model.Pokemon) bool {
		bestTier := metaCats.BestMetaTier(p.SpeciesID)
		return gamemaster.TierRank(bestTier) >= gamemaster.TierRank("A") && (p.BaseStats.Atk+p.BaseStats.Def+p.BaseStats.HP) >= 600
	}))

	data.HundoDetail = analysis.SortByCP(analysis.FilterBool(collection, func(p model.Pokemon) bool { return p.IVPct == 100.0 }))

	data.ShadowDetail = analysis.SortByScore(analysis.FilterBool(collection, func(p model.Pokemon) bool { return p.IsShadow }))

	data.PvPDetail = analysis.TopN(analysis.SortByScore(analysis.FilterBool(collection, func(p model.Pokemon) bool {
		purpose := strings.ToLower(p.Purpose)
		return purpose == "pvp" || purpose == "gl" || purpose == "ul" || purpose == "ml" ||
			gamemaster.TierRank(p.GLTier) >= gamemaster.TierRank("A") || gamemaster.TierRank(p.ULTier) >= gamemaster.TierRank("A") || gamemaster.TierRank(p.MLTier) >= gamemaster.TierRank("A")
	})), 30)

	data.MetaStars = make(map[string]bool)
	shiny := analysis.FilterBool(collection, func(p model.Pokemon) bool { return p.IsShiny })
	data.ShinyDetail = analysis.SortByScore(shiny)
	for _, p := range data.ShinyDetail {
		bestTier := metaCats.BestMetaTier(p.SpeciesID)
		data.MetaStars[p.SpeciesID] = gamemaster.TierRank(bestTier) >= gamemaster.TierRank("A")
	}

	data.OtherDetail = analysis.SortByScore(analysis.FilterBool(collection, func(p model.Pokemon) bool {
		return p.IsDynamax || (p.IsPurified && p.IVPct >= 80) || (p.IsTraded && p.IVPct >= 90)
	}))

	data.FullRanking = analysis.BuildFullRanking(collection)

	return data
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
	projectRoot string,
) {
	data := buildReportData(collection, konto, gm, gmMap, moveMap, metaData, metaTiers, metaCats)

	funcMap := template.FuncMap{
		"inc": func(i int) int { return i + 1 },
		"sub": func(a, b int) int { return a - b },
		"pct": func(a, b int) string { return fmt.Sprintf("%.1f", analysis.Pct(a, b)) },
		"boolStr": func(b bool) string { return analysis.BoolStr(b) },
		"strOrDash": func(s string) string { return analysis.StrOrDash(s) },
		"tierOrDash": func(p model.Pokemon) string { return analysis.TierOrDash(p) },
		"tierOrDashLeague": func(p model.Pokemon, league string) string {
			return analysis.StrOrDash(analysis.LeagueTier(p, league))
		},
		"purposeLabel": func(p model.Pokemon) string { return analysis.PurposeLabel(p) },
		"formatTags": func(p model.Pokemon) string { return analysis.FormatTags(p) },
		"formatLucky": func(p model.Pokemon) string { return analysis.FormatLucky(p) },
		"formatShadow": func(p model.Pokemon) string { return analysis.FormatShadow(p) },
		"formatEmojis": func(p model.Pokemon) string { return analysis.FormatEmojis(p) },
		"tierContext": func(p model.Pokemon) string { return tierContext(p) },
		"rating": func(score float64) int { return analysis.RatingFromScore(score) },
		"recommendation": func(p model.Pokemon) string {
			return analysis.Recommendation(p, analysis.RatingFromScore(p.DevScore))
		},
		"leagueScore": func(p model.Pokemon, league string) float64 { return leagueScore(p, league) },
		"title": func(s string) string {
			if len(s) == 0 {
				return s
			}
			return strings.ToUpper(s[:1]) + s[1:]
		},
		"join": func(elems []string, sep string) string { return strings.Join(elems, sep) },
		"moveDPS": func(p model.Pokemon) string {
			return fmt.Sprintf("%.1f / %.1f", p.MoveDPSFast, p.MoveDPSCharged)
		},
		"movesActual": func(r analysis.RaidDPSResult) string {
			s := fmt.Sprintf("%s / %s", analysis.StrOrDash(r.FastMove), analysis.StrOrDash(r.ChargedMove))
			if len(s) > config.Global.Report.MaxMoveDisplay {
				s = s[:config.Global.Report.MaxMoveDisplay-3] + "..."
			}
			return s
		},
		"movesBest": func(r analysis.RaidDPSResult) string {
			s := fmt.Sprintf("%s / %s", analysis.StrOrDash(r.BestFast), analysis.StrOrDash(r.BestCharged))
			if len(s) > config.Global.Report.MaxMoveDisplay {
				s = s[:config.Global.Report.MaxMoveDisplay-3] + "..."
			}
			return s
		},
		"formatMoves": func(p model.Pokemon) string {
			s := fmt.Sprintf("%s/%s/%s", analysis.StrOrDash(p.FastMove), analysis.StrOrDash(p.Charged1), analysis.StrOrDash(p.Charged2))
			if len(s) > 30 {
				s = s[:27] + "..."
			}
			return s
		},
	}

	templatePath := filepath.Join(projectRoot, "templates", "report.md.tmpl")
	tmplContent, err := os.ReadFile(templatePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: Read template: %v\n", err)
		os.Exit(1)
	}

	tmpl, err := template.New("report.md.tmpl").Funcs(funcMap).Parse(string(tmplContent))
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: Parse template: %v\n", err)
		os.Exit(1)
	}

	var sb strings.Builder
	if err := tmpl.Execute(&sb, data); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: Execute template: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(path, []byte(sb.String()), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: Zapisz Markdown: %v\n", err)
	}
}

func tierContext(p model.Pokemon) string {
	return fmt.Sprintf("R:%s GL:%s UL:%s ML:%s", analysis.StrOrDash(p.RaidTier), analysis.StrOrDash(p.GLTier), analysis.StrOrDash(p.ULTier), analysis.StrOrDash(p.MLTier))
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

