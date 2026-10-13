package analysis

import (
	"math"
	"strings"

	"github.com/adzik/pokemon/config"
	"github.com/adzik/pokemon/model"
	"github.com/adzik/pokemon/pkg/gamemaster"
)

// CPM_TABLE: level 1-50 in 0.5 increments, index 0 = L1, index 99 = L50
// Source: https://pokemongohub.net/post/article/pokemon-go-cpm-list/
var CPM_TABLE = []float64{
	0.094, 0.1351374318, 0.16639787, 0.192650919, 0.21573247, // L1-3
	0.2365726613, 0.25572005, 0.2735303812, 0.29024988, 0.3060573775, // L3.5-5.5
	0.3210876, 0.3354450362, 0.34921268, 0.3624577511, 0.3752356, // L6-8
	0.387592416, 0.39956728, 0.4111935514, 0.4225, 0.4329264091, // L8.5-10.5
	0.44310755, 0.4530599591, 0.4627984, 0.472336093, 0.48168495, // L11-13
	0.4908558003, 0.49985844, 0.508701765, 0.51739395, 0.5259425113, // L13.5-15.5
	0.5343543, 0.5426357375, 0.5507927, 0.5588305862, 0.5667545, // L16-18
	0.5745691333, 0.5822789, 0.5898879072, 0.5974, 0.6048236651, // L18.5-20.5
	0.6121573, 0.6194041216, 0.6265671, 0.6336491432, 0.64065295, // L21-23
	0.6475809666, 0.65443563, 0.6612192524, 0.667934, 0.6745818959, // L23.5-25.5
	0.6811649, 0.6876849038, 0.69414365, 0.70054287, 0.7068842, // L26-28
	0.7131691091, 0.7193991, 0.7255756136, 0.7317, 0.7347410093, // L28.5-30.5
	0.7377695, 0.7407855938, 0.74378943, 0.7467812109, 0.74976104, // L31-33
	0.7527290867, 0.7556855, 0.7586303683, 0.76156384, 0.7644860647, // L33.5-35.5
	0.76739717, 0.7702972656, 0.7731865, 0.7760649616, 0.77893275, // L36-38
	0.7817900548, 0.784637, 0.7874736075, 0.7903, 0.792803968, // L38.5-40.5
	0.79530001, 0.797800015, 0.8003, 0.802799995, 0.8053, // L41-43
	0.8078, 0.81029999, 0.812799985, 0.81529999, 0.81779999, // L43.5-45.5
	0.82029999, 0.82279999, 0.82529999, 0.82779999, 0.83029999, // L46-48
	0.83279999, 0.83529999, 0.83779999, 0.84029999, 0.84279999, // L48.5-50.5
}

