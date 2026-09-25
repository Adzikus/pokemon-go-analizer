package analysis

import (
	"fmt"
	"sort"

	"github.com/adzik/pokemon/config"
	"github.com/adzik/pokemon/model"
)

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

	// 1. Shiny HUNDO → always KEEP
	if p.IsShiny && p.IVPct == 100 {
		return Keep, "Shiny HUNDO"
	}

	// 2. HUNDO (not shiny) → KEEP, unless family has Shiny HUNDO → SOFT_KEEP
	if p.IVPct == 100 {
		for _, fp := range family {
			if fp.IsShiny && fp.IVPct == 100 && fp.UID != p.UID {
				return SoftKeep, "HUNDO, ale w rodzinie jest Shiny HUNDO"
			}
		}
		return Keep, "HUNDO"
	}

	// 3. Lucky → KEEP if alone, SOFT_KEEP if better Lucky exists in family
	if p.IsLucky {
		for _, fp := range family {
			if fp.IsLucky && fp.IVPct > p.IVPct && fp.UID != p.UID {
				return SoftKeep, "Lepszy Lucky w rodzinie"
			}
		}
		return Keep, "Jedyny/najlepszy Lucky w rodzinie"
	}

	// 4. Shiny (not HUNDO)
	if p.IsShiny {
		// Find best shiny in family
		bestShinyIV := p.IVPct
		for _, fp := range family {
			if fp.IsShiny && fp.IVPct > bestShinyIV {
				bestShinyIV = fp.IVPct
			}
		}
		hasBetterShiny := bestShinyIV > p.IVPct

		if !hasBetterShiny {
			return Keep, "Jedyny/najlepszy shiny w rodzinie"
		}
		if p.IVPct < 88 {
			return Trade, "Lepszy shiny w rodzinie, niskie IV"
		}
		return SoftKeep, "Lepszy shiny w rodzinie, ale IV >= 88%%"
	}

	// 5. Mythical → zawsze KEEP (nieodwracalna strata, nie da sie trade)
	if isMythical(p, gmMap) {
		return Keep, "Mityczny — nieodwracalna strata"
	}

	// 6. High IV (>= 88%) → SOFT_KEEP
	if p.IVPct >= 88 {
		return SoftKeep, fmt.Sprintf("Wysokie IV %.1f%%", p.IVPct)
	}

	// 7. Purpose = "kolekcja" → SOFT_KEEP
	if p.Purpose == "kolekcja" {
		return SoftKeep, "Element oznaczony do kolekcji"
	}

	// 8. Shadow → SOFT_KEEP (can't trade, like Lucky)
	if p.IsShadow {
		return SoftKeep, "Shadow — nie moze byc trade"
	}

	// Count species copies
	speciesCount := 0
	for _, fp := range family {
		if fp.SpeciesID == p.SpeciesID {
			speciesCount++
		}
	}

	// 9. Legendary → SOFT_KEEP if within limit, TRADE if over
	if isLegendary(p, gmMap) {
		maxLeg := config.Global.Transfer.MaxLegendaries
		if maxLeg <= 0 {
			maxLeg = 5
		}
		if speciesCount <= maxLeg {
			return SoftKeep, fmt.Sprintf("Legendarny — kopie: %d/%d", speciesCount, maxLeg)
		}
		return Trade, fmt.Sprintf("Duplikat legendarnego (%d szt.) — wart handlu", speciesCount)
	}

	// 10. Excess duplicates (>10 of same species) → TRADE
	if speciesCount > 10 {
		return Trade, fmt.Sprintf("Nadmiar: %d szt. tego gatunku", speciesCount)
	}

	// 11. Find best in family (by IV, then CP)
	best := p
	for _, fp := range family {
		if fp.IVPct > best.IVPct || (fp.IVPct == best.IVPct && fp.CP > best.CP) {
			best = fp
		}
	}

	// 12. CP diff > 50% vs best → TRADE
	if best.UID != p.UID {
		cpDiff := float64(best.CP-p.CP) / float64(best.CP) * 100
		if cpDiff > 50 {
			return Trade, fmt.Sprintf("Roznica CP %.0f%% vs najlepszy w rodzinie", cpDiff)
		}
		// 13. Lower IV than best → TRADE
		if best.IVPct > p.IVPct {
			return Trade, "Nizsze IV niz najlepszy w rodzinie"
		}
	}

	// 14. SAFE
	return Safe, "Brak wartosci bojowej/kolekcjonerskiej"
}

func isMythical(p model.Pokemon, gmMap map[string]*model.GamemasterPokemon) bool {
	gm, ok := gmMap[p.SpeciesID]
	if !ok {
		return false
	}
	for _, t := range gm.Tags {
		if t == "mythical" {
			return true
		}
	}
	return false
}

func isLegendary(p model.Pokemon, gmMap map[string]*model.GamemasterPokemon) bool {
	gm, ok := gmMap[p.SpeciesID]
	if !ok {
		return false
	}
	for _, t := range gm.Tags {
		if t == "legendary" {
			return true
		}
	}
	return false
}
