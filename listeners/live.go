package listeners

import (
	"time"

	"github.com/MRegterschot/GbxConnector/handlers"
	"github.com/MRegterschot/GbxConnector/lib"
	"github.com/MRegterschot/GbxConnector/structs"
	"github.com/MRegterschot/GbxRemoteGo/events"
	"github.com/MRegterschot/GbxRemoteGo/gbxclient"
	"go.uber.org/zap"
)

type LiveListener struct {
	Server *structs.Server
}

func AddLiveListeners(server *structs.Server) *LiveListener {
	ll := &LiveListener{Server: server}

	server.Client.OnPlayerFinish = append(server.Client.OnPlayerFinish, gbxclient.GbxCallbackStruct[events.PlayerWayPointEventArgs]{
		Key:  "gbxconnector",
		Call: ll.onPlayerFinish,
	})

	server.Client.OnPlayerCheckpoint = append(server.Client.OnPlayerCheckpoint, gbxclient.GbxCallbackStruct[events.PlayerWayPointEventArgs]{
		Key:  "gbxconnector",
		Call: ll.onPlayerCheckpoint,
	})

	server.Client.OnStartRound = append(server.Client.OnStartRound, gbxclient.GbxCallbackStruct[struct{}]{
		Key:  "gbxconnector",
		Call: ll.onStartRound,
	})

	server.Client.OnEndRound = append(server.Client.OnEndRound, gbxclient.GbxCallbackStruct[events.ScoresEventArgs]{
		Key:  "gbxconnector",
		Call: ll.onEndRound,
	})

	server.Client.OnBeginMap = append(server.Client.OnBeginMap, gbxclient.GbxCallbackStruct[events.MapEventArgs]{
		Key:  "gbxconnector",
		Call: ll.onBeginMap,
	})

	server.Client.OnEndMap = append(server.Client.OnEndMap, gbxclient.GbxCallbackStruct[events.MapEventArgs]{
		Key:  "gbxconnector",
		Call: ll.onEndMap,
	})

	server.Client.OnBeginMatch = append(server.Client.OnBeginMatch, gbxclient.GbxCallbackStruct[struct{}]{
		Key:  "gbxconnector",
		Call: ll.onBeginMatch,
	})

	server.Client.OnPlayerGiveUp = append(server.Client.OnPlayerGiveUp, gbxclient.GbxCallbackStruct[events.PlayerGiveUpEventArgs]{
		Key:  "gbxconnector",
		Call: ll.onPlayerGiveUp,
	})

	server.Client.OnWarmUpStart = append(server.Client.OnWarmUpStart, gbxclient.GbxCallbackStruct[struct{}]{
		Key:  "gbxconnector",
		Call: ll.onWarmUpStart,
	})

	server.Client.OnWarmUpEnd = append(server.Client.OnWarmUpEnd, gbxclient.GbxCallbackStruct[struct{}]{
		Key:  "gbxconnector",
		Call: ll.onWarmUpEnd,
	})

	server.Client.OnWarmUpStartRound = append(server.Client.OnWarmUpStartRound, gbxclient.GbxCallbackStruct[events.WarmUpEventArgs]{
		Key:  "gbxconnector",
		Call: ll.onWarmUpStartRound,
	})

	server.Client.OnPlayerInfoChanged = append(server.Client.OnPlayerInfoChanged, gbxclient.GbxCallbackStruct[events.PlayerInfoChangedEventArgs]{
		Key:  "gbxconnector",
		Call: ll.onPlayerInfoChanged,
	})

	server.Client.OnPlayerConnect = append(server.Client.OnPlayerConnect, gbxclient.GbxCallbackStruct[events.PlayerConnectEventArgs]{
		Key:  "gbxconnector",
		Call: ll.onPlayerConnect,
	})

	server.Client.OnPlayerDisconnect = append(server.Client.OnPlayerDisconnect, gbxclient.GbxCallbackStruct[events.PlayerDisconnectEventArgs]{
		Key:  "gbxconnector",
		Call: ll.onPlayerDisconnect,
	})

	return ll
}

