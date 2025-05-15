package structs

import (
	"context"

	"github.com/MRegterschot/GbxRemoteGo/gbxclient"
)

type Server struct {
	Id          int     `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Host        string  `json:"host"`
	XMLRPCPort  int     `json:"xmlrpcPort"`
	User        string  `json:"user"`
	Pass        string  `json:"pass"`
	FMUrl       *string `json:"fmUrl,omitempty"`

	// Internal
	Info       *ServerInfo          `json:"-"`
	Client     *gbxclient.GbxClient `json:"-"`
	CancelFunc context.CancelFunc   `json:"-"`
	Ctx        context.Context      `json:"-"`
}

type ServerList []*Server

type ServerInfo struct {
	ActiveMap     string       `json:"-"`
	ActivePlayers []PlayerInfo `json:"-"`
	LiveInfo      *LiveInfo    `json:"-"`
}

type LiveInfo struct {
	IsWarmup    bool     `json:"isWarmup"`
	Mode        string   `json:"mode"`
	CurrentMap  string   `json:"currentMap"`
	PointsLimit *int     `json:"pointsLimit,omitempty"`
	RoundsLimit *int     `json:"roundsLimit,omitempty"`
	MapLimit    *int     `json:"mapLimit,omitempty"`
	Maps        []string `json:"maps"`

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
	RoundPoints     int    `json:"roundPoints"`
	MatchPoints     int    `json:"matchPoints"`
	BestTime        int    `json:"bestTime"`
	BestCheckpoints []int  `json:"bestCheckpoints"`
	PrevTime        int    `json:"prevTime"`
	PrevCheckpoints []int  `json:"prevCheckpoints"`
}

type ActiveRound struct {
	Players map[string]PlayerWaypoint `json:"players"`
}

func (s *Server) ToServerResponse() ServerResponse {
	isConnected := false
	if s.Client != nil {
		isConnected = s.Client.IsConnected
	}

	return ServerResponse{
		Id:          s.Id,
		Name:        s.Name,
		Description: s.Description,
		Host:        s.Host,
		XMLRPCPort:  s.XMLRPCPort,
		User:        s.User,
		Pass:        s.Pass,
		FMUrl:       s.FMUrl,
		IsConnected: isConnected,
	}
}

func (servers ServerList) ToServerResponses() []ServerResponse {
	responses := make([]ServerResponse, len(servers))
	for i, s := range servers {
		isConnected := false
		if s.Client != nil {
			isConnected = s.Client.IsConnected
		}

		responses[i] = ServerResponse{
			Id:          s.Id,
			Name:        s.Name,
			Description: s.Description,
			Host:        s.Host,
			XMLRPCPort:  s.XMLRPCPort,
			User:        s.User,
			Pass:        s.Pass,
			FMUrl:       s.FMUrl,
			IsConnected: isConnected,
		}
	}
	return responses
}

func (s *Server) ResetLiveInfo() {
	liveInfo := &LiveInfo{
		Teams:   make(map[int]Team),
		Players: make(map[string]PlayerRound),
		ActiveRound: ActiveRound{
			Players: make(map[string]PlayerWaypoint),
		},
	}

	if s.Info == nil {
		s.Info = &ServerInfo{}
	}
	s.Info.LiveInfo = liveInfo
}
