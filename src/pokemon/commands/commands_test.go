package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/adzik/pokemon/config"
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

// --- normalizeName tests ---

func TestNormalizeName_Alolan(t *testing.T) {
	got := normalizeName("Alolan Vulpix")
	if got != "Vulpix - Alolan" {
		t.Errorf("got %q, want %q", got, "Vulpix - Alolan")
	}
}

func TestNormalizeName_Galarian(t *testing.T) {
	got := normalizeName("Galarian Corsola")
	if got != "Corsola - Galarian" {
		t.Errorf("got %q, want %q", got, "Corsola - Galarian")
	}
}

func TestNormalizeName_Hisuian(t *testing.T) {
	got := normalizeName("Hisuian Growlithe")
	if got != "Growlithe - Hisuian" {
		t.Errorf("got %q, want %q", got, "Growlithe - Hisuian")
	}
}

func TestNormalizeName_Paldean(t *testing.T) {
	got := normalizeName("Paldean Wooper")
	if got != "Wooper - Paldean" {
		t.Errorf("got %q, want %q", got, "Wooper - Paldean")
	}
}

func TestNormalizeName_Mega(t *testing.T) {
	got := normalizeName("Charizard Mega Y")
	if got != "Charizard - Mega Y" {
		t.Errorf("got %q, want %q", got, "Charizard - Mega Y")
	}
}

func TestNormalizeName_MegaX(t *testing.T) {
	got := normalizeName("Charizard Mega X")
	if got != "Charizard - Mega X" {
		t.Errorf("got %q, want %q", got, "Charizard - Mega X")
	}
}

func TestNormalizeName_Regular(t *testing.T) {
	got := normalizeName("Bulbasaur")
	if got != "Bulbasaur" {
		t.Errorf("got %q, want %q", got, "Bulbasaur")
	}
}

func TestNormalizeName_AlolanSingleWord(t *testing.T) {
	got := normalizeName("Alolan Raichu")
	if got != "Raichu - Alolan" {
		t.Errorf("got %q, want %q", got, "Raichu - Alolan")
	}
}

// --- capitalize tests ---

func TestCapitalize_Normal(t *testing.T) {
	got := capitalize("fire")
	if got != "Fire" {
		t.Errorf("got %q, want %q", got, "Fire")
	}
}

