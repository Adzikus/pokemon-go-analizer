package commands

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/adzik/pokemon/model"
)

// Type effectiveness: super effective vs type
var SuperEffective = map[string][]string{
	"normal":   {},
	"fire":     {"grass", "ice", "bug", "steel"},
	"water":    {"fire", "ground", "rock"},
	"electric": {"water", "flying"},
	"grass":    {"water", "ground", "rock"},
	"ice":      {"grass", "ground", "flying", "dragon"},
	"fighting": {"normal", "ice", "rock", "dark", "steel"},
	"poison":   {"grass", "fairy"},
	"ground":   {"fire", "electric", "poison", "rock", "steel"},
	"flying":   {"grass", "fighting", "bug"},
	"psychic":  {"fighting", "poison"},
	"bug":      {"grass", "psychic", "dark"},
	"rock":     {"fire", "ice", "flying", "bug"},
	"ghost":    {"psychic", "ghost"},
	"dragon":   {"dragon"},
	"dark":     {"psychic", "ghost"},
	"steel":    {"ice", "rock", "fairy"},
	"fairy":    {"fighting", "dragon", "dark"},
}

// Boss types for raid recommendations
var BossTypes = []string{
	"normal", "fire", "water", "electric", "grass", "ice",
	"fighting", "poison", "ground", "flying", "psychic",
	"bug", "rock", "ghost", "dragon", "dark", "steel", "fairy",
}

// League caps
var LeagueCaps = map[string]int{
	"GL": 1500,
	"UL": 2500,
	"ML": 0, // unlimited
}

// CollectionStats holds summary statistics.
type CollectionStats struct {
	Total       int
	Shiny       int
	Lucky       int
	Shadow      int
	Purified    int
	Dynamax     int
	Traded      int
	Hundo       int
	HighIV      int // >=90%
	Families    int
	TierS       int
	TierA       int
	TierB       int
	TierC       int
	PurposeRaid int
	PurposePvP  int
	PurposeColl int
	PurposeOther int
}

func ComputeStats(c []model.Pokemon) CollectionStats {
	var s CollectionStats
	s.Total = len(c)
	seenFamilies := map[string]bool{}

	for _, p := range c {
		if p.IsShiny {
			s.Shiny++
		}
		if p.IsLucky {
			s.Lucky++
		}
		if p.IsShadow {
			s.Shadow++
		}
		if p.IsPurified {
			s.Purified++
		}
		if p.IsDynamax {
			s.Dynamax++
		}
		if p.IsTraded {
			s.Traded++
		}
		if p.IVPct == 100.0 {
			s.Hundo++
		}
		if p.IVPct >= 90 {
			s.HighIV++
		}
		if p.FamilyID != "" {
			seenFamilies[p.FamilyID] = true
		}
		switch p.MetaTier {
		case "S":
			s.TierS++
		case "A":
			s.TierA++
		case "B":
			s.TierB++
		case "C":
			s.TierC++
		}
		purpose := strings.ToLower(p.Purpose)
		if purpose == "rajdy" || purpose == "raid" {
			s.PurposeRaid++
		} else if purpose == "pvp" || purpose == "gl" || purpose == "ul" || purpose == "ml" {
			s.PurposePvP++
		} else if purpose == "kolekcja" {
			s.PurposeColl++
		} else {
			s.PurposeOther++
		}
	}
	s.Families = len(seenFamilies)
	return s
}

func TopN(c []model.Pokemon, n int) []model.Pokemon {
	if len(c) <= n {
		return c
	}
	return c[:n]
}

func FilterIV90Plus(c []model.Pokemon) []model.Pokemon {
	var r []model.Pokemon
	for _, p := range c {
		if p.IVPct >= 90 {
			r = append(r, p)
		}
	}
	return r
}

// --- Section 2: Raid Tier Lists ---

type RaidTierEntry struct {
	Pokemon model.Pokemon
	Score   float64 // CP * IV% * meta_bonus
}

func BuildRaidTierList(c []model.Pokemon, gmMap map[string]*model.GamemasterPokemon) map[string][]RaidTierEntry {
	tiers := map[string][]RaidTierEntry{
		"S": {}, "A": {}, "B": {}, "C": {},
	}
	for _, p := range c {
		tier := p.MetaTier
		if tier == "" {
			tier = "C"
		}
		score := float64(p.CP) * (p.IVPct / 100)
		if p.IsLucky {
			score *= 1.2
		}
		entry := RaidTierEntry{Pokemon: p, Score: score}
		tiers[tier] = append(tiers[tier], entry)
	}
	// Sort each tier by score DESC
	for tier := range tiers {
		sort.Slice(tiers[tier], func(i, j int) bool {
			return tiers[tier][i].Score > tiers[tier][j].Score
		})
	}
	return tiers
}

