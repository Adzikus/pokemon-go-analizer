package commands

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/adzik/pokemon/ods"
	"github.com/adzik/pokemon/pkg/analysis"
	"github.com/adzik/pokemon/pkg/gamemaster"
)

func TestCalculateCP(t *testing.T) {
	tests := []struct {
		name                     string
		baseAtk, baseDef, baseHP int
		ivAtk, ivDef, ivHP       int
		level                    float64
		wantCP                   int
	}{
		{"bulbasaur 15/15/15 lvl 50", 118, 111, 128, 15, 15, 15, 50, 0},
		{"bulbasaur 10/10/10 lvl 30", 118, 111, 128, 10, 10, 10, 30, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cp := analysis.CalculateCP(tt.baseAtk, tt.baseDef, tt.baseHP, tt.ivAtk, tt.ivDef, tt.ivHP, tt.level)
			if cp <= 0 {
				t.Errorf("analysis.CalculateCP() = %d, want > 0", cp)
			}
		})
	}
}

func TestCalculateCPZeroIVs(t *testing.T) {
	cp := analysis.CalculateCP(100, 100, 100, 0, 0, 0, 50)
	if cp <= 0 {
		t.Errorf("analysis.CalculateCP(0 IVs) = %d, want > 0", cp)
	}
}

func TestCalculateCPMaxIVsHigher(t *testing.T) {
	cpLow := analysis.CalculateCP(100, 100, 100, 0, 0, 0, 50)
	cpHigh := analysis.CalculateCP(100, 100, 100, 15, 15, 15, 50)
	if cpHigh <= cpLow {
		t.Errorf("max IVs CP (%d) should be > min IVs CP (%d)", cpHigh, cpLow)
	}
}

func TestCalculateMaxCP(t *testing.T) {
	cp := analysis.CalculateMaxCP(118, 111, 128, 15, 15, 15)
	if cp <= 0 {
		t.Errorf("analysis.CalculateMaxCP() = %d, want > 0", cp)
	}
}

func TestIVPercentage(t *testing.T) {
	tests := []struct {
		atk, def, hp int
		wantPct      float64
	}{
		{15, 15, 15, 100.0},
		{0, 0, 0, 0.0},
		{10, 10, 10, 66.66666666666667},
		{5, 5, 5, 33.33333333333333},
	}
	for _, tt := range tests {
		name := fmt.Sprintf("%d/%d/%d", tt.atk, tt.def, tt.hp)
		t.Run(name, func(t *testing.T) {
			pct := float64(tt.atk+tt.def+tt.hp) / 45.0 * 100.0
			if math.Abs(pct-tt.wantPct) > 0.001 {
				t.Errorf("IV%% = %f, want %f", pct, tt.wantPct)
			}
		})
	}
}

func TestBuildKolekcjaRowsEmpty(t *testing.T) {
	rows := buildKolekcjaRows(100, 50, nil)
	if len(rows) != 1000 {
		t.Fatalf("expected 1000 rows, got %d", len(rows))
	}
	for i, row := range rows {
		if len(row) != 27 {
			t.Errorf("row %d: expected 27 cells, got %d", i, len(row))
		}
		if row[1].Validation != "valGat" {
			t.Errorf("row %d col B: validation = %q, want %q", i, row[1].Validation, "valGat")
		}
		if row[9].Type != "formula" {
			t.Errorf("row %d col J: type = %q, want formula", i, row[9].Type)
		}
	}
}

func TestBuildKolekcjaRowsWithData(t *testing.T) {
	userData := [][]ods.Cell{
		{
			{Type: "text", Value: "P-00001"},
			{Type: "text", Value: "Bulbasaur"},
			{Type: "text", Value: "Groszek"},
			{}, {}, {},
			{Type: "text", Value: "15"},
			{Type: "text", Value: "14"},
			{Type: "text", Value: "15"},
		},
	}
	rows := buildKolekcjaRows(10, 5, userData)
	if len(rows) != 1000 {
		t.Fatalf("expected 1000 rows, got %d", len(rows))
	}
	row0 := rows[0]
	if row0[0].Value != "P-00001" {
		t.Errorf("row 0 col A: value = %q, want P-00001", row0[0].Value)
	}
	if row0[1].Value != "Bulbasaur" {
		t.Errorf("row 0 col B: value = %q, want Bulbasaur", row0[1].Value)
	}
	if row0[3].Type != "formula" {
		t.Errorf("row 0 col D: type = %q, want formula", row0[3].Type)
	}
	if row0[1].Validation != "valGat" {
		t.Errorf("row 0 col B: validation = %q, want valGat", row0[1].Validation)
	}
	if row0[10].Validation != "valF2" {
		t.Errorf("row 0 col K: validation = %q, want valF2", row0[10].Validation)
	}
}