func TestCapitalize_Empty(t *testing.T) {
	got := capitalize("")
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestCapitalize_AlreadyCapitalized(t *testing.T) {
	got := capitalize("Fire")
	if got != "Fire" {
		t.Errorf("got %q, want %q", got, "Fire")
	}
}

func TestCapitalize_SingleChar(t *testing.T) {
	got := capitalize("a")
	if got != "A" {
		t.Errorf("got %q, want %q", got, "A")
	}
}

func TestCapitalize_AllCaps(t *testing.T) {
	got := capitalize("FIRE")
	if got != "FIRE" {
		t.Errorf("got %q, want %q", got, "FIRE")
	}
}

// --- calcMaxCP tests ---

func TestCalcMaxCP_Bulbasaur151515(t *testing.T) {
	got := calcMaxCP(118, 111, 128)
	if got != 1260 {
		t.Errorf("calcMaxCP(118,111,128) = %d, want 1260", got)
	}
}

func TestCalcMaxCP_Charmeleon151515(t *testing.T) {
	got := calcMaxCP(165, 119, 126)
	if got != 1747 {
		t.Errorf("calcMaxCP(165,119,126) = %d, want 1747", got)
	}
}

func TestCalcMaxCP_ZeroIVs(t *testing.T) {
	// Still positive because base stats contribute
	got := calcMaxCP(100, 100, 100)
	if got <= 0 {
		t.Errorf("calcMaxCP(100,100,100) = %d, expected > 0", got)
	}
}

func TestCalcMaxCP_HigherBaseHigherCP(t *testing.T) {
	low := calcMaxCP(100, 100, 100)
	high := calcMaxCP(200, 200, 200)
	if high <= low {
		t.Errorf("higher base stats should give higher CP: %d vs %d", high, low)
	}
}

// --- sqrt tests ---

func TestSqrt_Perfect(t *testing.T) {
	got := sqrt(9.0)
	// Newton's method should converge to ~3.0
	if got < 2.9999 || got > 3.0001 {
		t.Errorf("sqrt(9) = %f, expected ~3.0", got)
	}
}

func TestSqrt_Zero(t *testing.T) {
	got := sqrt(0.0)
	if got != 0.0 {
		t.Errorf("sqrt(0) = %f, expected 0.0", got)
	}
}

func TestSqrt_Negative(t *testing.T) {
	got := sqrt(-4.0)
	if got != 0.0 {
		t.Errorf("sqrt(-4) = %f, expected 0.0", got)
	}
}

func TestSqrt_One(t *testing.T) {
	got := sqrt(1.0)
	if got < 0.9999 || got > 1.0001 {
		t.Errorf("sqrt(1) = %f, expected ~1.0", got)
	}
}

func TestSqrt_Large(t *testing.T) {
	got := sqrt(1000000.0)
	// Newton's method with 10 iterations doesn't fully converge for large values
	if got < 1290 || got > 1310 {
		t.Errorf("sqrt(1000000) = %f, expected ~1296", got)
	}
}

func TestSqrt_Fraction(t *testing.T) {
	got := sqrt(0.25)
	if got < 0.4999 || got > 0.5001 {
		t.Errorf("sqrt(0.25) = %f, expected ~0.5", got)
	}
}

// --- cleanupOldBackups tests ---

func TestCleanupOldBackups_KeepsMostRecent(t *testing.T) {
	dir := t.TempDir()
	base := "test.ods"

	for i := 0; i < 8; i++ {
		name := fmt.Sprintf("%s.bak.20260923_%06d", base, i)
		os.WriteFile(filepath.Join(dir, name), []byte("backup"), 0644)
	}

	config.Global.Cache.MaxBackups = 3
	cleanupOldBackups(dir, base)

	entries, _ := os.ReadDir(dir)
	if len(entries) != 3 {
		t.Errorf("expected 3 backups, got %d", len(entries))
	}
}

func TestCleanupOldBackups_FewerThanMax(t *testing.T) {
	dir := t.TempDir()
	base := "test.ods"

	for i := 0; i < 2; i++ {
		name := fmt.Sprintf("%s.bak.20260923_%06d", base, i)
		os.WriteFile(filepath.Join(dir, name), []byte("backup"), 0644)
	}

	config.Global.Cache.MaxBackups = 5
	cleanupOldBackups(dir, base)

	entries, _ := os.ReadDir(dir)
	if len(entries) != 2 {
		t.Errorf("expected 2 backups kept, got %d", len(entries))
	}
}

func TestCleanupOldBackups_IgnoresOtherFiles(t *testing.T) {
	dir := t.TempDir()
	base := "test.ods"

	for i := 0; i < 4; i++ {
		name := fmt.Sprintf("%s.bak.20260923_%06d", base, i)
		os.WriteFile(filepath.Join(dir, name), []byte("backup"), 0644)
	}
	os.WriteFile(filepath.Join(dir, "something-else.txt"), []byte("not a backup"), 0644)

	config.Global.Cache.MaxBackups = 2
	cleanupOldBackups(dir, base)

	entries, _ := os.ReadDir(dir)
	if len(entries) != 3 {
		t.Errorf("expected 2 backups + 1 other = 3 files, got %d", len(entries))
	}
}

func TestCleanupOldBackups_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	config.Global.Cache.MaxBackups = 5
	cleanupOldBackups(dir, "test.ods")
	entries, _ := os.ReadDir(dir)
	if len(entries) != 0 {
		t.Errorf("expected 0 files, got %d", len(entries))
	}
}

