package structs

type Server struct {
	Id          int     `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Host        string  `json:"host"`
	XMLRPCPort  int     `json:"xmlrpcPort"`
	User        string  `json:"user"`
	Pass        string  `json:"pass"`
	IsLocal     bool    `json:"isLocal"`
	IsConnected bool    `json:"isConnected"`
}

type ServerList []Server

type Env struct {
	Port        int
	CorsOrigins []string
	LogLevel    string
	Servers     ServerList `json:"servers"`
}

type ServerResponse struct {
	Id          int     `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Host        string  `json:"host"`
	XMLRPCPort  int     `json:"xmlrpcPort"`
	IsLocal     bool    `json:"isLocal"`
	IsConnected bool    `json:"isConnected"`
}

func (servers ServerList) ToServerResponses() []ServerResponse {
	responses := make([]ServerResponse, len(servers))
	for i, s := range servers {
		responses[i] = ServerResponse{
			Id:          s.Id,
			Name:        s.Name,
			Description: s.Description,
			Host:        s.Host,
			XMLRPCPort:  s.XMLRPCPort,
			IsLocal:     s.IsLocal,
			IsConnected: s.IsConnected,
		}
	}
	return responses
}