func BuildRaidByType(c []model.Pokemon, bossType string) []model.Pokemon {
	effective := SuperEffective[bossType]
	if len(effective) == 0 {
		return nil
	}
	var candidates []model.Pokemon
	for _, p := range c {
		if p.CP < 1000 {
			continue
		}
		for _, t := range effective {
			if strings.ToLower(p.Type1) == t || strings.ToLower(p.Type2) == t {
				candidates = append(candidates, p)
				break
			}
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].CP != candidates[j].CP {
			return candidates[i].CP > candidates[j].CP
		}
		return candidates[i].IVPct > candidates[j].IVPct
	})
	return TopN(candidates, 5)
}

// --- Section 3: PvP Teams ---

type PvPTeam struct {
	Name    string
	Style   string
	Members [3]model.Pokemon
}

func BuildPvPTeams(c []model.Pokemon, league string) []PvPTeam {
	cap := LeagueCaps[league]
	var candidates []model.Pokemon
	for _, p := range c {
		if cap > 0 && p.CP > cap {
			continue
		}
		if cap > 0 && p.CP < cap-200 {
			continue
		}
		if p.MetaTier == "S" || p.MetaTier == "A" || p.MetaTier == "B" {
			candidates = append(candidates, p)
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].DevScore > candidates[j].DevScore
	})

	var teams []PvPTeam
	for i := 0; i+2 < len(candidates) && len(teams) < 3; i += 3 {
		team := PvPTeam{
			Name:    fmt.Sprintf("Druzyna %c", 'A'+len(teams)),
			Style:   "balanced",
			Members: [3]model.Pokemon{candidates[i], candidates[i+1], candidates[i+2]},
		}
		teams = append(teams, team)
	}
	return teams
}

// --- Section 4: Transfer Candidates ---

type TransferStatus string

const (
	Safe     TransferStatus = "SAFE"
	Keep     TransferStatus = "KEEP"
	Trade    TransferStatus = "TRADE"
	SoftKeep TransferStatus = "SOFT_KEEP"
)

type TransferResult struct {
	Pokemon model.Pokemon
	Status  TransferStatus
	Reason  string
}

func BuildTransferCandidates(c []model.Pokemon, gmMap map[string]*model.GamemasterPokemon) []TransferResult {
	// Build family groups
	families := map[string][]model.Pokemon{}
	for _, p := range c {
		fid := p.FamilyID
		if fid == "" {
			fid = "__unknown__" + p.SpeciesID
		}
		families[fid] = append(families[fid], p)
	}

	var results []TransferResult
	for _, p := range c {
		status, reason := evaluateTransfer(p, families, gmMap)
		results = append(results, TransferResult{Pokemon: p, Status: status, Reason: reason})
	}

	// Sort: SAFE (CP DESC) then TRADE then KEEP
	sort.Slice(results, func(i, j int) bool {
	 order := map[TransferStatus]int{Safe: 0, Trade: 1, SoftKeep: 2, Keep: 3}
		if order[results[i].Status] != order[results[j].Status] {
			return order[results[i].Status] < order[results[j].Status]
		}
		return results[i].Pokemon.CP > results[j].Pokemon.CP
	})
	return results
}

func evaluateTransfer(p model.Pokemon, families map[string][]model.Pokemon, gmMap map[string]*model.GamemasterPokemon) (TransferStatus, string) {
	fid := p.FamilyID
	if fid == "" {
		fid = "__unknown__" + p.SpeciesID
	}
	family := families[fid]

	// Filter 1: Shiny
	if p.IsShiny {
		betterShiny := false
		for _, fp := range family {
			if fp.IsShiny && fp.IVPct > p.IVPct {
				betterShiny = true
				break
			}
		}
		if betterShiny {
			if p.IsTraded {
				return Trade, "Lepszy shiny w rodzinie, juz tradeowany"
			}
			return Trade, "Lepszy shiny w rodzinie"
		}
		return Keep, "Jedyny shiny w rodzinie"
	}

	// Filter 2: IV% + CP
	if len(family) == 1 {
		return Keep, "Jedyny w rodzinie"
	}
	if p.IVPct == 100 {
		return Keep, "HUNDO"
	}
	// Find best in family
	best := p
	for _, fp := range family {
		if fp.IVPct > best.IVPct || (fp.IVPct == best.IVPct && fp.CP > best.CP) {
			best = fp
		}
	}
	if best.UID != p.UID {
		cpDiff := float64(best.CP-p.CP) / float64(best.CP) * 100
		if cpDiff > 50 {
			return Trade, fmt.Sprintf("CP difference %.0f%% vs best in family", cpDiff)
		}
		if best.IVPct > p.IVPct {
			return Trade, "Nizsze IV niz najlepszy w rodzinie"
		}
	}

	// Filter 3: Meta + Quantity
	if p.MetaTier == "S" || p.MetaTier == "A" {
		count := 0
		for _, fp := range family {
			if fp.SpeciesID == p.SpeciesID {
				count++
			}
		}
		if count >= 10 {
			return Trade, fmt.Sprintf(" %d sztuk tego gatunku, wiec najs slabiej", count)
		}
		return Keep, fmt.Sprintf("Meta %s, %d szt.", p.MetaTier, count)
	}

	// Filter 4: Purpose
	if strings.ToLower(p.Purpose) == "kolekcja" {
		return SoftKeep, "Element oznaczony do kolekcji - usuwac w ostatecznosci"
	}

	return Safe, "Brak wartosci bojowej/kolekcjonerskiej"
}

