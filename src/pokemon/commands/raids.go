package commands

import (
	"sort"
	"strings"

	"github.com/adzik/pokemon/config"
	"github.com/adzik/pokemon/model"
)

// Type effectiveness multipliers for raids (attacking)
const NeutralEff = 1.0

// CounterType returns the types that are super effective against a boss type.
func CounterType(bossType string) []string {
	counters := map[string][]string{
		"bug":      {"fire", "flying", "rock"},
		"dark":     {"fighting", "bug", "fairy"},
		"dragon":   {"ice", "dragon", "fairy"},
		"electric": {"ground"},
		"fairy":    {"poison", "steel"},
		"fighting": {"flying", "psychic", "fairy"},
		"fire":     {"water", "ground", "rock"},
		"flying":   {"electric", "ice", "rock"},
		"ghost":    {"ghost", "dark"},
		"grass":    {"fire", "ice", "poison", "flying", "bug"},
		"ground":   {"water", "grass", "ice"},
		"ice":      {"fire", "fighting", "rock", "steel"},
		"normal":   {"fighting"},
		"poison":   {"ground", "psychic"},
		"psychic":  {"bug", "ghost", "dark"},
		"rock":     {"water", "grass", "fighting", "ground", "steel"},
		"steel":    {"fire", "fighting", "ground"},
		"water":    {"electric", "grass"},
	}
	return counters[bossType]
}

// IsSuperEffective checks if moveType is super effective vs bossType.
func IsSuperEffective(moveType, bossType string) bool {
	for _, t := range CounterType(bossType) {
		if moveType == t {
			return true
		}
	}
	return false
}

// IsSTAB checks if move type matches any of the Pokemon's types.
func IsSTAB(moveType string, pokemonTypes []string) bool {
	for _, t := range pokemonTypes {
		if moveType == t {
			return true
		}
	}
	return false
}

// Effectiveness returns the type effectiveness multiplier.
func Effectiveness(moveType, bossType string) float64 {
	if bossType == "" {
		return NeutralEff
	}
	if IsSuperEffective(moveType, bossType) {
		return config.Global.Raids.SEMultiplier
	}
	return NeutralEff
}

// RaidDPSResult holds the DPS calculation result for a Pokemon.
type RaidDPSResult struct {
	Pokemon     model.Pokemon
	DPS         float64 // DPS with current moves
	OptimalDPS  float64 // DPS with best possible moves
	FastMove    string  // current fast move name
	ChargedMove string  // current charged move name
	BestFast    string  // optimal fast move name
	BestCharged string  // optimal charged move name
	SuperEff    bool    // any move is super effective?
}

// CalculateMoveDPS calculates DPS for a specific fast+charged move combination.
// Uses proper ATK/DEF ratio against Tier 5 boss (DEF ~162.3).
func CalculateMoveDPS(
	baseAtk, ivAtk int,
	fastPower, fastCD, fastEnergy int,
	chargedPower, chargedCD, chargedEnergy int,
	fastType, chargedType string,
	pokemonTypes []string,
	bossType string,
	isShadow bool,
) float64 {
	atk := float64(baseAtk + ivAtk) * CPM_L40
	if isShadow {
		atk *= config.Global.Raids.ShadowATKBonus
	}

	// ATK/DEF ratio against Tier 5 boss
	atkDefRatio := atk / config.Global.Raids.BossDefenseRef

	// Fast move DPS (damage per second)
	fastDmgPerHit := 0.5 * float64(fastPower) * atkDefRatio
	fastDPS := fastDmgPerHit / (float64(fastCD) / 1000.0)

	// Charged move contribution (accounting for energy generation)
	chargedDPSContrib := 0.0
	if fastEnergy > 0 && chargedEnergy > 0 {
		timeToCharge := float64(chargedEnergy) / float64(fastEnergy) * (float64(fastCD) / 1000.0)
		chargedDmgPerHit := 0.5 * float64(chargedPower) * atkDefRatio
		chargedDPSContrib = chargedDmgPerHit / timeToCharge
	}

	rawDPS := fastDPS + chargedDPSContrib

	// Apply STAB (if either fast or charged has STAB, apply bonus)
	stab := 1.0
	if IsSTAB(fastType, pokemonTypes) || IsSTAB(chargedType, pokemonTypes) {
		stab = 1.2
	}

	// Apply type effectiveness (use best of fast/charged)
	eff := NeutralEff
	if bossType != "" {
		fastEff := Effectiveness(fastType, bossType)
		chargedEff := Effectiveness(chargedType, bossType)
		if chargedEff > fastEff {
			eff = chargedEff
		} else {
			eff = fastEff
		}
	}

	return rawDPS * stab * eff
}

// CalculateMoveDPSRaw calculates pure move DPS without Pokemon stats.
// Returns fast move DPS and charged move DPS as raw values.
func CalculateMoveDPSRaw(
	fastPower, fastCD, fastEnergy int,
	chargedPower, chargedCD, chargedEnergy int,
) (float64, float64) {
	fastDPS := float64(fastPower) / (float64(fastCD) / 1000.0)

	chargedDPS := 0.0
	if fastEnergy > 0 && chargedEnergy > 0 {
		timeToCharge := float64(chargedEnergy) / float64(fastEnergy) * (float64(fastCD) / 1000.0)
		chargedDPS = float64(chargedPower) / timeToCharge
	}

	return fastDPS, chargedDPS
}

