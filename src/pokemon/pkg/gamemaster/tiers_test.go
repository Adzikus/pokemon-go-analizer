package gamemaster

import (
	"testing"

	"github.com/adzik/pokemon/config"
)

func init() {
	config.Global = config.Default()
}

// Config for BuildMetaTiers tests:
//
//	config.Global.MetaTiers = MetaTiersConfig{ S: 10, A: 20, B: 30 }
//
// Meaning: pos 0-9 → S, 10-19 → A, 20-29 → B, 30+ → C
// Raid loop only assigns S (pos<10) and A (pos<10-19), no B/C.

func TestRankToTier(t *testing.T) {
	// config.Global.Tiers = {S:20, Ap:50, A:100, Bp:150, B:200, C:250}
	tests := []struct {
		pos  int
		want string
	}{
		{0, "S"}, {1, "S"}, {19, "S"},       // < 20 → S
		{20, "A+"}, {30, "A+"}, {49, "A+"},  // 20-49 → A+
		{50, "A"}, {75, "A"}, {99, "A"},      // 50-99 → A
		{100, "B+"}, {125, "B+"}, {149, "B+"},// 100-149 → B+
		{150, "B"}, {175, "B"}, {199, "B"},   // 150-199 → B
		{200, "C"}, {225, "C"}, {249, "C"},   // 200-249 → C
		{250, "D"}, {300, "D"},               // 250+ → D
	}
	for _, tt := range tests {
		got := RankToTier(tt.pos)
		if got != tt.want {
			t.Errorf("RankToTier(%d) = %q, want %q", tt.pos, got, tt.want)
		}
	}
}

func TestTierToScore(t *testing.T) {
	tests := []struct {
		tier string
		want float64
	}{
		{"S", 20}, {"A+", 17}, {"A", 14}, {"B+", 11},
		{"B", 8}, {"C", 4}, {"D", 0}, {"", 0}, {"X", 0},
	}
	for _, tt := range tests {
		got := TierToScore(tt.tier)
		if got != tt.want {
			t.Errorf("TierToScore(%q) = %f, want %f", tt.tier, got, tt.want)
		}
	}
}

func TestTierRank(t *testing.T) {
	tests := []struct {
		tier string
		want int
	}{
		{"S", 6}, {"A+", 5}, {"A", 4}, {"B+", 3},
		{"B", 2}, {"C", 1}, {"D", 0}, {"", 0},
	}
	for _, tt := range tests {
		got := TierRank(tt.tier)
		if got != tt.want {
			t.Errorf("TierRank(%q) = %d, want %d", tt.tier, got, tt.want)
		}
	}
}

