package structs

type LiveInfo struct {
	IsWarmUp          bool     `json:"isWarmUp"`
	WarmUpRound       *int     `json:"warmUpRound,omitempty"`
	WarmUpTotalRounds *int     `json:"warmUpTotalRounds,omitempty"`
	Mode              string   `json:"mode"`
	UseTeams          bool     `json:"useTeams"`
	CurrentMap        string   `json:"currentMap"`
	PointsLimit       *int     `json:"pointsLimit,omitempty"`
	RoundsLimit       *int     `json:"roundsLimit,omitempty"`
	MapLimit          *int     `json:"mapLimit,omitempty"`
	Maps              []string `json:"maps"`

	Teams   map[int]Team           `json:"teams,omitempty"`
	Players map[string]PlayerRound `json:"players,omitempty"`

	ActiveRound ActiveRound `json:"activeRound"`
}

type Team struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	RoundPoints int    `json:"roundPoints"`
	MatchPoints int    `json:"matchPoints"`
}

type PlayerRound struct {
	Login           string `json:"login"`
	AccountId       string `json:"accountId"`
	Name            string `json:"name"`
	Team            int    `json:"team"`
	Rank            int    `json:"rank"`
	Finalist        bool   `json:"finalist"`
	Winner          bool   `json:"winner"`
	RoundPoints     int    `json:"roundPoints"`
	MatchPoints     int    `json:"matchPoints"`
	BestTime        int    `json:"bestTime"`
	BestCheckpoints []int  `json:"bestCheckpoints"`
	PrevTime        int    `json:"prevTime"`
	PrevCheckpoints []int  `json:"prevCheckpoints"`
}

type ActiveRound struct {
	Players map[string]PlayerWaypoint `json:"players,omitempty"`
}

type PlayerWaypoint struct {
	Login       string `json:"login"`
	AccountId   string `json:"accountId"`
	Time        int    `json:"time"`
	HasFinished bool   `json:"hasFinished"`
	HasGivenUp  bool   `json:"hasGivenUp"`
	IsFinalist  bool   `json:"isFinalist"`
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

type UseTeams struct {
	ResponseId string `json:"responseid"`
	Teams      bool   `json:"teams"`
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