// --- Section 5: Lucky Trade Targets ---

type LuckyTarget struct {
	Pokemon  model.Pokemon
	Priority int
	HasLucky bool
	Reason   string
}

func BuildLuckyTargets(c []model.Pokemon, gmData *GamemasterData, metaData *MetaData) []LuckyTarget {
	// Build lucky species set
	luckySpecies := map[string]bool{}
	for _, p := range c {
		if p.IsLucky {
			luckySpecies[p.SpeciesID] = true
		}
	}

	// Meta S (top 10 raid) and A (10-19)
	metaS := map[string]bool{}
	metaA := map[string]bool{}
	for i, sid := range metaData.Rankings.Raid {
		if i < 10 {
			metaS[sid] = true
		} else if i < 20 {
			metaA[sid] = true
		}
	}

	var targets []LuckyTarget

	// Priority 1: Legendary/Mythical + Meta S
	for _, gm := range gmData.Pokemon {
		isLegend := false
		for _, tag := range gm.Tags {
			if tag == "legendary" || tag == "mythical" {
				isLegend = true
				break
			}
		}
		if !isLegend {
			continue
		}
		if !metaS[gm.SpeciesID] {
			continue
		}
		reason := fmt.Sprintf("%s -- Meta S raid", gm.SpeciesName)
		targets = append(targets, LuckyTarget{
			Priority: 1,
			HasLucky: luckySpecies[gm.SpeciesID],
			Reason:   reason,
		})
	}

	// Priority 2: Shadow-eligible + Meta S/A
	for _, gm := range gmData.Pokemon {
		isShadowElig := false
		for _, tag := range gm.Tags {
			if tag == "shadoweligible" {
				isShadowElig = true
				break
			}
		}
		if !isShadowElig {
			continue
		}
		if !metaS[gm.SpeciesID] && !metaA[gm.SpeciesID] {
			continue
		}
		tier := "A"
		if metaS[gm.SpeciesID] {
			tier = "S"
		}
		reason := fmt.Sprintf("%s — Shadow-eligible, Meta %s", gm.SpeciesName, tier)
		targets = append(targets, LuckyTarget{
			Priority: 2,
			HasLucky: luckySpecies[gm.SpeciesID],
			Reason:   reason,
		})
	}

	return targets
}

// --- Section 6: Detailed Analysis Rating ---

func RatingFromScore(devScore float64) int {
	r := int(devScore / 10)
	if r < 1 {
		r = 1
	}
	if r > 10 {
		r = 10
	}
	return r
}

func Recommendation(p model.Pokemon, rating int) string {
	if p.IVPct == 100 {
		return "MAX OUT (L50)"
	}
	if p.MetaTier == "S" && rating >= 7 {
		return "Inwestuj — meta-relevantny"
	}
	if p.IsShadow && rating >= 6 {
		return "Usun Frustration po eventach, potem inwestuj"
	}
	if p.IsLucky && rating >= 5 {
		return "Lucky — tanie ulepszanie, warto inwestowac"
	}
	if rating >= 5 {
		return "Dobra opcja do inwestycji"
	}
	if rating >= 3 {
		return "Trzymaj, ale nie inwestuj teraz"
	}
	return "Kandydat do transferu"
}

// --- Section 7: Full Ranking Table ---

type RankingRow struct {
	Rank    int
	Pokemon model.Pokemon
}

func BuildFullRanking(c []model.Pokemon) []RankingRow {
	sorted := SortByScore(c)
	rows := make([]RankingRow, len(sorted))
	for i, p := range sorted {
		rows[i] = RankingRow{Rank: i + 1, Pokemon: p}
	}
	return rows
}

// Helper: unique count of species
func UniqueSpecies(c []model.Pokemon) int {
	seen := map[string]bool{}
	for _, p := range c {
		seen[p.SpeciesID] = true
	}
	return len(seen)
}

// MaxCPForLeague returns the max CP a Pokemon can have for a league.
func MaxCPForLeague(p model.Pokemon, league string) int {
	cap := LeagueCaps[league]
	if cap == 0 {
		return p.CP // ML: unlimited
	}
	return cap
}

