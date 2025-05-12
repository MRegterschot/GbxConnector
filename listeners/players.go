package listeners

import (
	"fmt"

	"github.com/MRegterschot/GbxConnector/structs"
	"github.com/MRegterschot/GbxRemoteGo/events"
	"github.com/MRegterschot/GbxRemoteGo/gbxclient"
	gbxStructs "github.com/MRegterschot/GbxRemoteGo/structs"
	"go.uber.org/zap"
	"slices"
)

type PlayersListener struct {
	Server *structs.Server
}

func AddPlayersListeners(server *structs.Server) {
	pl := &PlayersListener{Server: server}
	server.Client.OnPlayerConnect = append(server.Client.OnPlayerConnect, gbxclient.GbxCallbackStruct[events.PlayerConnectEventArgs]{
		Key:  "PlayerConnectListener",
		Call: pl.onPlayerConnect,
	})
	server.Client.OnPlayerDisconnect = append(server.Client.OnPlayerDisconnect, gbxclient.GbxCallbackStruct[events.PlayerDisconnectEventArgs]{
		Key:  "PlayerDisconnectListener",
		Call: pl.onPlayerDisconnect,
	})
	server.Client.OnPlayerInfoChanged = append(server.Client.OnPlayerInfoChanged, gbxclient.GbxCallbackStruct[events.PlayerInfoChangedEventArgs]{
		Key:  "PlayerInfoChangedListener",
		Call: pl.onPlayerInfoChanged,
	})
}

func (pl *PlayersListener) onPlayerConnect(playerConnectEvent events.PlayerConnectEventArgs) {
	playerInfo, err := pl.Server.Client.GetPlayerInfo(playerConnectEvent.Login)
	if err != nil {
		zap.L().Error("Failed to get player info", zap.Error(err))
		return
	}

	pl.Server.ActivePlayers = append(pl.Server.ActivePlayers, playerInfo)
}

func (pl *PlayersListener) onPlayerDisconnect(playerDisconnectEvent events.PlayerDisconnectEventArgs) {
	for i, player := range pl.Server.ActivePlayers {
		if player.Login == playerDisconnectEvent.Login {
			pl.Server.ActivePlayers = slices.Delete(pl.Server.ActivePlayers, i, i+1)
			break
		}
	}

	fmt.Println("Player disconnected:", playerDisconnectEvent.Login)
}

func (pl *PlayersListener) onPlayerInfoChanged(playerInfoChangedEvent events.PlayerInfoChangedEventArgs) {
	for i, player := range pl.Server.ActivePlayers {
		if player.Login == playerInfoChangedEvent.PlayerInfo.Login {
			pl.Server.ActivePlayers[i] = gbxStructs.TMPlayerInfo{
				Login:           playerInfoChangedEvent.PlayerInfo.Login,
				NickName:        playerInfoChangedEvent.PlayerInfo.NickName,
				PlayerId:        playerInfoChangedEvent.PlayerInfo.PlayerId,
				TeamId:          playerInfoChangedEvent.PlayerInfo.TeamId,
				SpectatorStatus: playerInfoChangedEvent.PlayerInfo.SpectatorStatus,
				LadderRanking:   playerInfoChangedEvent.PlayerInfo.LadderRanking,
				Flags:           playerInfoChangedEvent.PlayerInfo.Flags,
			}
			break
		}
	}

	fmt.Println("Player info changed:", playerInfoChangedEvent.PlayerInfo.Login)
}
