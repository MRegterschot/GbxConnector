package structs

import (
	"context"

	"github.com/MRegterschot/GbxRemoteGo/gbxclient"
)

type Server struct {
	Uuid        string  `json:"uuid"`
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

type ServerResponse struct {
	Uuid        string  `json:"uuid"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Host        string  `json:"host"`
	XMLRPCPort  int     `json:"xmlrpcPort"`
	User        string  `json:"user"`
	Pass        string  `json:"pass"`
	FMUrl       *string `json:"fmUrl,omitempty"`
	IsConnected bool    `json:"isConnected"`
}

type ServerList []*Server

type ServerInfo struct {
	ActiveMap     string       `json:"-"`
	ActivePlayers []PlayerInfo `json:"-"`
	LiveInfo      *LiveInfo    `json:"-"`
	Chat          ChatConfig   `json:"-"`
}

func (s *Server) ToServerResponse() ServerResponse {
	isConnected := false
	if s.Client != nil {
		isConnected = s.Client.IsConnected
	}

	return ServerResponse{
		Uuid:        s.Uuid,
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
			Uuid:        s.Uuid,
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
		PointsRepartition: []int{},
		Teams:             make(map[int]Team),
		Players:           make(map[string]PlayerRound),
		ActiveRound: ActiveRound{
			Players: make(map[string]PlayerWaypoint),
		},
	}

	if s.Info == nil {
		s.Info = &ServerInfo{}
	}
	s.Info.LiveInfo = liveInfo
}

func (s *Server) UpdateServer(name string, description *string, host string, xmlrpcPort int, user, pass string, fmUrl *string) {
	s.Name = name
	s.Description = description
	s.Host = host
	s.XMLRPCPort = xmlrpcPort
	s.User = user
	s.Pass = pass
	s.FMUrl = fmUrl

	s.ResetLiveInfo()
}
