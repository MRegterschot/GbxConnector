package structs

import (
	"time"
)

type Env struct {
	Port              int
	LogLevel          string
	ReconnectInterval time.Duration
	JwtSecret         string
	Servers           ServerList `json:"servers"`
}
