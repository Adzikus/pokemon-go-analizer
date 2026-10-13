package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adzik/pokemon/config"
	"github.com/adzik/pokemon/model"
	"github.com/adzik/pokemon/pkg/gamemaster"
)

func testCollection() []model.Pokemon {
	return []model.Pokemon{
		{
			PokemonRaw: model.PokemonRaw{
				UID: "P-001", Species: "Mewtwo", Nick: "",
				Type1: "psychic", Type2: "",
				CP: 3000, IVAtk: 15, IVDef: 15, IVHP: 14,
				FastMove: "Confusion", Charged1: "Shadow Ball", Charged2: "Focus Blast",
				Purpose: "rajdy",
			},
			SpeciesID:  "mewtwo",
			IVPct:      97.8,
			IsLucky:    true,
			IsShiny:    true,
			BaseStats:  model.BaseStats{Atk: 300, Def: 182, HP: 214},
			FamilyID:   "mewtwo",
			RaidTier:   "S",
			ULTier:     "A+",
			MLTier:     "S",
			DevScore:   92,
			MaxCP:      4178,
		},
		{
			PokemonRaw: model.PokemonRaw{
				UID: "P-002", Species: "Bulbasaur", Nick: "Groszek",
				Type1: "grass", Type2: "poison",
				CP: 1500, IVAtk: 15, IVDef: 14, IVHP: 15,
				FastMove: "Vine Whip", Charged1: "Sludge Bomb", Charged2: "Seed Bomb",
				Purpose: "pvp",
			},
			SpeciesID:  "bulbasaur",
			IVPct:      97.8,
			IsShadow:   true,
			IsDynamax:  true,
			BaseStats:  model.BaseStats{Atk: 118, Def: 111, HP: 128},
			FamilyID:   "bulbasaur",
			RaidTier:   "A",
			GLTier:     "S",
			DevScore:   85,
			MaxCP:      1491,
		},
		{
			PokemonRaw: model.PokemonRaw{
				UID: "P-003", Species: "Charizard", Nick: "",
				Type1: "fire", Type2: "flying",
				CP: 2500, IVAtk: 15, IVDef: 15, IVHP: 15,
				FastMove: "Ember", Charged1: "Blast Burn", Charged2: "Dragon Claw",
				Purpose: "rajdy",
			},
			SpeciesID:  "charizard",
			IVPct:      100.0,
			IsLucky:    true,
			BaseStats:  model.BaseStats{Atk: 223, Def: 173, HP: 186},
			FamilyID:   "charizard",
			RaidTier:   "A",
			GLTier:     "B",
			ULTier:     "A",
			DevScore:   88,
			MaxCP:      2967,
		},
		{
			PokemonRaw: model.PokemonRaw{
				UID: "P-004", Species: "Pidgey", Nick: "",
				Type1: "normal", Type2: "flying",
				CP: 100, IVAtk: 5, IVDef: 5, IVHP: 5,
				FastMove: "Tackle", Charged1: "Twister", Charged2: "",
				Purpose: "transfer",
			},
			SpeciesID:  "pidgey",
			IVPct:      22.2,
			BaseStats:  model.BaseStats{Atk: 85, Def: 73, HP: 121},
			FamilyID:   "pidgey",
			RaidTier:   "D",
			GLTier:     "D",
			DevScore:   8,
			MaxCP:      736,
		},
	}
}

func testGamemasterMap() map[string]*model.GamemasterPokemon {
	return map[string]*model.GamemasterPokemon{
		"mewtwo":    {SpeciesID: "mewtwo", SpeciesName: "Mewtwo", Family: &model.Family{ID: "mewtwo"}},
		"bulbasaur": {SpeciesID: "bulbasaur", SpeciesName: "Bulbasaur", FastMoves: []string{"VINE_WHIP", "TACKLE"}, ChargedMoves: []string{"SLUDGE_BOMB", "SEED_BOMB"}, Family: &model.Family{ID: "bulbasaur"}},
		"charizard": {SpeciesID: "charizard", SpeciesName: "Charizard", Family: &model.Family{ID: "charizard"}},
		"pidgey":    {SpeciesID: "pidgey", SpeciesName: "Pidgey", Family: &model.Family{ID: "pidgey"}},
	}
}

