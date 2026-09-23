package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.URLs.Gamemaster == "" {
		t.Error("Default().URLs.Gamemaster is empty")
	}
	if len(cfg.URLs.Rankings) == 0 {
		t.Error("Default().URLs.Rankings is empty")
	}
	if cfg.Paths.Data == "" {
		t.Error("Default().Paths.Data is empty")
	}
	if cfg.Cache.MaxAgeDays <= 0 {
		t.Errorf("Default().Cache.MaxAgeDays = %d, want > 0", cfg.Cache.MaxAgeDays)
	}
	if cfg.Cache.MaxBackups <= 0 {
		t.Errorf("Default().Cache.MaxBackups = %d, want > 0", cfg.Cache.MaxBackups)
	}
	if cfg.Raids.BossDefenseRef <= 0 {
		t.Errorf("Default().Raids.BossDefenseRef = %f, want > 0", cfg.Raids.BossDefenseRef)
	}
	if cfg.Raids.SEMultiplier <= 1.0 {
		t.Errorf("Default().Raids.SEMultiplier = %f, want > 1.0", cfg.Raids.SEMultiplier)
	}
}

func TestDefaultTiersDescending(t *testing.T) {
	cfg := Default()
	tiers := cfg.Tiers

	if tiers.S >= tiers.Ap {
		t.Errorf("Tiers: S(%d) should be < Ap(%d)", tiers.S, tiers.Ap)
	}
	if tiers.Ap >= tiers.A {
		t.Errorf("Tiers: Ap(%d) should be < A(%d)", tiers.Ap, tiers.A)
	}
	if tiers.A >= tiers.Bp {
		t.Errorf("Tiers: A(%d) should be < Bp(%d)", tiers.A, tiers.Bp)
	}
	if tiers.Bp >= tiers.B {
		t.Errorf("Tiers: Bp(%d) should be < B(%d)", tiers.Bp, tiers.B)
	}
	if tiers.B >= tiers.C {
		t.Errorf("Tiers: B(%d) should be < C(%d)", tiers.B, tiers.C)
	}
}

func TestDefaultLeagueCaps(t *testing.T) {
	cfg := Default()

	if cfg.PVP.LeagueCaps["GL"] != 1500 {
		t.Errorf("GL cap = %d, want 1500", cfg.PVP.LeagueCaps["GL"])
	}
	if cfg.PVP.LeagueCaps["UL"] != 2500 {
		t.Errorf("UL cap = %d, want 2500", cfg.PVP.LeagueCaps["UL"])
	}
	if cfg.PVP.LeagueCaps["ML"] != 0 {
		t.Errorf("ML cap = %d, want 0", cfg.PVP.LeagueCaps["ML"])
	}
}

func TestLoadExisting(t *testing.T) {
	dir := t.TempDir()
	dataDir := filepath.Join(dir, "data")
	os.MkdirAll(dataDir, 0755)

	cfgJSON := `{"cache":{"max_age_days":7,"max_backups":3},"raids":{"min_cp":500}}`
	os.WriteFile(filepath.Join(dataDir, "config.json"), []byte(cfgJSON), 0644)

	cfg := Load(dir)

	if cfg.Cache.MaxAgeDays != 7 {
		t.Errorf("Cache.MaxAgeDays = %d, want 7", cfg.Cache.MaxAgeDays)
	}
	if cfg.Cache.MaxBackups != 3 {
		t.Errorf("Cache.MaxBackups = %d, want 3", cfg.Cache.MaxBackups)
	}
	if cfg.Raids.MinCP != 500 {
		t.Errorf("Raids.MinCP = %d, want 500", cfg.Raids.MinCP)
	}
	// Non-overridden values should be defaults
	if cfg.Raids.BossDefenseRef != 162.3 {
		t.Errorf("Raids.BossDefenseRef = %f, want 162.3 (default)", cfg.Raids.BossDefenseRef)
	}
}

func TestLoadMissing(t *testing.T) {
	dir := t.TempDir()
	// No config.json → defaults
	cfg := Load(dir)

	if cfg.Cache.MaxAgeDays != 30 {
		t.Errorf("Cache.MaxAgeDays = %d, want 30 (default)", cfg.Cache.MaxAgeDays)
	}
	if cfg.URLs.Gamemaster == "" {
		t.Error("URLs.Gamemaster should have default value")
	}
}

func TestLoadBadJSON(t *testing.T) {
	dir := t.TempDir()
	dataDir := filepath.Join(dir, "data")
	os.MkdirAll(dataDir, 0755)

	os.WriteFile(filepath.Join(dataDir, "config.json"), []byte("not json at all"), 0644)

	cfg := Load(dir)

	// Should fall back to defaults
	if cfg.Cache.MaxAgeDays != 30 {
		t.Errorf("Cache.MaxAgeDays = %d, want 30 (default after bad JSON)", cfg.Cache.MaxAgeDays)
	}
}

func TestLoadPartial(t *testing.T) {
	dir := t.TempDir()
	dataDir := filepath.Join(dir, "data")
	os.MkdirAll(dataDir, 0755)

	// Only override scoring, leave everything else as default
	cfgJSON := `{"scoring":{"iv_max":20}}`
	os.WriteFile(filepath.Join(dataDir, "config.json"), []byte(cfgJSON), 0644)

	cfg := Load(dir)

	if cfg.Scoring.IVMax != 20 {
		t.Errorf("Scoring.IVMax = %d, want 20", cfg.Scoring.IVMax)
	}
	// Other scoring values should be defaults
	if cfg.Scoring.LuckyPts != 20 {
		t.Errorf("Scoring.LuckyPts = %f, want 20 (default)", cfg.Scoring.LuckyPts)
	}
	// Non-scoring values should be defaults
	if cfg.Cache.MaxAgeDays != 30 {
		t.Errorf("Cache.MaxAgeDays = %d, want 30 (default)", cfg.Cache.MaxAgeDays)
	}
}

func TestGlobalSetByLoad(t *testing.T) {
	dir := t.TempDir()
	// Load sets Global
	Load(dir)

	if Global == nil {
		t.Fatal("Global is nil after Load()")
	}
	if Global.Cache.MaxAgeDays != 30 {
		t.Errorf("Global.Cache.MaxAgeDays = %d, want 30", Global.Cache.MaxAgeDays)
	}
}
