package structs

import "github.com/MRegterschot/GbxRemoteGo/structs"

type PlayerInfo struct {
	Login           string `json:"login"`
	NickName        string `json:"nickName"`
	PlayerId        int    `json:"playerId"`
	TeamId          int    `json:"teamId"`
	SpectatorStatus int    `json:"spectatorStatus"`
}

func ToPlayerInfo(playerInfo structs.TMPlayerInfo) PlayerInfo {
	return PlayerInfo{
		Login:           playerInfo.Login,
		NickName:        playerInfo.NickName,
		PlayerId:        playerInfo.PlayerId,
		TeamId:          playerInfo.TeamId,
		SpectatorStatus: playerInfo.SpectatorStatus,
	}
}
