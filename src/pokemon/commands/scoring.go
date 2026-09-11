package commands

import (
	"math"
	"strings"

	"github.com/adzik/pokemon/model"
)

// CPM_TABLE: level 1-50 in 0.5 increments, index 0 = L1, index 99 = L50
var CPM_TABLE = []float64{
	0.094000000, 0.106000000, 0.116000000, 0.126000000, 0.136000000, 0.146000000,
	0.156000000, 0.166000000, 0.176000000, 0.186000000, 0.196000000, 0.206000000,
	0.216000000, 0.226000000, 0.236000000, 0.246000000, 0.256000000, 0.266000000,
	0.276000000, 0.286000000, 0.296000000, 0.306000000, 0.316000000, 0.326000000,
	0.336000000, 0.346000000, 0.356000000, 0.366000000, 0.376000000, 0.386000000,
	0.396000000, 0.406000000, 0.416000000, 0.426000000, 0.436000000, 0.446000000,
	0.456000000, 0.466000000, 0.476000000, 0.486000000, 0.496000000, 0.506000000,
	0.516000000, 0.526000000, 0.536000000, 0.546000000, 0.556000000, 0.566000000,
	0.576000000, 0.586000000, 0.596000000, 0.606000000, 0.616000000, 0.626000000,
	0.636000000, 0.646000000, 0.656000000, 0.666000000, 0.676000000, 0.686000000,
	0.696000000, 0.706000000, 0.716000000, 0.726000000, 0.736000000, 0.746000000,
	0.756000000, 0.766000000, 0.776000000, 0.786000000, 0.796000000, 0.806000000,
	0.816000000, 0.826000000, 0.836000000, 0.846000000, 0.856000000, 0.866000000,
	0.876000000, 0.886000000, 0.896000000, 0.906000000, 0.916000000, 0.926000000,
	0.936000000, 0.946000000, 0.956000000, 0.966000000, 0.976000000, 0.986000000,
	0.996000000, 1.006000000, 1.016000000, 1.026000000, 1.036000000, 1.046000000,
	1.056000000, 1.066000000, 1.076000000, 1.086000000, 1.096000000, 1.106000000,
}

// CPM for level 50 (index 99)
const CPM_L50 = 0.971700003

// CPM for key league caps
const (
	CPM_L15 = 0.375999998 // Great League cap
	CPM_L20 = 0.595999998 // Raid/Trade level
	CPM_L25 = 0.675999998 // Weather boosted
	CPM_L40 = 0.791700003 // Old max
)

func GetCPM(level float64) float64 {
	idx := max(int((level - 1) * 2), 0)

	if idx >= len(CPM_TABLE)-1 {
		return CPM_TABLE[len(CPM_TABLE)-1]
	}

	frac := (level-1)*2 - float64(idx)
	
	if frac < 0 {
		frac = 0
	}
	
	if frac > 1 {
		frac = 1
	}
	return CPM_TABLE[idx]*(1-frac) + CPM_TABLE[idx+1]*frac
}

// CalculateCP computes CP at a given level using the standard formula.
func CalculateCP(baseAtk, baseDef, baseHP, ivAtk, ivDef, ivHP int, level float64) int {
	atk := float64(baseAtk + ivAtk)
	def := float64(baseDef + ivDef)
	hp := float64(baseHP + ivHP)
	cpm := GetCPM(level)
	return int((atk * math.Sqrt(def) * math.Sqrt(hp) * cpm * cpm) / 10)
}

// CalculateMaxCP returns CP at level 50.
func CalculateMaxCP(baseAtk, baseDef, baseHP, ivAtk, ivDef, ivHP int) int {
	return CalculateCP(baseAtk, baseDef, baseHP, ivAtk, ivDef, ivHP, 50)
}

// ScoreResult holds the scoring breakdown for a Pokemon.
type ScoreResult struct {
	Score      float64
	IVScore    float64 // 0-15
	LuckyPts   float64 // 0-20
	MetaPts    float64 // 0-20
	MovePts    float64 // 0-15
	MegaPts    float64 // 0-10
	CPPts      float64 // 0-10
	PurposePts float64 // 0-10
	Notes      []string
}