func GetCPM(level float64) float64 {
	idx := max(int((level-1)*2), 0)
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

func CalculateCP(baseAtk, baseDef, baseHP, ivAtk, ivDef, ivHP int, level float64) int {
	atk := float64(baseAtk + ivAtk)
	def := float64(baseDef + ivDef)
	hp := float64(baseHP + ivHP)
	cpm := GetCPM(level)
	return int((atk * math.Sqrt(def) * math.Sqrt(hp) * cpm * cpm) / 10)
}

func CalculateMaxCP(baseAtk, baseDef, baseHP, ivAtk, ivDef, ivHP int) int {
	return CalculateCP(baseAtk, baseDef, baseHP, ivAtk, ivDef, ivHP, 50)
}

// CalculateLevel finds the Pokemon's level by finding the highest level
// where CP(level) <= currentCP. Returns a level between 1.0 and 50.0.
func CalculateLevel(baseAtk, baseDef, baseHP, ivAtk, ivDef, ivHP, currentCP int) float64 {
	bestLevel := 1.0
	for level := 1.0; level <= 50.0; level += 0.5 {
		cp := CalculateCP(baseAtk, baseDef, baseHP, ivAtk, ivDef, ivHP, level)
		if cp <= currentCP {
			bestLevel = level
		} else {
			break
		}
	}
	return bestLevel
}

// NextPowerUpCP returns the CP after one power-up (level + 0.5).
// If already at level 50, returns currentCP.
func NextPowerUpCP(baseAtk, baseDef, baseHP, ivAtk, ivDef, ivHP, currentCP int) int {
	level := CalculateLevel(baseAtk, baseDef, baseHP, ivAtk, ivDef, ivHP, currentCP)
	if level >= 50 {
		return currentCP
	}
	return CalculateCP(baseAtk, baseDef, baseHP, ivAtk, ivDef, ivHP, level+0.5)
}

// PurposeCap returns the CP cap for a given purpose and development goal.
// CelRozwoju takes precedence over Purpose. Returns 0 if no cap applies.
func PurposeCap(purpose, celRozwoju string, maxCP int) int {
	// CelRozwoju determines the specific league/target
	switch strings.ToLower(celRozwoju) {
	case "ul", "ultra league", "pvp ul":
		return 2500
	case "gl", "great league":
		return 1500
	case "ml", "master league", "max":
		return maxCP
	}
	// Fallback on Purpose — PvP without CelRozwoju = ML (no cap)
	switch strings.ToLower(purpose) {
	case "rajdy", "raid", "gym":
		return maxCP
	default:
		return maxCP
	}
}

type ScoreResult struct {
	Score      float64
	IVScore    float64
	LuckyPts   float64
	MetaPts    float64
	MovePts    float64
	MegaPts    float64
	CPPts      float64
	PurposePts float64
	Notes      []string
}

func CalculateContextScore(p model.Pokemon, tier string, gmMap map[string]*model.GamemasterPokemon) ScoreResult {
	var r ScoreResult

	if p.Purpose == "Transfer" || p.Purpose == "Kolekcja" {
		return r
	}

	r.IVScore = math.Min(p.IVPct/config.Global.Scoring.IVWeight*float64(config.Global.Scoring.IVMax), float64(config.Global.Scoring.IVMax))

	if p.IsLucky {
		r.LuckyPts = config.Global.Scoring.LuckyPts
	}

	r.MetaPts = gamemaster.TierToScore(tier)

	// Mega detection from species name (e.g. "charizard_mega", "charizard_mega_x")
	if strings.Contains(p.SpeciesID, "_mega") {
		r.MegaPts = config.Global.Scoring.MegaPts
	}

	gmPokemon := gmMap[p.SpeciesID]
	if gmPokemon != nil {
		moveScore := 0
		if p.FastMove != "" && isMoveAvailable(p.FastMove, gmPokemon.FastMoves) {
			moveScore += int(config.Global.Scoring.MovePts)
		} else if p.FastMove != "" && len(gmPokemon.FastMoves) > 0 {
			r.Notes = append(r.Notes, "Zmien szybki na: "+gmPokemon.FastMoves[0])
		}
		if p.Charged1 != "" && isMoveAvailable(p.Charged1, gmPokemon.ChargedMoves) {
			moveScore += int(config.Global.Scoring.MovePts)
		} else if p.Charged1 != "" && len(gmPokemon.ChargedMoves) > 0 {
			r.Notes = append(r.Notes, "Zmien charged 1 na: "+gmPokemon.ChargedMoves[0])
		}
		if p.Charged2 != "" && isMoveAvailable(p.Charged2, gmPokemon.ChargedMoves) {
			moveScore += int(config.Global.Scoring.MovePts)
		} else if p.Charged2 != "" && len(gmPokemon.ChargedMoves) > 1 {
			r.Notes = append(r.Notes, "Zmien charged 2 na: "+gmPokemon.ChargedMoves[1])
		}
		r.MovePts = float64(moveScore)
	}

	if p.BaseStats.Atk > 0 && p.BaseStats.Def > 0 && p.BaseStats.HP > 0 {
		maxCP := CalculateMaxCP(p.BaseStats.Atk, p.BaseStats.Def, p.BaseStats.HP, 15, 15, 15)
		if maxCP > 0 {
			r.CPPts = math.Min(float64(p.CP)/float64(maxCP)*config.Global.Scoring.CPMaxPts, config.Global.Scoring.CPMaxPts)
		}
	}

	r.Score = r.IVScore + r.LuckyPts + r.MetaPts + r.MovePts + r.MegaPts + r.CPPts
	r.Score = math.Max(0, math.Min(r.Score, config.Global.Scoring.ScoreMax))

	return r
}

func CalculateDevScore(p model.Pokemon, gmMap map[string]*model.GamemasterPokemon, metaCats gamemaster.MetaTiersByCategory) ScoreResult {
	if p.Purpose == "Transfer" || p.Purpose == "Kolekcja" {
		var r ScoreResult
		r.Notes = append(r.Notes, "Transfer/Kolekcja — nie inwestuj")
		return r
	}

	tier := resolveTierForPurpose(p, metaCats)
	return CalculateContextScore(p, tier, gmMap)
}

func resolveTierForPurpose(p model.Pokemon, metaCats gamemaster.MetaTiersByCategory) string {
	purposeLower := strings.ToLower(p.Purpose)

	switch purposeLower {
	case "rajdy", "raid":
		return p.RaidTier
	case "gl", "great league":
		if p.GLTier != "" {
			return p.GLTier
		}
		return metaCats.BestMetaTier(p.SpeciesID)
	case "ul", "ultra league":
		if p.ULTier != "" {
			return p.ULTier
		}
		return metaCats.BestMetaTier(p.SpeciesID)
	case "ml", "master league":
		if p.MLTier != "" {
			return p.MLTier
		}
		return metaCats.BestMetaTier(p.SpeciesID)
	case "pvp":
		if p.CP <= 1500 && p.GLTier != "" {
			return p.GLTier
		} else if p.CP <= 2500 && p.ULTier != "" {
			return p.ULTier
		} else if p.MLTier != "" {
			return p.MLTier
		}
		return metaCats.BestMetaTier(p.SpeciesID)
	case "gym":
		return p.RaidTier
	default:
		return metaCats.BestMetaTier(p.SpeciesID)
	}
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

// EnrichWithScores computes per-context scores, DevScore, and raid DPS for all Pokemon.
func EnrichWithScores(collection []model.Pokemon, gmMap map[string]*model.GamemasterPokemon, moveMap map[string]*gamemaster.MoveEntry, metaCats gamemaster.MetaTiersByCategory) []model.Pokemon {
	for i := range collection {
		p := &collection[i]

		p.MaxCP = CalculateMaxCP(p.BaseStats.Atk, p.BaseStats.Def, p.BaseStats.HP, p.IVAtk, p.IVDef, p.IVHP)
		p.EffectiveATK = CalculateEffectiveATK(*p)

		fast := moveMap[normalizeMoveID(p.FastMove)]
		charged := moveMap[normalizeMoveID(p.Charged1)]
		if fast != nil && charged != nil {
			p.MoveDPSFast, p.MoveDPSCharged = CalculateMoveDPSRaw(
				fast.Power, fast.Cooldown, fast.EnergyGain,
				charged.Power, charged.Cooldown, charged.Energy,
			)
		}

		raidResult := CalculateContextScore(*p, p.RaidTier, gmMap)
		glResult := CalculateContextScore(*p, p.GLTier, gmMap)
		ulResult := CalculateContextScore(*p, p.ULTier, gmMap)
		mlResult := CalculateContextScore(*p, p.MLTier, gmMap)

		p.RaidScore = raidResult.Score
		p.GLScore = glResult.Score
		p.ULScore = ulResult.Score
		p.MLScore = mlResult.Score

		devResult := CalculateDevScore(*p, gmMap, metaCats)
		p.DevScore = devResult.Score

		dps, _, _ := CalculatePokemonDPS(*p, gmMap, moveMap, "")
		p.RaidDPS = dps

		_, _, optimalDPS := FindOptimalMoves(*p, gmMap, moveMap, "")
		p.RaidDPSOptimal = optimalDPS
	}
	return collection
}
