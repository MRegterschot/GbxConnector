package structs

import (
	"context"
	"time"

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

	// Internal
	IsLocal    bool                 `json:"-"`
	ActiveMap  string               `json:"-"`
	Client     *gbxclient.GbxClient `json:"-"`
	CancelFunc context.CancelFunc   `json:"-"`
	Ctx        context.Context      `json:"-"`
}

type ServerList []*Server

type Env struct {
	Port              int
	CorsOrigins       []string
	LogLevel          string
	ReconnectInterval time.Duration
	Servers           ServerList `json:"servers"`
}

type ServerResponse struct {
	Id          int     `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Host        string  `json:"host"`
	XMLRPCPort  int     `json:"xmlrpcPort"`
	User        string  `json:"user"`
	Pass        string  `json:"pass"`
	IsLocal     bool    `json:"isLocal"`
	IsConnected bool    `json:"isConnected"`
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
			IsLocal:     s.IsLocal,
			IsConnected: isConnected,
		}
	}
	return responses
}
