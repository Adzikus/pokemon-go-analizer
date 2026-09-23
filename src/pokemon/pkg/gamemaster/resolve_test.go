package gamemaster

import (
	"testing"

	"github.com/adzik/pokemon/model"
)

func TestNormalizeNameExtended(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"Deoxys - Defense", "deoxys_defense"},
		{"Deoxys - Attack", "deoxys_attack"},
		{"Deoxys - Speed", "deoxys_speed"},
		{"Rotom - Heat", "rotom_heat"},
		{"Rotom - Wash", "rotom_wash"},
		{"Burmy - Sandy", "burmy_sandy"},
		{"Mr. Mime", "mr._mime"},
		{"Ho-Oh", "ho-oh"},
		{"Porygon-Z", "porygon-z"},
		{"Jangmo-o", "jangmo-o"},
		{"Mime Jr.", "mime_jr."},
		{"  Bulbasaur  ", "bulbasaur"},
		{"BULBASAUR", "bulbasaur"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := NormalizeName(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestResolveSpeciesExactMatch(t *testing.T) {
	gmMap := map[string]*model.GamemasterPokemon{
		"bulbasaur": {SpeciesID: "bulbasaur", SpeciesName: "Bulbasaur"},
		"charmander": {SpeciesID: "charmander", SpeciesName: "Charmander"},
	}

	got := ResolveSpecies("bulbasaur", gmMap)
	if got != "bulbasaur" {
		t.Errorf("ResolveSpecies(bulbasaur) = %q, want bulbasaur", got)
	}
}

func TestResolveSpeciesCaseInsensitive(t *testing.T) {
	gmMap := map[string]*model.GamemasterPokemon{
		"bulbasaur": {SpeciesID: "bulbasaur", SpeciesName: "Bulbasaur"},
	}

	got := ResolveSpecies("Bulbasaur", gmMap)
	if got != "bulbasaur" {
		t.Errorf("ResolveSpecies(Bulbasaur) = %q, want bulbasaur", got)
	}
}

func TestResolveSpeciesNotFound(t *testing.T) {
	gmMap := map[string]*model.GamemasterPokemon{
		"bulbasaur": {SpeciesID: "bulbasaur", SpeciesName: "Bulbasaur"},
	}

	// Unknown species → normalized form
	got := ResolveSpecies("Unknown Pokemon", gmMap)
	if got != "unknown_pokemon" {
		t.Errorf("ResolveSpecies(Unknown Pokemon) = %q, want unknown_pokemon", got)
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"123", 123},
		{"0", 0},
		{"-5", -5},
		{"", 0},
		{"  42  ", 42},
		{"abc", 0},
		{"12.5", 0},
		{"1000000", 1000000},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParseInt(tt.input)
			if got != tt.want {
				t.Errorf("ParseInt(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}