func (ll *LiveListener) onPlayerFinish(playerFinishEvent events.PlayerWayPointEventArgs) {
	pw := ll.Server.Info.LiveInfo.ActiveRound.Players[playerFinishEvent.Login]

	pw.Time = playerFinishEvent.RaceTime
	pw.HasFinished = true
	pw.Checkpoint = playerFinishEvent.CheckpointInRace + 1

	ll.Server.Info.LiveInfo.ActiveRound.Players[playerFinishEvent.Login] = pw

	handlers.BroadcastLive(ll.Server.Id, map[string]structs.ActiveRound{
		"finish": ll.Server.Info.LiveInfo.ActiveRound,
	})
}

func (ll *LiveListener) onPlayerCheckpoint(playerCheckpointEvent events.PlayerWayPointEventArgs) {
	pw := ll.Server.Info.LiveInfo.ActiveRound.Players[playerCheckpointEvent.Login]

	pw.Time = playerCheckpointEvent.RaceTime
	pw.Checkpoint = playerCheckpointEvent.CheckpointInRace + 1

	ll.Server.Info.LiveInfo.ActiveRound.Players[playerCheckpointEvent.Login] = pw

	handlers.BroadcastLive(ll.Server.Id, map[string]structs.ActiveRound{
		"checkpoint": ll.Server.Info.LiveInfo.ActiveRound,
	})
}

func (ll *LiveListener) onStartRound(_ struct{}) {
	playerList, err := ll.Server.Client.GetPlayerList(1000, 0)
	if err != nil {
		zap.L().Error("Failed to get player list", zap.Int("server_id", ll.Server.Id), zap.Error(err))
	}

	ll.Server.Info.LiveInfo.ActiveRound = structs.ActiveRound{
		Players: make(map[string]structs.PlayerWaypoint),
	}

	for _, player := range playerList {
		if player.SpectatorStatus == 0 {
			playerInfo := ll.Server.Info.LiveInfo.Players[player.Login]

			playerWaypoint := structs.PlayerWaypoint{
				Login:       player.Login,
				AccountId:   playerInfo.AccountId,
				Time:        0,
				HasFinished: false,
				IsFinalist:  playerInfo.MatchPoints == *ll.Server.Info.LiveInfo.PointsLimit,
				Checkpoint:  0,
			}

			ll.Server.Info.LiveInfo.ActiveRound.Players[player.Login] = playerWaypoint
		}
	}

	handlers.BroadcastLive(ll.Server.Id, map[string]structs.ActiveRound{
		"beginRound": ll.Server.Info.LiveInfo.ActiveRound,
	})
}

func (ll *LiveListener) onEndRound(endRoundEvent events.ScoresEventArgs) {
	if endRoundEvent.UseTeams {
		for _, team := range endRoundEvent.Teams {
			t := ll.Server.Info.LiveInfo.Teams[team.ID]
			t.RoundPoints = team.RoundPoints
			t.MatchPoints = team.MatchPoints
			ll.Server.Info.LiveInfo.Teams[team.ID] = t
		}
	}

	for _, player := range endRoundEvent.Players {
		p := ll.Server.Info.LiveInfo.Players[player.Login]
		p.RoundPoints = player.RoundPoints
		p.MatchPoints = player.MatchPoints
		p.Finalist = player.MatchPoints == *ll.Server.Info.LiveInfo.PointsLimit
		p.Winner = player.MatchPoints > *ll.Server.Info.LiveInfo.PointsLimit
		p.BestTime = player.BestRaceTime
		p.BestCheckpoints = player.BestRaceCheckpoints
		p.PrevTime = player.PrevRaceTime
		p.PrevCheckpoints = player.PrevRaceCheckpoints
		ll.Server.Info.LiveInfo.Players[player.Login] = p
	}

	handlers.BroadcastLive(ll.Server.Id, map[string]*structs.LiveInfo{
		"endRound": ll.Server.Info.LiveInfo,
	})
}

func (ll *LiveListener) onBeginMap(beginMapEvent events.MapEventArgs) {
	ll.Server.Info.LiveInfo.CurrentMap = beginMapEvent.Map.Uid

	handlers.BroadcastLive(ll.Server.Id, map[string]string{
		"beginMap": beginMapEvent.Map.Uid,
	})
}

func (ll *LiveListener) onEndMap(endMapEvent events.MapEventArgs) {
	handlers.BroadcastLive(ll.Server.Id, map[string]string{
		"endMap": endMapEvent.Map.Uid,
	})
}