func TestBestTier(t *testing.T) {
	tests := []struct {
		a, b, want string
	}{
		{"S", "A", "S"}, {"A", "S", "S"}, {"S", "S", "S"},
		{"B", "A", "A"}, {"A", "B", "A"},
		{"C", "D", "C"}, {"D", "C", "C"},
		{"", "S", "S"}, {"S", "", "S"},
		{"", "", ""},
	}
	for _, tt := range tests {
		got := BestTier(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("BestTier(%q, %q) = %q, want %q", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestBuildMetaTiers_RankToTierMappings(t *testing.T) {
	// MetaTiers config: S=10, A=20, B=30
	// Each Pokemon appears once in GL at a known position.
	// We verify that the correct tier is assigned based on position.
	tests := []struct {
		name     string // species name
		pos      int    // position in GreatLeague list
		wantTier string
	}{
		{"bulbasaur", 0, "S"},   // pos 0  < 10 → S
		{"charmander", 5, "S"},  // pos 5  < 10 → S
		{"squirtle", 9, "S"},    // pos 9  < 10 → S (last S)
		{"pikachu", 10, "A"},    // pos 10 < 20 → A (first A)
		{"eevee", 15, "A"},      // pos 15 < 20 → A
		{"snorlax", 19, "A"},    // pos 19 < 20 → A (last A)
		{"dragonite", 20, "B"},  // pos 20 < 30 → B (first B)
		{"tyranitar", 25, "B"},  // pos 25 < 30 → B
		{"metagross", 29, "B"},  // pos 29 < 30 → B (last B)
		{"rayquaza", 30, "C"},   // pos 30 >= 30 → C (first C)
		{"mewtwo", 50, "C"},     // pos 50 >= 30 → C
	}

	// Build GL list with known positions
	gl := make([]string, 60)
	for _, tt := range tests {
		gl[tt.pos] = tt.name
	}

	meta := &MetaData{}
	meta.Rankings.GreatLeague = gl
	meta.Rankings.UltraLeague = []string{}
	meta.Rankings.MasterLeague = []string{}
	meta.Rankings.Raid = []string{}

	tiers := BuildMetaTiers(meta)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tiers[tt.name]
			if got != tt.wantTier {
				t.Errorf("%s at pos %d: got tier %q, want %q", tt.name, tt.pos, got, tt.wantTier)
			}
		})
	}
}

func TestBuildMetaTiers_SkipAlreadyS(t *testing.T) {
	// Scenario: mewtwo is S in GL (pos 0), then appears again in ML at pos 15.
	// ML pos 15 < 20 → would be A, but skip logic should keep it as S.
	gl := []string{"mewtwo"}
	ml := make([]string, 20)
	ml[0] = "mewtwo"   // ML pos 0 → S
	ml[15] = "mewtwo"  // ML pos 15 → would be A, but mewtwo already S

	meta := &MetaData{}
	meta.Rankings.GreatLeague = gl
	meta.Rankings.UltraLeague = []string{}
	meta.Rankings.MasterLeague = ml
	meta.Rankings.Raid = []string{}

	tiers := BuildMetaTiers(meta)

	if tiers["mewtwo"] != "S" {
		t.Errorf("mewtwo = %q, want S (skip should preserve S from GL)", tiers["mewtwo"])
	}
}

func TestBuildMetaTiers_RaidSkipS(t *testing.T) {
	// Scenario: pikachu is S from GL (pos 5), then appears in Raid at pos 3.
	// Raid pos 3 < 10 → S, but pikachu already S → skip.
	gl := []string{"bulbasaur", "charmander", "squirtle", "pikachu", "eevee", "snorlax"}
	raid := []string{"rayquaza", "mewtwo", "dragonite", "pikachu"}

	meta := &MetaData{}
	meta.Rankings.GreatLeague = gl
	meta.Rankings.UltraLeague = []string{}
	meta.Rankings.MasterLeague = []string{}
	meta.Rankings.Raid = raid

	tiers := BuildMetaTiers(meta)

	if tiers["pikachu"] != "S" {
		t.Errorf("pikachu = %q, want S (skip in raid loop)", tiers["pikachu"])
	}
}

func TestBuildMetaTiers_RaidOnlyAssignsSAndA(t *testing.T) {
	// Scenario: only Raid rankings, no GL/UL/ML.
	// Raid loop only assigns S (pos<10) and A (pos<10-19).
	// pos >= 20 → NOT assigned by raid loop.
	raid := make([]string, 25)
	raid[0] = "rayquaza"  // pos 0 → S
	raid[9] = "mewtwo"    // pos 9 → S
	raid[10] = "groudon"  // pos 10 → A
	raid[19] = "kyogre"   // pos 19 → A
	raid[20] = "dialga"   // pos 20 → NOT assigned (raid only S/A)

	meta := &MetaData{}
	meta.Rankings.GreatLeague = []string{}
	meta.Rankings.UltraLeague = []string{}
	meta.Rankings.MasterLeague = []string{}
	meta.Rankings.Raid = raid

	tiers := BuildMetaTiers(meta)

	if tiers["rayquaza"] != "S" {
		t.Errorf("rayquaza (raid pos 0) = %q, want S", tiers["rayquaza"])
	}
	if tiers["mewtwo"] != "S" {
		t.Errorf("mewtwo (raid pos 9) = %q, want S", tiers["mewtwo"])
	}
	if tiers["groudon"] != "A" {
		t.Errorf("groudon (raid pos 10) = %q, want A", tiers["groudon"])
	}
	if tiers["kyogre"] != "A" {
		t.Errorf("kyogre (raid pos 19) = %q, want A", tiers["kyogre"])
	}
	if _, ok := tiers["dialga"]; ok {
		t.Errorf("dialga (raid pos 20) should NOT be assigned (raid loop only handles S/A)")
	}
}

func TestBuildMetaTiers_ShadowPropagation(t *testing.T) {
	// Scenario: mewtwo is S from GL. No shadow/mega in any ranking list.
	// They should inherit S from base "mewtwo" via propagation.
	gl := []string{"mewtwo"}

	meta := &MetaData{}
	meta.Rankings.GreatLeague = gl
	meta.Rankings.UltraLeague = []string{}
	meta.Rankings.MasterLeague = []string{}
	meta.Rankings.Raid = []string{}

	tiers := BuildMetaTiers(meta)

	// Base species gets tier from rankings
	if tiers["mewtwo"] != "S" {
		t.Errorf("mewtwo = %q, want S", tiers["mewtwo"])
	}
	// Variants inherit from base
	if tiers["mewtwo_shadow"] != "S" {
		t.Errorf("mewtwo_shadow = %q, want S (inherited from mewtwo)", tiers["mewtwo_shadow"])
	}
	if tiers["mewtwo_mega"] != "S" {
		t.Errorf("mewtwo_mega = %q, want S (inherited from mewtwo)", tiers["mewtwo_mega"])
	}
	if tiers["mewtwo_mega_x"] != "S" {
		t.Errorf("mewtwo_mega_x = %q, want S (inherited from mewtwo)", tiers["mewtwo_mega_x"])
	}
	if tiers["mewtwo_mega_y"] != "S" {
		t.Errorf("mewtwo_mega_y = %q, want S (inherited from mewtwo)", tiers["mewtwo_mega_y"])
	}
}

func TestBuildMetaTiers_PropagationDoesNotOverwriteExplicit(t *testing.T) {
	// Scenario: mewtwo is S (pos 0), mewtwo_shadow is A (pos 15), mewtwo_mega not in list.
	// mewtwo_shadow is already explicitly set → propagation should not overwrite it.
	gl := make([]string, 20)
	gl[0] = "mewtwo"
	gl[15] = "mewtwo_shadow"

	meta := &MetaData{}
	meta.Rankings.GreatLeague = gl
	meta.Rankings.UltraLeague = []string{}
	meta.Rankings.MasterLeague = []string{}
	meta.Rankings.Raid = []string{}

	tiers := BuildMetaTiers(meta)

	if tiers["mewtwo"] != "S" {
		t.Errorf("mewtwo = %q, want S", tiers["mewtwo"])
	}
	if tiers["mewtwo_shadow"] != "A" {
		t.Errorf("mewtwo_shadow = %q, want A (explicit from GL pos 15)", tiers["mewtwo_shadow"])
	}
	// mewtwo_mega not in any list → inherits from base
	if tiers["mewtwo_mega"] != "S" {
		t.Errorf("mewtwo_mega = %q, want S (inherited from mewtwo, best of base)", tiers["mewtwo_mega"])
	}
}

func TestBuildMetaTiers_CrossCategoryBoost(t *testing.T) {
	// Scenario: groudon is B in GL (pos 22) but S in Raid (pos 5).
	// Raid loop should upgrade groudon to S.
	gl := make([]string, 30)
	gl[22] = "groudon"

	raid := make([]string, 10)
	raid[5] = "groudon"

	meta := &MetaData{}
	meta.Rankings.GreatLeague = gl
	meta.Rankings.UltraLeague = []string{}
	meta.Rankings.MasterLeague = []string{}
	meta.Rankings.Raid = raid

	tiers := BuildMetaTiers(meta)

	if tiers["groudon"] != "S" {
		t.Errorf("groudon = %q, want S (Raid pos 5 upgrades B from GL)", tiers["groudon"])
	}
}

func TestBuildMetaTiersByCategory(t *testing.T) {
	meta := &MetaData{}
	meta.Rankings.GreatLeague = []string{"mewtwo", "groudon", "kyogre"}
	meta.Rankings.UltraLeague = []string{"dialga", "giratina", "mewtwo"}
	meta.Rankings.MasterLeague = []string{"rayquaza", "mewtwo"}
	meta.Rankings.Raid = []string{"rayquaza", "kyogre"}

	cats := BuildMetaTiersByCategory(meta)

	// GL: mewtwo rank 0 → S
	if cats.GL["mewtwo"] != "S" {
		t.Errorf("GL[mewtwo] = %q, want S", cats.GL["mewtwo"])
	}
	if cats.GLRank["mewtwo"] != 0 {
		t.Errorf("GLRank[mewtwo] = %d, want 0", cats.GLRank["mewtwo"])
	}
	// UL: dialga rank 0 → S
	if cats.UL["dialga"] != "S" {
		t.Errorf("UL[dialga] = %q, want S", cats.UL["dialga"])
	}
	// ML: rayquaza rank 0 → S
	if cats.ML["rayquaza"] != "S" {
		t.Errorf("ML[rayquaza] = %q, want S", cats.ML["rayquaza"])
	}
	// Raid: rayquaza rank 0 → S
	if cats.Raid["rayquaza"] != "S" {
		t.Errorf("Raid[rayquaza] = %q, want S", cats.Raid["rayquaza"])
	}
}

func TestBestMetaTier(t *testing.T) {
	cats := MetaTiersByCategory{
		Raid: map[string]string{"mewtwo": "A"},
		GL:   map[string]string{"mewtwo": "S"},
		UL:   map[string]string{"mewtwo": "B"},
		ML:   map[string]string{"mewtwo": "A"},
	}

	got := cats.BestMetaTier("mewtwo")
	if got != "S" {
		t.Errorf("BestMetaTier(mewtwo) = %q, want S (best from GL)", got)
	}

	// Unknown species → D
	got = cats.BestMetaTier("unknown")
	if got != "D" {
		t.Errorf("BestMetaTier(unknown) = %q, want D", got)
	}
}
