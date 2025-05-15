package structs

type PlayerWaypoint struct {
	Login       string `json:"login"`
	AccountId   string `json:"accountid"`
	Time        int    `json:"time"`
	HasFinished bool   `json:"hasFinished"`
	Checkpoint  int    `json:"checkpoint"`
}

type WarmUpStatus struct {
	ResponseId string `json:"responseid"`
	Available  bool   `json:"available"`
	Active     bool   `json:"active"`
}

type Scores struct {
	ResponseId string                `json:"responseid"`
	Section    string                `json:"section"`
	UseTeams   bool                  `json:"useteams"`
	WinnerTeam int                   `json:"winnerteam"`
	Teams      []Team                `json:"teams"`
	Players    []CallbackPlayerRound `json:"players"`
}

type CallbackPlayerRound struct {
	Login              string `json:"login"`
	AccountId          string `json:"accountid"`
	Name               string `json:"name"`
	Team               int    `json:"team"`
	Rank               int    `json:"rank"`
	RoundPoints        int    `json:"roundpoints"`
	MapPoints          int    `json:"mappoints"`
	MatchPoints        int    `json:"matchpoints"`
	BestRaceTime       int    `json:"bestracetime"`
	BestCheckpoints    []int  `json:"bestcheckpoints"`
	BestLapTime        int    `json:"bestlaptime"`
	BestLapCheckpoints []int  `json:"bestlapcheckpoints"`
	PrevRaceTime       int    `json:"prevracetime"`
	PrevCheckpoints    []int  `json:"prevcheckpoints"`
}
