package gamemaster

import (
	"encoding/json"
	"os"
)

func LoadGamemaster(path string) (*GamemasterData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var gm GamemasterData
	if err := json.Unmarshal(data, &gm); err != nil {
		return nil, err
	}
	return &gm, nil
}

func LoadMeta(path string) (*MetaData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m MetaData
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}