func (ll *LiveListener) onBeginMatch(_ struct{}) {
	ll.SyncLiveInfo()

	time.Sleep(300 * time.Millisecond) // Wait a bit for callbacks to be set

	handlers.BroadcastLive(ll.Server.Id, map[string]*structs.LiveInfo{
		"beginMatch": ll.Server.Info.LiveInfo,
	})
}

func (ll *LiveListener) onPlayerGiveUp(playerGiveUpEvent events.PlayerGiveUpEventArgs) {
	r := ll.Server.Info.LiveInfo.ActiveRound.Players[playerGiveUpEvent.Login]
	r.HasGivenUp = true
	ll.Server.Info.LiveInfo.ActiveRound.Players[playerGiveUpEvent.Login] = r

	handlers.BroadcastLive(ll.Server.Id, map[string]structs.ActiveRound{
		"giveUp": ll.Server.Info.LiveInfo.ActiveRound,
	})
}

func (ll *LiveListener) onWarmUpStart(_ struct{}) {
	ll.Server.Info.LiveInfo.IsWarmUp = true

	handlers.BroadcastLive(ll.Server.Id, map[string]*structs.LiveInfo{
		"warmUpStart": ll.Server.Info.LiveInfo,
	})
}

func (ll *LiveListener) onWarmUpEnd(_ struct{}) {
	ll.Server.Info.LiveInfo.IsWarmUp = false
	handlers.BroadcastLive(ll.Server.Id, map[string]*structs.LiveInfo{
		"warmUpEnd": ll.Server.Info.LiveInfo,
	})
}

func (ll *LiveListener) onWarmUpStartRound(warmUpEvent events.WarmUpEventArgs) {
	ll.Server.Info.LiveInfo.IsWarmUp = true
	ll.Server.Info.LiveInfo.WarmUpRound = &warmUpEvent.Current
	ll.Server.Info.LiveInfo.WarmUpTotalRounds = &warmUpEvent.Total

	playerList, err := ll.Server.Client.GetPlayerList(1000, 0)
	if err != nil {
		zap.L().Error("Failed to get player list", zap.Int("server_id", ll.Server.Id), zap.Error(err))
	}

	ll.Server.Info.LiveInfo.ActiveRound = structs.ActiveRound{
		Players: make(map[string]structs.PlayerWaypoint),
	}

	for _, player := range playerList {
		if player.SpectatorStatus == 0 {
			playerWaypoint := structs.PlayerWaypoint{
				Login:       player.Login,
				AccountId:   "",
				Time:        0,
				HasFinished: false,
				Checkpoint:  0,
			}

			ll.Server.Info.LiveInfo.ActiveRound.Players[player.Login] = playerWaypoint
		}
	}

	handlers.BroadcastLive(ll.Server.Id, map[string]*structs.LiveInfo{
		"warmUpStartRound": ll.Server.Info.LiveInfo,
	})
}

func (ll *LiveListener) onPlayerInfoChanged(playerInfoChangedEvent events.PlayerInfoChangedEventArgs) {
	p := ll.Server.Info.LiveInfo.Players[playerInfoChangedEvent.PlayerInfo.Login]
	p.Team = playerInfoChangedEvent.PlayerInfo.TeamId
	p.Name = playerInfoChangedEvent.PlayerInfo.NickName
	ll.Server.Info.LiveInfo.Players[playerInfoChangedEvent.PlayerInfo.Login] = p

	spectator := playerInfoChangedEvent.PlayerInfo.SpectatorStatus != 0
	if spectator {
		delete(ll.Server.Info.LiveInfo.ActiveRound.Players, playerInfoChangedEvent.PlayerInfo.Login)
	} else {
		playerWaypoint := structs.PlayerWaypoint{
			Login: playerInfoChangedEvent.PlayerInfo.Login,
		}

		ll.Server.Info.LiveInfo.ActiveRound.Players[playerInfoChangedEvent.PlayerInfo.Login] = playerWaypoint
	}

	handlers.BroadcastLive(ll.Server.Id, map[string]structs.ActiveRound{
		"playerInfoChanged": ll.Server.Info.LiveInfo.ActiveRound,
	})
}