// CalculateDevScore computes the 0-100 development score for a Pokemon.
// Requires gamemaster map for moveset and mega/shadow checks.
func CalculateDevScore(p model.Pokemon, gmMap map[string]*model.GamemasterPokemon, metaTiers map[string]string) ScoreResult {
	var r ScoreResult

	// Phase 1: Early exclusion
	if p.Purpose == "Transfer" || p.Purpose == "Kolekcja" {
		r.Notes = append(r.Notes, "Transfer/Kolekcja — nie inwestuj")
		return r
	}

	// IV% (0-15)
	r.IVScore = math.Min(p.IVPct/100*15, 15)

	// Lucky (0-20)
	if p.IsLucky {
		r.LuckyPts = 20
	}

	// Meta Tier (0-20)
	switch p.MetaTier {
	case "S":
		r.MetaPts = 20
	case "A":
		r.MetaPts = 15
	case "B":
		r.MetaPts = 8
	}

	// Moveset (0-15): +5 per meta move currently equipped
	gmPokemon := gmMap[p.SpeciesID]
	if gmPokemon != nil {
		moveScore := 0
		// Check fast move
		if p.FastMove != "" && isMoveAvailable(p.FastMove, gmPokemon.FastMoves) {
			moveScore += 5
		} else if p.FastMove != "" {
			// Not in available moves — suggest best
			if len(gmPokemon.FastMoves) > 0 {
				r.Notes = append(r.Notes, "Zmien szybki na: "+gmPokemon.FastMoves[0])
			}
		}
		// Check charged 1
		if p.Charged1 != "" && isMoveAvailable(p.Charged1, gmPokemon.ChargedMoves) {
			moveScore += 5
		} else if p.Charged1 != "" {
			if len(gmPokemon.ChargedMoves) > 0 {
				r.Notes = append(r.Notes, "Zmien charged 1 na: "+gmPokemon.ChargedMoves[0])
			}
		}
		// Check charged 2
		if p.Charged2 != "" && isMoveAvailable(p.Charged2, gmPokemon.ChargedMoves) {
			moveScore += 5
		} else if p.Charged2 != "" {
			if len(gmPokemon.ChargedMoves) > 1 {
				r.Notes = append(r.Notes, "Zmien charged 2 na: "+gmPokemon.ChargedMoves[1])
			}
		}
		r.MovePts = float64(moveScore)
	}

	// Mega/Shadow Potential (0-10)
	if gmPokemon != nil && gmPokemon.Family != nil {
		familyID := gmPokemon.Family.ID
		// Check if mega form exists in family and is tier S/A
		for _, sid := range getFamilyMembers(familyID, gmMap) {
			if strings.Contains(sid, "_mega") {
				tier := metaTiers[sid]
				if (tier == "S" || tier == "A") && !p.IsDynamax {
					r.MegaPts = 5
					break
				}
			}
		}
		// Check if shadow form exists and is tier S
		if !p.IsShadow && !p.IsPurified {
			shadowID := p.SpeciesID + "_shadow"
			if metaTiers[shadowID] == "S" {
				r.MegaPts += 5
			}
		}
	}

	// CP Investment (0-10)
	if p.BaseStats.Atk > 0 && p.BaseStats.Def > 0 && p.BaseStats.HP > 0 {
		maxCP := CalculateMaxCP(p.BaseStats.Atk, p.BaseStats.Def, p.BaseStats.HP, 15, 15, 15)
		if maxCP > 0 {
			r.CPPts = math.Min(float64(p.CP)/float64(maxCP)*10, 10)
		}
	}

	// Purpose Bonus (0-10)
	purposeLower := strings.ToLower(p.Purpose)
	switch purposeLower {
	case "rajdy", "raid", "pvp", "gl", "ul", "ml":
		r.PurposePts = 8
	case "kolekcja":
		r.PurposePts = 3
	}

	// Total
	r.Score = r.IVScore + r.LuckyPts + r.MetaPts + r.MovePts + r.MegaPts + r.CPPts + r.PurposePts
	r.Score = math.Max(0, math.Min(r.Score, 100))

	return r
}

func isMoveAvailable(move string, available []string) bool {
	moveUpper := strings.ToUpper(strings.ReplaceAll(move, " ", "_"))
	for _, m := range available {
		if strings.ToUpper(m) == moveUpper {
			return true
		}
	}
	return false
}

func getFamilyMembers(familyID string, gmMap map[string]*model.GamemasterPokemon) []string {
	var members []string
	for sid, gm := range gmMap {
		if gm.Family != nil && gm.Family.ID == familyID {
			members = append(members, sid)
		}
	}
	return members
}

// SortByScore sorts collection by dev_score DESC, then IV% DESC, then CP DESC.
func SortByScore(c []model.Pokemon) []model.Pokemon {
	sorted := make([]model.Pokemon, len(c))
	copy(sorted, c)
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].DevScore > sorted[i].DevScore ||
				(sorted[j].DevScore == sorted[i].DevScore && sorted[j].IVPct > sorted[i].IVPct) ||
				(sorted[j].DevScore == sorted[i].DevScore && sorted[j].IVPct == sorted[i].IVPct && sorted[j].CP > sorted[i].CP) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	return sorted
}

// EnrichWithScores computes dev_score for all Pokemon in collection.
func EnrichWithScores(collection []model.Pokemon, gmMap map[string]*model.GamemasterPokemon, metaTiers map[string]string) []model.Pokemon {
	for i := range collection {
		result := CalculateDevScore(collection[i], gmMap, metaTiers)
		collection[i].DevScore = result.Score
		collection[i].MaxCP = CalculateMaxCP(
			collection[i].BaseStats.Atk, collection[i].BaseStats.Def, collection[i].BaseStats.HP,
			collection[i].IVAtk, collection[i].IVDef, collection[i].IVHP,
		)
	}
	return collection
}
