package analysis

import (
	"testing"

	"github.com/adzik/pokemon/model"
)

func TestTierOrDash(t *testing.T) {
	tests := []struct {
		name string
		p    model.Pokemon
		want string
	}{
		{
			name: "has RaidTier",
			p:    model.Pokemon{RaidTier: "S", GLTier: "A", ULTier: "B", MLTier: "C"},
			want: "S",
		},
		{
			name: "no RaidTier, has MLTier",
			p:    model.Pokemon{GLTier: "A", MLTier: "A+"},
			want: "A+",
		},
		{
			name: "only GLTier",
			p:    model.Pokemon{GLTier: "B"},
			want: "B",
		},
		{
			name: "only ULTier",
			p:    model.Pokemon{ULTier: "C"},
			want: "C",
		},
		{
			name: "all empty",
			p:    model.Pokemon{},
			want: "-",
		},
		{
			name: "RaidTier empty, rest present — ML wins",
			p:    model.Pokemon{GLTier: "A", ULTier: "B", MLTier: "S"},
			want: "S", // loop order: Raid, ML, UL, GL → ML is first non-empty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TierOrDash(tt.p)
			if got != tt.want {
				t.Errorf("TierOrDash() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatTags(t *testing.T) {
	tests := []struct {
		name string
		p    model.Pokemon
		want string
	}{
		{
			name: "no tags",
			p:    model.Pokemon{},
			want: "",
		},
		{
			name: "Dynamax only",
			p:    model.Pokemon{IsDynamax: true},
			want: "Dynamax",
		},
		{
			name: "Purified only",
			p:    model.Pokemon{IsPurified: true},
			want: "Purified",
		},
		{
			name: "Traded only",
			p:    model.Pokemon{IsTraded: true},
			want: "Traded",
		},
		{
			name: "Dynamax + Purified",
			p:    model.Pokemon{IsDynamax: true, IsPurified: true},
			want: "Dynamax, Purified",
		},
		{
			name: "all three",
			p:    model.Pokemon{IsDynamax: true, IsPurified: true, IsTraded: true},
			want: "Dynamax, Purified, Traded",
		},
		{
			name: "Shiny ignored",
			p:    model.Pokemon{IsShiny: true},
			want: "",
		},
		{
			name: "Lucky ignored",
			p:    model.Pokemon{IsLucky: true},
			want: "",
		},
		{
			name: "Shadow ignored",
			p:    model.Pokemon{IsShadow: true},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatTags(tt.p)
			if got != tt.want {
				t.Errorf("FormatTags() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPct(t *testing.T) {
	tests := []struct {
		name   string
		a, b   int
		want   float64
	}{
		{"50 of 100", 50, 100, 50.0},
		{"0 of 100", 0, 100, 0.0},
		{"100 of 100", 100, 100, 100.0},
		{"1 of 3", 1, 3, 33.33333333333333},
		{"div by zero", 50, 0, 0.0},
		{"0 of 0", 0, 0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Pct(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("Pct(%d, %d) = %f, want %f", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