func (ll *LiveListener) onPlayerConnect(playerConnectEvent events.PlayerConnectEventArgs) {
	playerInfo, err := ll.Server.Client.GetPlayerInfo(playerConnectEvent.Login)
	if err != nil {
		zap.L().Error("Failed to get player info", zap.Int("server_id", ll.Server.Id), zap.Error(err))
		return
	}

	if _, ok := ll.Server.Info.LiveInfo.Players[playerConnectEvent.Login]; !ok {
		ll.Server.Info.LiveInfo.Players[playerConnectEvent.Login] = structs.PlayerRound{
			Login: playerConnectEvent.Login,
			Name:  playerInfo.NickName,
			Team:  playerInfo.TeamId,
		}
	} else {
		p := ll.Server.Info.LiveInfo.Players[playerConnectEvent.Login]
		p.Team = playerInfo.TeamId
		ll.Server.Info.LiveInfo.Players[playerConnectEvent.Login] = p
	}

	if playerInfo.SpectatorStatus == 0 {
		playerWaypoint := structs.PlayerWaypoint{
			Login: playerConnectEvent.Login,
		}

		ll.Server.Info.LiveInfo.ActiveRound.Players[playerConnectEvent.Login] = playerWaypoint
	} else {
		delete(ll.Server.Info.LiveInfo.ActiveRound.Players, playerConnectEvent.Login)
	}

	handlers.BroadcastLive(ll.Server.Id, map[string]*structs.LiveInfo{
		"playerConnect": ll.Server.Info.LiveInfo,
	})
}

func (ll *LiveListener) onPlayerDisconnect(playerDisconnectEvent events.PlayerDisconnectEventArgs) {
	delete(ll.Server.Info.LiveInfo.ActiveRound.Players, playerDisconnectEvent.Login)

	handlers.BroadcastLive(ll.Server.Id, map[string]structs.ActiveRound{
		"playerDisconnect": ll.Server.Info.LiveInfo.ActiveRound,
	})
}

func (ll *LiveListener) SyncLiveInfo() {
	// Set warmup status
	ll.Server.Client.AddScriptCallback("Trackmania.WarmUp.Status", "server", func(event any) {
		onWarmUpStatus(event, ll.Server)
	})
	ll.Server.Client.TriggerModeScriptEventArray("Trackmania.WarmUp.GetStatus", []string{"gbxconnector"})

	// Set the current game mode
	mode, err := ll.Server.Client.GetScriptName()
	if err != nil {
		zap.L().Error("Failed to get script name", zap.Int("server_id", ll.Server.Id), zap.Error(err))
	}

	ll.Server.Info.LiveInfo.Mode = mode.CurrentValue

	// Set the use teams status
	ll.Server.Client.AddScriptCallback("Maniaplanet.Mode.UseTeams", "server", func(event any) {
		onUseTeams(event, ll.Server)
	})
	ll.Server.Client.TriggerModeScriptEventArray("Maniaplanet.Mode.GetUseTeams", []string{"gbxconnector"})

	mapInfo, err := ll.Server.Client.GetCurrentMapInfo()
	if err != nil {
		zap.L().Error("Failed to get current map info", zap.Int("server_id", ll.Server.Id), zap.Error(err))
	}

	ll.Server.Info.LiveInfo.CurrentMap = mapInfo.UId

	// Get script settings
	scriptSettings, err := ll.Server.Client.GetModeScriptSettings()
	if err != nil {
		zap.L().Error("Failed to get script settings", zap.Int("server_id", ll.Server.Id), zap.Error(err))
	}

	// Set points limit
	pointsLimit, ok := scriptSettings["S_PointsLimit"].(int)
	if !ok {
		zap.L().Debug("PointsLimit not found in script settings", zap.Int("server_id", ll.Server.Id))
	} else if pointsLimit > 0 {
		ll.Server.Info.LiveInfo.PointsLimit = &pointsLimit
	}

	// Set rounds limit
	roundsLimit, ok := scriptSettings["S_RoundsPerMap"].(int)
	if !ok {
		zap.L().Debug("RoundsLimit not found in script settings", zap.Int("server_id", ll.Server.Id))
	} else if roundsLimit > 0 {
		ll.Server.Info.LiveInfo.RoundsLimit = &roundsLimit
	}

	// Set map limit
	mapLimit, ok := scriptSettings["S_MapsPerMatch"].(int)
	if !ok {
		zap.L().Debug("MapLimit not found in script settings", zap.Int("server_id", ll.Server.Id))
	} else if mapLimit > 0 {
		ll.Server.Info.LiveInfo.MapLimit = &mapLimit
	}

	// Set map list
	mapList, err := ll.Server.Client.GetMapList(1000, 0)
	if err != nil {
		zap.L().Error("Failed to get map list", zap.Int("server_id", ll.Server.Id), zap.Error(err))
	}

	ll.Server.Info.LiveInfo.Maps = make([]string, len(mapList))
	for i, m := range mapList {
		ll.Server.Info.LiveInfo.Maps[i] = m.UId
	}

	ll.Server.Client.AddScriptCallback("Trackmania.Scores", "server", func(event any) {
		onScores(event, ll.Server)
	})
	ll.Server.Client.TriggerModeScriptEventArray("Trackmania.GetScores", []string{"gbxconnector"})
}

