package structs

type User struct {
	ID          string `json:"_id"`
	AccountID   string `json:"accountId"`
	Login       string `json:"login"`
	DisplayName string `json:"displayName"`
	Admin       bool   `json:"admin"`
	UbiId       string `json:"ubiId"`
}
