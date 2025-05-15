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