func onWarmUpStatus(event any, server *structs.Server) {
	var status structs.WarmUpStatus
	if err := lib.ConvertCallbackData(event, &status); err != nil {
		zap.L().Error("Failed to get callback data", zap.Error(err))
		return
	}

	if status.ResponseId != "gbxconnector" {
		return
	}

	server.Info.LiveInfo.IsWarmUp = status.Active
}

func onScores(event any, server *structs.Server) {
	var scores structs.Scores
	if err := lib.ConvertCallbackData(event, &scores); err != nil {
		zap.L().Error("Failed to get callback data", zap.Error(err))
		return
	}

	if scores.ResponseId != "gbxconnector" {
		return
	}

	if scores.UseTeams {
		server.Info.LiveInfo.Teams = make(map[int]structs.Team)
		for _, team := range scores.Teams {
			server.Info.LiveInfo.Teams[team.Id] = structs.Team{
				Id:          team.Id,
				Name:        team.Name,
				RoundPoints: team.RoundPoints,
				MatchPoints: team.MatchPoints,
			}
		}
	}

	server.Info.LiveInfo.Players = make(map[string]structs.PlayerRound)
	for _, player := range scores.Players {
		server.Info.LiveInfo.Players[player.Login] = structs.PlayerRound{
			Login:           player.Login,
			AccountId:       player.AccountId,
			Name:            player.Name,
			Team:            player.Team,
			Rank:            player.Rank,
			Finalist:        player.MatchPoints == *server.Info.LiveInfo.PointsLimit,
			Winner:          player.MatchPoints > *server.Info.LiveInfo.PointsLimit,
			RoundPoints:     player.RoundPoints,
			MatchPoints:     player.MatchPoints,
			BestTime:        player.BestRaceTime,
			BestCheckpoints: player.BestCheckpoints,
			PrevTime:        player.PrevRaceTime,
			PrevCheckpoints: player.PrevCheckpoints,
		}
	}

	playerList, err := server.Client.GetPlayerList(1000, 0)
	if err != nil {
		zap.L().Error("Failed to get player list", zap.Int("server_id", server.Id), zap.Error(err))
	}

	server.Info.LiveInfo.ActiveRound = structs.ActiveRound{
		Players: make(map[string]structs.PlayerWaypoint),
	}

	for _, player := range playerList {
		if player.SpectatorStatus == 0 {
			playerInfo := server.Info.LiveInfo.Players[player.Login]

			playerWaypoint := structs.PlayerWaypoint{
				Login:       player.Login,
				AccountId:   playerInfo.AccountId,
				Time:        0,
				HasFinished: false,
				IsFinalist:  playerInfo.Finalist,
				Checkpoint:  0,
			}

			server.Info.LiveInfo.ActiveRound.Players[player.Login] = playerWaypoint
		}
	}
}

func onUseTeams(event any, server *structs.Server) {
	var useTeams structs.UseTeams
	if err := lib.ConvertCallbackData(event, &useTeams); err != nil {
		zap.L().Error("Failed to get callback data", zap.Error(err))
		return
	}

	if useTeams.ResponseId != "gbxconnector" {
		return
	}

	server.Info.LiveInfo.UseTeams = useTeams.Teams
}