func FormatLucky(p model.Pokemon) string {
	if p.IsLucky {
		return "OK"
	}
	return ""
}

func FormatShadow(p model.Pokemon) string {
	if p.IsShadow {
		return "👻"
	}
	return ""
}

func FormatEmojis(p model.Pokemon) string {
	var tags []string
	if p.IsShiny {
		tags = append(tags, "✨")
	}
	if p.IsLucky {
		tags = append(tags, "🍀")
	}
	if p.IsShadow {
		tags = append(tags, "👻")
	}
	if p.IsDynamax {
		tags = append(tags, "⭐")
	}
	return strings.Join(tags, "")
}

func purposeLabel(p model.Pokemon) string {
	purpose := p.Purpose
	if purpose == "" {
		purpose = "-"
	}
	return purpose
}

// safeMin returns the smaller of two ints
func safeMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// UniqueFamilies returns the number of unique family IDs
func UniqueFamilies(c []model.Pokemon) int {
	seen := map[string]bool{}
	for _, p := range c {
		if p.FamilyID != "" {
			seen[p.FamilyID] = true
		}
	}
	return len(seen)
}

func minFloat(a, b float64) float64 {
	return math.Min(a, b)
}

// MetaCompareRow holds comparison data for one meta raid attacker.
type MetaCompareRow struct {
	Name     string  // display name from gamemaster
	Position int     // position in raid_attackers (0-indexed)
	Tier     string  // S (0-9) or A (10-19)
	Owned    bool    // player has this Pokemon
	Copies   int     // how many copies
	BestIV   float64 // best IV% among copies
	IsLucky  bool    // best copy is lucky
	BestCP   int     // best CP among copies
	Score    float64 // best dev_score among copies
}

// BuildMetaComparison compares collection against meta raid attackers.
func BuildMetaComparison(c []model.Pokemon, raidAttackers []string) []MetaCompareRow {
	// Build collection index by SpeciesID
	bySpecies := map[string][]model.Pokemon{}
	for _, p := range c {
		bySpecies[p.SpeciesID] = append(bySpecies[p.SpeciesID], p)
	}

	var rows []MetaCompareRow
	limit := len(raidAttackers)
	if limit > 20 {
		limit = 20
	}

	for i := 0; i < limit; i++ {
		sid := raidAttackers[i]
		tier := "A"
		if i < 10 {
			tier = "S"
		}

		// Get display name from gamemaster (first char upper)
		name := sid

		copies := bySpecies[sid]
		row := MetaCompareRow{
			Name:     name,
			Position: i + 1,
			Tier:     tier,
			Owned:    len(copies) > 0,
			Copies:   len(copies),
		}

		// Find best copy
		for _, p := range copies {
			if p.IVPct > row.BestIV || (p.IVPct == row.BestIV && p.CP > row.BestCP) {
				row.BestIV = p.IVPct
				row.BestCP = p.CP
				row.IsLucky = p.IsLucky
				row.Score = p.DevScore
				row.Name = p.Species // use display name from collection
			}
		}

		rows = append(rows, row)
	}
	return rows
}

// FormatMetaComparison generates Markdown table + coverage summary.
func FormatMetaComparison(rows []MetaCompareRow) string {
	var sb strings.Builder

	sb.WriteString("### Meta vs Collection (Raid Attackers TOP 20)\n\n")
	sb.WriteString("| # | Pokemon | Tier | Masz? | Ile | Najl. IV | Lucky | Score |\n")
	sb.WriteString("|--:|---------|:----:|:-----:|----:|---------:|:-----:|------:|\n")

	owned := 0
	var missing []string
	for _, r := range rows {
		have := "✗"
		iv := "-"
		lucky := "-"
		score := 0.0
		if r.Owned {
			have = "✓"
			iv = fmt.Sprintf("%.1f%%", r.BestIV)
			score = r.Score
			if r.IsLucky {
				lucky = "TAK"
			} else {
				lucky = "NIE"
			}
			owned++
		} else {
			missing = append(missing, r.Name)
		}
		sb.WriteString(fmt.Sprintf("| %d | %s | %s | %s | %d | %s | %s | %.1f |\n",
			r.Position, r.Name, r.Tier, have, r.Copies, iv, lucky, score))
	}
	sb.WriteString("\n")

	pct := float64(owned) / float64(len(rows)) * 100
	sb.WriteString(fmt.Sprintf("**Meta Coverage:** %d/%d (%.0f%%)", owned, len(rows), pct))
	if len(missing) > 0 {
		sb.WriteString(fmt.Sprintf(" — brak: %s", strings.Join(missing, ", ")))
	}
	sb.WriteString("\n\n")

	return sb.String()
}
