package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	URLs      URLsConfig      `json:"urls"`
	Paths     PathsConfig     `json:"paths"`
	Cache     CacheConfig     `json:"cache"`
	Raids     RaidsConfig     `json:"raids"`
	Tiers     TiersConfig     `json:"tiers"`
	MetaTiers MetaTiersConfig `json:"meta_tiers"`
	Scoring   ScoringConfig   `json:"scoring"`
	Report    ReportConfig    `json:"report"`
	PVP       PVPConfig       `json:"pvp"`
	Meta      MetaConfig      `json:"meta"`
}

type MetaConfig struct {
	TopN int `json:"top_n"`
}

type URLsConfig struct {
	Gamemaster string            `json:"gamemaster"`
	Rankings   map[string]string `json:"rankings"`
}

type PathsConfig struct {
	Data    string `json:"data"`
	Backups string `json:"backups"`
	Analizy string `json:"analizy"`
}

type CacheConfig struct {
	MaxAgeDays int `json:"max_age_days"`
	MaxBackups int `json:"max_backups"`
}

type RaidsConfig struct {
	MinCP          int     `json:"min_cp"`
	BossDefenseRef float64 `json:"boss_defense_ref"`
	SEMultiplier   float64 `json:"se_multiplier"`
	NVEMultiplier  float64 `json:"nve_multiplier"`
	ShadowATKBonus float64 `json:"shadow_atk_bonus"`
}

type TiersConfig struct {
	S  int `json:"S"`
	Ap int `json:"A+"`
	A  int `json:"A"`
	Bp int `json:"B+"`
	B  int `json:"B"`
	C  int `json:"C"`
}

type MetaTiersConfig struct {
	S int `json:"S"`
	A int `json:"A"`
	B int `json:"B"`
}

type ScoringConfig struct {
	IVMax    int     `json:"iv_max"`
	LuckyPts float64 `json:"lucky_pts"`
	MovePts  float64 `json:"move_pts"`
	CPMaxPts float64 `json:"cp_max_pts"`
	ScoreMax float64 `json:"score_max"`
	IVWeight float64 `json:"iv_weight"`
	MegaPts  float64 `json:"mega_pts"`
}

type ReportConfig struct {
	ModelName       string `json:"model_name"`
	MaxMoveDisplay  int    `json:"max_move_display"`
	MaxPVPEntries   int    `json:"max_pvp_entries"`
	MaxTransferCand int    `json:"max_transfer_candidates"`
	MaxTeamSize     int    `json:"max_team_size"`
}

type PVPConfig struct {
	LeagueCaps map[string]int `json:"league_caps"`
}

// Global holds the active config, set by Load().
var Global *Config

func Default() *Config {
	return &Config{
		URLs: URLsConfig{
			Gamemaster: "https://raw.githubusercontent.com/pvpoke/pvpoke/master/src/data/gamemaster.json",
			Rankings: map[string]string{
				"great_league_overall":  "https://raw.githubusercontent.com/pvpoke/pvpoke/master/src/data/rankings/all/overall/rankings-1500.json",
				"ultra_league_overall":  "https://raw.githubusercontent.com/pvpoke/pvpoke/master/src/data/rankings/all/overall/rankings-2500.json",
				"master_league_overall": "https://raw.githubusercontent.com/pvpoke/pvpoke/master/src/data/rankings/all/overall/rankings-10000.json",
				"raid_attackers":        "https://raw.githubusercontent.com/pvpoke/pvpoke/master/src/data/rankings/all/attackers/rankings-10000.json",
			},
		},
		Paths: PathsConfig{
			Data:    "data",
			Backups: "backups",
			Analizy: "analizy",
		},
		Cache: CacheConfig{
			MaxAgeDays: 30,
			MaxBackups: 5,
		},
		Raids: RaidsConfig{
			MinCP:          1000,
			BossDefenseRef: 162.3,
			SEMultiplier:   1.6,
			NVEMultiplier:  0.625,
			ShadowATKBonus: 1.2,
		},
		Tiers: TiersConfig{
			S:  20,
			Ap: 50,
			A:  100,
			Bp: 150,
			B:  200,
			C:  250,
		},
		MetaTiers: MetaTiersConfig{
			S: 10,
			A: 20,
			B: 30,
		},
		Scoring: ScoringConfig{
			IVMax:    15,
			LuckyPts: 20,
			MovePts:  5,
			CPMaxPts: 10,
			ScoreMax: 100,
			IVWeight: 100,
			MegaPts:  5,
		},
		Report: ReportConfig{
			ModelName:       "opencode",
			MaxMoveDisplay:  25,
			MaxPVPEntries:   30,
			MaxTransferCand: 30,
			MaxTeamSize:     3,
		},
		PVP: PVPConfig{
			LeagueCaps: map[string]int{
				"GL": 1500,
				"UL": 2500,
				"ML": 0,
			},
		},
		Meta: MetaConfig{
			TopN: 20,
		},
	}
}

func Load(projectRoot string) *Config {
	cfg := Default()
	cfgPath := filepath.Join(projectRoot, cfg.Paths.Data, "config.json")

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		Global = cfg
		return cfg
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		Global = cfg
		return cfg
	}

	Global = cfg
	return cfg
}