func testMoveMap() map[string]*gamemaster.MoveEntry {
	return map[string]*gamemaster.MoveEntry{
		"CONFUSION":   {MoveID: "CONFUSION", Name: "Confusion", Type: "psychic", Power: 20, Cooldown: 600, EnergyGain: 15},
		"SHADOW_BALL": {MoveID: "SHADOW_BALL", Name: "Shadow Ball", Type: "ghost", Power: 100, Cooldown: 3000, Energy: 100},
		"FOCUS_BLAST": {MoveID: "FOCUS_BLAST", Name: "Focus Blast", Type: "fighting", Power: 140, Cooldown: 3500, Energy: 100},
		"VINE_WHIP":   {MoveID: "VINE_WHIP", Name: "Vine Whip", Type: "grass", Power: 7, Cooldown: 600, EnergyGain: 12},
		"TACKLE":      {MoveID: "TACKLE", Name: "Tackle", Type: "normal", Power: 5, Cooldown: 500, EnergyGain: 5},
		"SLUDGE_BOMB": {MoveID: "SLUDGE_BOMB", Name: "Sludge Bomb", Type: "poison", Power: 80, Cooldown: 2300, Energy: 50},
		"SEED_BOMB":   {MoveID: "SEED_BOMB", Name: "Seed Bomb", Type: "grass", Power: 65, Cooldown: 2100, Energy: 45},
		"EMBER":       {MoveID: "EMBER", Name: "Ember", Type: "fire", Power: 10, Cooldown: 1000, EnergyGain: 10},
		"BLAST_BURN":  {MoveID: "BLAST_BURN", Name: "Blast Burn", Type: "fire", Power: 110, Cooldown: 3300, Energy: 100},
		"DRAGON_CLAW": {MoveID: "DRAGON_CLAW", Name: "Dragon Claw", Type: "dragon", Power: 50, Cooldown: 1700, Energy: 35},
		"TWISTER":     {MoveID: "TWISTER", Name: "Twister", Type: "dragon", Power: 45, Cooldown: 2500, Energy: 45},
	}
}

func testMetaTiersByCategory() gamemaster.MetaTiersByCategory {
	return gamemaster.MetaTiersByCategory{
		Raid: map[string]string{"mewtwo": "S", "charizard": "A", "bulbasaur": "A", "pidgey": "D"},
		GL:   map[string]string{"bulbasaur": "S"},
		UL:   map[string]string{"mewtwo": "A+", "charizard": "A"},
		ML:   map[string]string{"mewtwo": "S"},
	}
}

func testMetaData() *gamemaster.MetaData {
	md := &gamemaster.MetaData{}
	md.Rankings.Raid = []string{"mewtwo", "charizard", "bulbasaur"}
	md.Rankings.GreatLeague = []string{"bulbasaur"}
	md.Rankings.UltraLeague = []string{"mewtwo", "charizard"}
	md.Rankings.MasterLeague = []string{"mewtwo"}
	return md
}

func testGamemasterData() *gamemaster.GamemasterData {
	return &gamemaster.GamemasterData{
		Pokemon: []model.GamemasterPokemon{
			{SpeciesID: "mewtwo", SpeciesName: "Mewtwo"},
			{SpeciesID: "bulbasaur", SpeciesName: "Bulbasaur"},
			{SpeciesID: "charizard", SpeciesName: "Charizard"},
			{SpeciesID: "pidgey", SpeciesName: "Pidgey"},
		},
	}
}

// --- Helper tests ---

func TestTierContext(t *testing.T) {
	p := model.Pokemon{
		RaidTier: "S",
		GLTier:   "A+",
		ULTier:   "B",
		MLTier:   "",
	}
	got := tierContext(p)
	want := "R:S GL:A+ UL:B ML:-"
	if got != want {
		t.Errorf("tierContext() = %q, want %q", got, want)
	}
}

func TestLeagueScore(t *testing.T) {
	p := model.Pokemon{GLScore: 80, ULScore: 70, MLScore: 90}

	tests := []struct {
		league string
		want   float64
	}{
		{"GL", 80},
		{"UL", 70},
		{"ML", 90},
		{"unknown", 0},
	}

	for _, tt := range tests {
		got := leagueScore(p, tt.league)
		if got != tt.want {
			t.Errorf("leagueScore(%s) = %f, want %f", tt.league, got, tt.want)
		}
	}
}

// --- BuildReportData tests ---