// normalizeMoveID converts a move name to a gamemaster moveId format.
func normalizeMoveID(name string) string {
	return strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(name), " ", "_"))
}

// CalculatePokemonDPS calculates DPS for a Pokemon with its actual moves.
func CalculatePokemonDPS(p model.Pokemon, gmMap map[string]*model.GamemasterPokemon, moveMap map[string]*MoveEntry, bossType string) (float64, string, string) {
	gm := gmMap[p.SpeciesID]
	if gm == nil {
		return 0, "", ""
	}

	pokemonTypes := []string{strings.ToLower(p.Type1), strings.ToLower(p.Type2)}

	fast := moveMap[normalizeMoveID(p.FastMove)]
	charged := moveMap[normalizeMoveID(p.Charged1)]

	if fast == nil || charged == nil {
		return 0, p.FastMove, p.Charged1
	}

	dps := CalculateMoveDPS(
		p.BaseStats.Atk, p.IVAtk,
		fast.Power, fast.Cooldown, fast.EnergyGain,
		charged.Power, charged.Cooldown, charged.Energy,
		fast.Type, charged.Type,
		pokemonTypes,
		bossType,
		p.IsShadow,
	)

	return dps, fast.Name, charged.Name
}

// CalculateEffectiveATK returns the Pokemon's effective ATK stat.
func CalculateEffectiveATK(p model.Pokemon) float64 {
	atk := float64(p.BaseStats.Atk + p.IVAtk) * CPM_L40
	if p.IsShadow {
		atk *= config.Global.Raids.ShadowATKBonus
	}
	return atk
}

// FindOptimalMoves finds the best fast+charged move combination for DPS.
func FindOptimalMoves(p model.Pokemon, gmMap map[string]*model.GamemasterPokemon, moveMap map[string]*MoveEntry, bossType string) (string, string, float64) {
	gm := gmMap[p.SpeciesID]
	if gm == nil {
		return "", "", 0
	}

	pokemonTypes := []string{strings.ToLower(p.Type1), strings.ToLower(p.Type2)}

	bestFast := ""
	bestCharged := ""
	bestDPS := 0.0

	for _, fastID := range gm.FastMoves {
		fast := moveMap[fastID]
		if fast == nil {
			continue
		}
		for _, chargedID := range gm.ChargedMoves {
			charged := moveMap[chargedID]
			if charged == nil {
				continue
			}
			dps := CalculateMoveDPS(
				p.BaseStats.Atk, p.IVAtk,
				fast.Power, fast.Cooldown, fast.EnergyGain,
				charged.Power, charged.Cooldown, charged.Energy,
				fast.Type, charged.Type,
				pokemonTypes,
				bossType,
				p.IsShadow,
			)
			if dps > bestDPS {
				bestDPS = dps
				bestFast = fast.Name
				bestCharged = charged.Name
			}
		}
	}

	return bestFast, bestCharged, bestDPS
}

// FindBestCounters returns the best raid counters from collection for a boss type.
func FindBestCounters(
	collection []model.Pokemon,
	gmMap map[string]*model.GamemasterPokemon,
	moveMap map[string]*MoveEntry,
	bossType string,
	n int,
) []RaidDPSResult {
	var results []RaidDPSResult

	for _, p := range collection {
		if p.CP < config.Global.Raids.MinCP {
			continue
		}

		dps, fast, charged := CalculatePokemonDPS(p, gmMap, moveMap, bossType)
		if dps == 0 {
			continue
		}

		bestFast, bestCharged, optimalDPS := FindOptimalMoves(p, gmMap, moveMap, bossType)

		pokemonTypes := []string{strings.ToLower(p.Type1), strings.ToLower(p.Type2)}
		superEff := false
		if bossType != "" {
			superEff = IsSuperEffective(fast, bossType) || IsSuperEffective(charged, bossType)
		}
		_ = IsSTAB(fast, pokemonTypes) // used for future reference

		results = append(results, RaidDPSResult{
			Pokemon:     p,
			DPS:         dps,
			OptimalDPS:  optimalDPS,
			FastMove:    fast,
			ChargedMove: charged,
			BestFast:    bestFast,
			BestCharged: bestCharged,
			SuperEff:    superEff,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].DPS > results[j].DPS
	})

	if len(results) > n {
		results = results[:n]
	}
	return results
}

// RaidRankings holds per-type raid rankings.
type RaidRankings struct {
	ByType    map[string][]RaidDPSResult // bossType -> TOP N counters
	Universal []RaidDPSResult            // effectiveness=1.0 for all
}

// BuildRaidRankings builds per-type raid rankings for all 18 boss types + universal.
func BuildRaidRankings(
	collection []model.Pokemon,
	gmMap map[string]*model.GamemasterPokemon,
	moveMap map[string]*MoveEntry,
	n int,
) RaidRankings {
	rankings := RaidRankings{
		ByType: make(map[string][]RaidDPSResult),
	}

	for _, bossType := range BossTypes {
		rankings.ByType[bossType] = FindBestCounters(collection, gmMap, moveMap, bossType, n)
	}

	rankings.Universal = FindBestCounters(collection, gmMap, moveMap, "", n)

	return rankings
}