func TestBuildKolekcjaRowsFormulas(t *testing.T) {
	rows := buildKolekcjaRows(100, 50, nil)
	// Row 2 (first data row) should have specific formula references
	row0 := rows[0]
	// D formula references row 2
	if !contains(row0[3].Formula, "[.B2]") {
		t.Errorf("row 0 col D formula %q should reference [.B2]", row0[3].Formula)
	}
	// J formula references row 2
	if !contains(row0[9].Formula, "[.G2]") {
		t.Errorf("row 0 col J formula %q should reference [.G2]", row0[9].Formula)
	}
	// Row 100 (99th data row) should reference row 101
	row99 := rows[99]
	if !contains(row99[3].Formula, "[.B101]") {
		t.Errorf("row 99 col D formula %q should reference [.B101]", row99[3].Formula)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestExtractCollectionRows(t *testing.T) {
	sheet := &ods.Sheet{
		Rows: [][]string{
			{"UID", "Species", "Nick"},
			{"P-00001", "Bulbasaur", "Groszek"},
			{"P-00002", "Ivysaur", ""},
			{},
			{"P-00003", "Venusaur", "Wielki"},
			{},
		},
	}
	rows := extractCollectionRows(sheet)
	if len(rows) != 3 {
		t.Fatalf("expected 3 data rows, got %d", len(rows))
	}
	if rows[0][0] != "P-00001" {
		t.Errorf("row 0: %q, want P-00001", rows[0][0])
	}
	if rows[2][1] != "Venusaur" {
		t.Errorf("row 2: %q, want Venusaur", rows[2][1])
	}
}

func TestExtractCollectionRowsAllEmpty(t *testing.T) {
	sheet := &ods.Sheet{
		Rows: [][]string{
			{"UID", "Species"},
			{},
			{},
		},
	}
	rows := extractCollectionRows(sheet)
	if len(rows) != 0 {
		t.Errorf("expected 0 data rows, got %d", len(rows))
	}
}

func TestNormalizeName(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"bulbasaur", "bulbasaur"},
		{"mr-mime", "mr-mime"},
		{"mime-jr", "mime-jr"},
		{"porygon-z", "porygon-z"},
		{"ho-oh", "ho-oh"},
		{"jangmo-o", "jangmo-o"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := gamemaster.NormalizeName(tt.input)
			if got != tt.want {
				t.Errorf("gamemaster.NormalizeName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestExtractKontoFromFilename(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"Pokemon-GO-Adzik.ods", "Adzik"},
		{"Pokemon-GO-Shizaya09.ods", "Shizaya09"},
		{"Pokemon-GO-TEST.ods", "TEST"},
		{"random-file.ods", ""},
		{"Pokemon-GO.ods", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := extractKontoFromFilename(tt.input)
			if got != tt.want {
				t.Errorf("extractKontoFromFilename(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestLoadMetaCacheExpired(t *testing.T) {
	dir := t.TempDir()
	cachePath := filepath.Join(dir, "meta_rankings.json")
	content := `{"fetched_at":"2020-01-01T00:00:00Z","rankings":{}}`
	if err := os.WriteFile(cachePath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	_, ok := loadMetaCache(cachePath)
	if ok {
		t.Error("expected cache to be expired (not ok)")
	}
}

func TestLoadMetaCacheFresh(t *testing.T) {
	dir := t.TempDir()
	cachePath := filepath.Join(dir, "meta_rankings.json")
	now := time.Now().Format(time.RFC3339)
	content := fmt.Sprintf(`{"fetched_at":"%s","rankings":{"great_league_overall":["test"]}}`, now)
	if err := os.WriteFile(cachePath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	result, ok := loadMetaCache(cachePath)
	if !ok {
		t.Error("expected cache to be fresh (ok)")
	}
	if result.age < 0 || result.age > 1 {
		t.Errorf("expected age 0-1 days, got %d", result.age)
	}
}

func TestLoadMetaCacheNotFound(t *testing.T) {
	_, ok := loadMetaCache("/nonexistent/path.json")
	if ok {
		t.Error("expected not found to return false")
	}
}

func TestLoadMetaCacheBadJSON(t *testing.T) {
	dir := t.TempDir()
	cachePath := filepath.Join(dir, "meta_rankings.json")
	if err := os.WriteFile(cachePath, []byte("not json"), 0644); err != nil {
		t.Fatal(err)
	}
	_, ok := loadMetaCache(cachePath)
	if ok {
		t.Error("expected bad JSON to return false")
	}
}

func TestLoadMetaCacheNoFetchedAt(t *testing.T) {
	dir := t.TempDir()
	cachePath := filepath.Join(dir, "meta_rankings.json")
	content := `{"rankings":{}}`
	if err := os.WriteFile(cachePath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	_, ok := loadMetaCache(cachePath)
	if ok {
		t.Error("expected missing fetched_at to return false")
	}
}

func TestGetCPM(t *testing.T) {
	// CPM table is linear in this implementation
	// L1 = index 0, L50 = index 98
	cpm1 := analysis.GetCPM(1)
	if cpm1 < 0.09 || cpm1 > 0.10 {
		t.Errorf("analysis.GetCPM(1) = %f, want ~0.094", cpm1)
	}
	cpm50 := analysis.GetCPM(50)
	if cpm50 < 1.07 || cpm50 > 1.09 {
		t.Errorf("analysis.GetCPM(50) = %f, want ~1.076", cpm50)
	}
}

func TestGetCPMNonDecreasing(t *testing.T) {
	var prev float64
	for level := 1.0; level <= 50; level += 0.5 {
		cpm := analysis.GetCPM(level)
		if cpm < prev {
			t.Errorf("analysis.GetCPM(%f) = %f < previous %f", level, cpm, prev)
		}
		prev = cpm
	}
}
