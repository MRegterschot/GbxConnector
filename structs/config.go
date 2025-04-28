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
}

type Env struct {
	Port       int
	CorsOrigin string
	Servers    []Server `json:"servers"`
}
