package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/adzik/pokemon/ods"
)

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
	row0 := rows[0]
	if !contains(row0[3].Formula, "[.B2]") {
		t.Errorf("row 0 col D formula %q should reference [.B2]", row0[3].Formula)
	}
	if !contains(row0[9].Formula, "[.G2]") {
		t.Errorf("row 0 col J formula %q should reference [.G2]", row0[9].Formula)
	}
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
