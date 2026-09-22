package analysis

import (
	"fmt"
	"strings"

	"github.com/adzik/pokemon/model"
)

type MetaCompareRow struct {
	Name     string
	Position int
	Tier     string
	Owned    bool
	Copies   int
	BestIV   float64
	IsLucky  bool
	BestCP   int
	Score    float64
}

func BuildMetaComparison(c []model.Pokemon, raidAttackers []string, gmMap map[string]*model.GamemasterPokemon) []MetaCompareRow {
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

		name := sid
		if gm := gmMap[sid]; gm != nil {
			name = gm.SpeciesName
		}

		copies := bySpecies[sid]
		row := MetaCompareRow{
			Name:     name,
			Position: i + 1,
			Tier:     tier,
			Owned:    len(copies) > 0,
			Copies:   len(copies),
		}

		for _, p := range copies {
			if p.IVPct > row.BestIV || (p.IVPct == row.BestIV && p.CP > row.BestCP) {
				row.BestIV = p.IVPct
				row.BestCP = p.CP
				row.IsLucky = p.IsLucky
				row.Score = p.DevScore
				row.Name = p.Species
			}
		}

		rows = append(rows, row)
	}
	return rows
}

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