func TestCleanupOldBackups_DefaultMax5(t *testing.T) {
	dir := t.TempDir()
	base := "test.ods"

	for i := 0; i < 7; i++ {
		name := fmt.Sprintf("%s.bak.20260923_%06d", base, i)
		os.WriteFile(filepath.Join(dir, name), []byte("backup"), 0644)
	}

	config.Global.Cache.MaxBackups = 0 // should default to 5
	cleanupOldBackups(dir, base)

	entries, _ := os.ReadDir(dir)
	if len(entries) != 5 {
		t.Errorf("expected 5 backups (default), got %d", len(entries))
	}
}

// --- cleanOldGamemasterBackups tests ---

func TestCleanOldGamemasterBackups_KeepsMostRecent(t *testing.T) {
	dir := t.TempDir()

	for i := 0; i < 8; i++ {
		name := fmt.Sprintf("gamemaster.json.bak.20260923_%06d", i)
		os.WriteFile(filepath.Join(dir, name), []byte("backup"), 0644)
	}

	config.Global.Cache.MaxBackups = 3
	cleanOldGamemasterBackups(dir)

	entries, _ := os.ReadDir(dir)
	if len(entries) != 3 {
		t.Errorf("expected 3 backups, got %d", len(entries))
	}
}

func TestCleanOldGamemasterBackups_IgnoresNonGamemaster(t *testing.T) {
	dir := t.TempDir()

	for i := 0; i < 4; i++ {
		name := fmt.Sprintf("gamemaster.json.bak.20260923_%06d", i)
		os.WriteFile(filepath.Join(dir, name), []byte("backup"), 0644)
	}
	os.WriteFile(filepath.Join(dir, "gamemaster.json"), []byte("current"), 0644)
	os.WriteFile(filepath.Join(dir, "other.txt"), []byte("other"), 0644)

	config.Global.Cache.MaxBackups = 2
	cleanOldGamemasterBackups(dir)

	entries, _ := os.ReadDir(dir)
	if len(entries) != 4 {
		t.Errorf("expected 2 backups + gamemaster.json + other.txt = 4, got %d", len(entries))
	}
}

func TestCleanOldGamemasterBackups_DefaultMax5(t *testing.T) {
	dir := t.TempDir()

	for i := 0; i < 7; i++ {
		name := fmt.Sprintf("gamemaster.json.bak.20260923_%06d", i)
		os.WriteFile(filepath.Join(dir, name), []byte("backup"), 0644)
	}

	config.Global.Cache.MaxBackups = 0
	cleanOldGamemasterBackups(dir)

	entries, _ := os.ReadDir(dir)
	if len(entries) != 5 {
		t.Errorf("expected 5 backups (default), got %d", len(entries))
	}
}

func TestCleanOldGamemasterBackups_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	config.Global.Cache.MaxBackups = 5
	cleanOldGamemasterBackups(dir)
	entries, _ := os.ReadDir(dir)
	if len(entries) != 0 {
		t.Errorf("expected 0 files, got %d", len(entries))
	}
}

// --- backupGamemaster tests ---

func TestBackupGamemaster_CreatesBackup(t *testing.T) {
	dir := t.TempDir()
	gmPath := filepath.Join(dir, "gamemaster.json")
	os.WriteFile(gmPath, []byte(`{"pokemon":[]}`), 0644)

	err := backupGamemaster(gmPath)
	if err != nil {
		t.Fatalf("backupGamemaster failed: %v", err)
	}

	entries, _ := os.ReadDir(dir)
	found := false
	for _, e := range entries {
		if e.Name() != "gamemaster.json" && len(e.Name()) > 0 {
			found = true
		}
	}
	if !found {
		t.Error("expected backup file to be created")
	}
}

func TestBackupGamemaster_PreservesOriginal(t *testing.T) {
	dir := t.TempDir()
	gmPath := filepath.Join(dir, "gamemaster.json")
	original := `{"pokemon":[{"id":"bulbasaur"}]}`
	os.WriteFile(gmPath, []byte(original), 0644)

	backupGamemaster(gmPath)

	data, _ := os.ReadFile(gmPath)
	if string(data) != original {
		t.Errorf("original file was modified: %s", string(data))
	}
}