func TestBuildReportData_Section1(t *testing.T) {
	config.Global = config.Default()
	c := testCollection()
	gm := testGamemasterData()
	gmMap := testGamemasterMap()
	moveMap := testMoveMap()
	metaData := testMetaData()
	metaCats := testMetaTiersByCategory()
	metaTiers := gamemaster.BuildMetaTiers(metaData)

	data := buildReportData(c, "Test", gm, gmMap, moveMap, metaData, metaTiers, metaCats, nil)

	if data.Stats.Total != 4 {
		t.Errorf("Stats.Total = %d, want 4", data.Stats.Total)
	}
	if len(data.TopHundo) != 1 {
		t.Errorf("TopHundo = %d, want 1", len(data.TopHundo))
	}
	if data.TopHundo[0].Species != "Charizard" {
		t.Errorf("TopHundo[0] = %s, want Charizard", data.TopHundo[0].Species)
	}
	if len(data.TopCP) != 4 {
		t.Errorf("TopCP = %d, want 4", len(data.TopCP))
	}
	if data.TopCP[0].Species != "Mewtwo" {
		t.Errorf("TopCP[0] = %s, want Mewtwo", data.TopCP[0].Species)
	}
}

func TestBuildReportData_MetaComparison(t *testing.T) {
	config.Global = config.Default()
	c := testCollection()
	gm := testGamemasterData()
	gmMap := testGamemasterMap()
	moveMap := testMoveMap()
	metaData := testMetaData()
	metaCats := testMetaTiersByCategory()
	metaTiers := gamemaster.BuildMetaTiers(metaData)

	data := buildReportData(c, "Test", gm, gmMap, moveMap, metaData, metaTiers, metaCats, nil)

	if len(data.MetaRows) != 3 {
		t.Errorf("MetaRows = %d, want 3", len(data.MetaRows))
	}
	if data.MetaTotal != 3 {
		t.Errorf("MetaTotal = %d, want 3", data.MetaTotal)
	}
}

func TestBuildReportData_Transfers(t *testing.T) {
	config.Global = config.Default()
	c := testCollection()
	gm := testGamemasterData()
	gmMap := testGamemasterMap()
	moveMap := testMoveMap()
	metaData := testMetaData()
	metaCats := testMetaTiersByCategory()
	metaTiers := gamemaster.BuildMetaTiers(metaData)

	data := buildReportData(c, "Test", gm, gmMap, moveMap, metaData, metaTiers, metaCats, nil)

	total := len(data.SafeTransfer) + len(data.TradeTransfer) + len(data.SoftKeepTransfer)
	if total == 0 {
		t.Error("expected some transfer candidates")
	}
}

func TestBuildReportData_LuckyTargets(t *testing.T) {
	config.Global = config.Default()
	c := testCollection()
	gm := testGamemasterData()
	gmMap := testGamemasterMap()
	moveMap := testMoveMap()
	metaData := testMetaData()
	metaCats := testMetaTiersByCategory()
	metaTiers := gamemaster.BuildMetaTiers(metaData)

	data := buildReportData(c, "Test", gm, gmMap, moveMap, metaData, metaTiers, metaCats, nil)

	if len(data.LuckyTargets) == 0 {
		t.Error("expected some lucky targets")
	}
}

func TestBuildReportData_Section6(t *testing.T) {
	config.Global = config.Default()
	c := testCollection()
	gm := testGamemasterData()
	gmMap := testGamemasterMap()
	moveMap := testMoveMap()
	metaData := testMetaData()
	metaCats := testMetaTiersByCategory()
	metaTiers := gamemaster.BuildMetaTiers(metaData)

	data := buildReportData(c, "Test", gm, gmMap, moveMap, metaData, metaTiers, metaCats, nil)

	if len(data.HundoDetail) != 1 {
		t.Errorf("HundoDetail = %d, want 1", len(data.HundoDetail))
	}
	if len(data.ShadowDetail) != 1 {
		t.Errorf("ShadowDetail = %d, want 1", len(data.ShadowDetail))
	}
	if len(data.ShinyDetail) != 1 {
		t.Errorf("ShinyDetail = %d, want 1", len(data.ShinyDetail))
	}
}

func TestBuildReportData_EmptyCollection(t *testing.T) {
	config.Global = config.Default()
	gm := testGamemasterData()
	gmMap := testGamemasterMap()
	moveMap := testMoveMap()
	metaData := testMetaData()
	metaCats := testMetaTiersByCategory()
	metaTiers := gamemaster.BuildMetaTiers(metaData)

	data := buildReportData(nil, "Test", gm, gmMap, moveMap, metaData, metaTiers, metaCats, nil)

	if data.Stats.Total != 0 {
		t.Errorf("Stats.Total = %d, want 0", data.Stats.Total)
	}
	if len(data.TopHundo) != 0 {
		t.Errorf("TopHundo = %d, want 0", len(data.TopHundo))
	}
	if len(data.FullRanking.Rows) != 0 {
		t.Errorf("FullRanking.Rows = %d, want 0", len(data.FullRanking.Rows))
	}
}

