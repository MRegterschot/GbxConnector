package structs

type MapEvent struct {
	Count int `json:"count"`
	Valid int `json:"valid"`
	Map   Map `json:"map"`
}

type Map struct {
	UId            string `json:"uid"`
	Name           string `json:"name"`
	Filename       string `json:"filename"`
	Author         string `json:"author"`
	AuthorNickname string `json:"authornickname"`
	Environment    string `json:"environment"`
	Mood           string `json:"mood"`
	BronzeTime     int    `json:"bronzetime"`
	SilverTime     int    `json:"silvertime"`
	GoldTime       int    `json:"goldtime"`
	AuthorTime     int    `json:"authortime"`
	CopperPrice    int    `json:"copperprice"`
	LapRace        bool   `json:"laprace"`
	NbLaps         int    `json:"nblaps"`
	MapType        string `json:"maptype"`
	MapStyle       string `json:"mapstyle"`
}
