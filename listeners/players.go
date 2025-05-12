package listeners

import (
	"slices"

	"github.com/MRegterschot/GbxConnector/handlers"
	"github.com/MRegterschot/GbxConnector/structs"
	"github.com/MRegterschot/GbxRemoteGo/events"
	"github.com/MRegterschot/GbxRemoteGo/gbxclient"
	gbxstructs "github.com/MRegterschot/GbxRemoteGo/structs"
	"go.uber.org/zap"
)

type PlayersListener struct {
	Server *structs.Server
}

func AddPlayersListeners(server *structs.Server) *PlayersListener {
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

	return pl
}

func (pl *PlayersListener) onPlayerConnect(playerConnectEvent events.PlayerConnectEventArgs) {
	playerInfo, err := pl.Server.Client.GetPlayerInfo(playerConnectEvent.Login)
	if err != nil {
		zap.L().Error("Failed to get player info", zap.Error(err))
		return
	}

	pl.Server.ActivePlayers = append(pl.Server.ActivePlayers, playerInfo)

	handlers.BroadcastPlayers(pl.Server.Id, map[string]gbxstructs.TMPlayerInfo{
		"connect": playerInfo,
	})
}

func (pl *PlayersListener) onPlayerDisconnect(playerDisconnectEvent events.PlayerDisconnectEventArgs) {
	for i, player := range pl.Server.ActivePlayers {
		if player.Login == playerDisconnectEvent.Login {
			pl.Server.ActivePlayers = slices.Delete(pl.Server.ActivePlayers, i, i+1)
			break
		}
	}

	handlers.BroadcastPlayers(pl.Server.Id, map[string]string{
		"disconnect": playerDisconnectEvent.Login,
	})
}

func (pl *PlayersListener) onPlayerInfoChanged(playerInfoChangedEvent events.PlayerInfoChangedEventArgs) {
	var playerInfo gbxstructs.TMPlayerInfo
	for i, player := range pl.Server.ActivePlayers {
		if player.Login == playerInfoChangedEvent.PlayerInfo.Login {
			playerInfo = gbxstructs.TMPlayerInfo{
				Login:           playerInfoChangedEvent.PlayerInfo.Login,
				NickName:        playerInfoChangedEvent.PlayerInfo.NickName,
				PlayerId:        playerInfoChangedEvent.PlayerInfo.PlayerId,
				TeamId:          playerInfoChangedEvent.PlayerInfo.TeamId,
				SpectatorStatus: playerInfoChangedEvent.PlayerInfo.SpectatorStatus,
				LadderRanking:   playerInfoChangedEvent.PlayerInfo.LadderRanking,
				Flags:           playerInfoChangedEvent.PlayerInfo.Flags,
			}
			pl.Server.ActivePlayers[i] = playerInfo
			break
		}
	}

	handlers.BroadcastPlayers(pl.Server.Id, map[string]gbxstructs.TMPlayerInfo{
		"infoChanged": playerInfo,
	})
}

func (pl *PlayersListener) SyncPlayerList() {
	players, err := pl.Server.Client.GetPlayerList(1000, 0)
	if err != nil {
		zap.L().Error("Failed to get player list", zap.Error(err))
		return
	}

	pl.Server.ActivePlayers = players

	handlers.BroadcastPlayers(pl.Server.Id, map[string][]gbxstructs.TMPlayerInfo{
		"playerList": players,
	})
}