// --- Full report integration test ---

func setupTestProjectRoot(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	templatesDir := filepath.Join(tmpDir, "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Copy template from real location
	src, err := os.ReadFile("../../../templates/report.md.tmpl")
	if err != nil {
		t.Fatalf("cannot read source template: %v", err)
	}
	if err := os.WriteFile(filepath.Join(templatesDir, "report.md.tmpl"), src, 0644); err != nil {
		t.Fatal(err)
	}
	return tmpDir
}

func TestWriteMarkdownReport_Full(t *testing.T) {
	config.Global = config.Default()
	c := testCollection()
	gm := testGamemasterData()
	gmMap := testGamemasterMap()
	moveMap := testMoveMap()
	metaData := testMetaData()
	metaCats := testMetaTiersByCategory()
	metaTiers := gamemaster.BuildMetaTiers(metaData)

	projectRoot := setupTestProjectRoot(t)
	mdPath := filepath.Join(projectRoot, "test_report.md")

	writeMarkdownReport(mdPath, c, "Test", gm, gmMap, moveMap, metaData, metaTiers, metaCats, projectRoot, nil)

	data, err := os.ReadFile(mdPath)
	if err != nil {
		t.Fatalf("failed to read report: %v", err)
	}
	out := string(data)

	for i := 1; i <= 7; i++ {
		if !strings.Contains(out, "## 1. Ogolne") && i == 1 {
			t.Error("missing section 1")
		}
	}

	if !strings.HasPrefix(out, "# Analiza Pokemon") {
		t.Error("missing report header")
	}
	if !strings.Contains(out, "---") {
		t.Error("missing horizontal rule separator")
	}

	for _, name := range []string{"Mewtwo", "Bulbasaur", "Charizard", "Pidgey"} {
		if !strings.Contains(out, name) {
			t.Errorf("missing %s in report", name)
		}
	}

	if !strings.Contains(out, "## 2. Rajdy") {
		t.Error("missing section 2")
	}
	if !strings.Contains(out, "## 3. Zestawy do PvP") {
		t.Error("missing section 3")
	}
	if !strings.Contains(out, "## 4. Kandydaci do Transferu") {
		t.Error("missing section 4")
	}
	if !strings.Contains(out, "## 5. Poszukiwani do statusu Lucky") {
		t.Error("missing section 5")
	}
	if !strings.Contains(out, "## 6. Szczegolowa analiza") {
		t.Error("missing section 6")
	}
	if !strings.Contains(out, "## 7. Pokemon-y do rozwoju") {
		t.Error("missing section 7")
	}
}

func TestWriteMarkdownReport_EmptyCollection(t *testing.T) {
	config.Global = config.Default()
	gm := testGamemasterData()
	gmMap := testGamemasterMap()
	moveMap := testMoveMap()
	metaData := testMetaData()
	metaCats := testMetaTiersByCategory()
	metaTiers := gamemaster.BuildMetaTiers(metaData)

	projectRoot := setupTestProjectRoot(t)
	mdPath := filepath.Join(projectRoot, "test_report.md")

	writeMarkdownReport(mdPath, nil, "Test", gm, gmMap, moveMap, metaData, metaTiers, metaCats, projectRoot, nil)

	data, err := os.ReadFile(mdPath)
	if err != nil {
		t.Fatalf("failed to read report: %v", err)
	}
	out := string(data)

	if !strings.Contains(out, "**Laczna ilosc Pokemon:** 0") {
		t.Error("should show 0 Pokemon")
	}
	if !strings.Contains(out, "Brak Pokemon") {
		t.Error("should show 'Brak Pokemon'")
	}
}

// --- JSON snapshot ---

func TestWriteJSONSnapshot(t *testing.T) {
	c := testCollection()
	tmpDir := t.TempDir()
	jsonPath := filepath.Join(tmpDir, "test.json")

	writeJSONSnapshot(jsonPath, c, "Test")

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("failed to read JSON: %v", err)
	}

	var snapshot map[string]interface{}
	if err := json.Unmarshal(data, &snapshot); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if snapshot["konto"] != "Test" {
		t.Errorf("konto = %v, want 'Test'", snapshot["konto"])
	}
	if snapshot["count"] != float64(4) {
		t.Errorf("count = %v, want 4", snapshot["count"])
	}
	if snapshot["generated"] == nil {
		t.Error("missing generated timestamp")
	}
	collection, ok := snapshot["collection"].([]interface{})
	if !ok || len(collection) != 4 {
		t.Error("collection should have 4 entries")
	}
}
