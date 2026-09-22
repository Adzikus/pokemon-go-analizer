package analysis

import (
	"fmt"
	"sort"
	"strings"

	"github.com/adzik/pokemon/model"
	"github.com/adzik/pokemon/pkg/gamemaster"
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

	if len(family) == 1 {
		return Keep, "Jedyny w rodzinie"
	}
	if p.IVPct == 100 {
		return Keep, "HUNDO"
	}

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

	bestTier := gamemaster.BestTier(gamemaster.BestTier(p.RaidTier, p.GLTier), gamemaster.BestTier(p.ULTier, p.MLTier))
	if gamemaster.TierRank(bestTier) >= gamemaster.TierRank("A") {
		count := 0
		for _, fp := range family {
			if fp.SpeciesID == p.SpeciesID {
				count++
			}
		}
		if count >= 10 {
			return Trade, fmt.Sprintf(" %d sztuk tego gatunku, wiec najs slabiej", count)
		}
		return Keep, fmt.Sprintf("Meta %s, %d szt.", bestTier, count)
	}

	if strings.ToLower(p.Purpose) == "kolekcja" {
		return SoftKeep, "Element oznaczony do kolekcji - usuwac w ostatecznosci"
	}

	return Safe, "Brak wartosci bojowej/kolekcjonerskiej"
}
