package model

// MetaRankings represents the meta_rankings.json structure.
type MetaRankings struct {
	Rankings  MetaRankingsList `json:"rankings"`
	FetchedAt string           `json:"fetched_at"`
}

// MetaRankingsList contains ranked speciesId lists per league.
type MetaRankingsList struct {
	GreatLeagueOverall  []string `json:"great_league_overall"`
	UltraLeagueOverall  []string `json:"ultra_league_overall"`
	MasterLeagueOverall []string `json:"master_league_overall"`
	RaidAttackers       []string `json:"raid_attackers"`
}
