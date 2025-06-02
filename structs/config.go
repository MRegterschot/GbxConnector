package structs

import (
	"time"
)

type Env struct {
	Port              int
	LogLevel          string
	ReconnectInterval time.Duration
	JwtSecret         string
	InternalApiKey    string
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
	FMUrl       *string `json:"fmUrl,omitempty"`
	IsConnected bool    `json:"isConnected"`
}
