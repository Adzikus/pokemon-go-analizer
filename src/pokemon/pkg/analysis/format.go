package analysis

import (
	"strings"

	"github.com/adzik/pokemon/model"
)

func BoolStr(b bool) string {
	if b {
		return "TAK"
	}
	return "NIE"
}

func StrOrDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func TierOrDash(p model.Pokemon) string {
	for _, tier := range []string{p.RaidTier, p.MLTier, p.ULTier, p.GLTier} {
		if tier != "" {
			return tier
		}
	}
	return "-"
}

func TierContext(p model.Pokemon) string {
	return "R:" + p.RaidTier + " GL:" + p.GLTier + " UL:" + p.ULTier + " ML:" + p.MLTier
}

func FormatTags(p model.Pokemon) string {
	var tags []string
	if p.IsDynamax {
		tags = append(tags, "Dynamax")
	}
	if p.IsPurified {
		tags = append(tags, "Purified")
	}
	if p.IsTraded {
		tags = append(tags, "Traded")
	}
	return strings.Join(tags, ", ")
}

func FormatLucky(p model.Pokemon) string {
	if p.IsLucky {
		return "OK"
	}
	return ""
}

func FormatShadow(p model.Pokemon) string {
	if p.IsShadow {
		return "\U0001f47b"
	}
	return ""
}

func FormatEmojis(p model.Pokemon) string {
	var tags []string
	if p.IsShiny {
		tags = append(tags, "\u2728")
	}
	if p.IsLucky {
		tags = append(tags, "\U0001f340")
	}
	if p.IsShadow {
		tags = append(tags, "\U0001f47b")
	}
	if p.IsDynamax {
		tags = append(tags, "\u2b50")
	}
	return strings.Join(tags, "")
}

func PurposeLabel(p model.Pokemon) string {
	purpose := p.Purpose
	if purpose == "" {
		purpose = "-"
	}
	return purpose
}

func SafeMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Pct(a, b int) float64 {
	if b == 0 {
		return 0
	}
	return float64(a) / float64(b) * 100
}

func FilterBool(c []model.Pokemon, fn func(model.Pokemon) bool) []model.Pokemon {
	var r []model.Pokemon
	for _, p := range c {
		if fn(p) {
			r = append(r, p)
		}
	}
	return r
}
