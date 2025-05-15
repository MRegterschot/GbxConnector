package listeners

import (
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

	return ll
}

func (ll *LiveListener) onPlayerFinish(playerFinishEvent events.PlayerWayPointEventArgs) {
	playerWaypoint := structs.PlayerWaypoint{
		Login:       playerFinishEvent.Login,
		AccountId:   playerFinishEvent.AccountId,
		Time:        playerFinishEvent.RaceTime,
		HasFinished: true,
		Checkpoint:  playerFinishEvent.CheckpointInRace,
	}

	ll.Server.Info.LiveInfo.ActiveRound.Players[playerFinishEvent.Login] = playerWaypoint

	handlers.BroadcastLive(ll.Server.Id, map[string]structs.ActiveRound{
		"finish": ll.Server.Info.LiveInfo.ActiveRound,
	})
}

func (ll *LiveListener) onPlayerCheckpoint(playerCheckpointEvent events.PlayerWayPointEventArgs) {
	playerWaypoint := structs.PlayerWaypoint{
		Login:       playerCheckpointEvent.Login,
		AccountId:   playerCheckpointEvent.AccountId,
		Time:        playerCheckpointEvent.RaceTime,
		HasFinished: false,
		Checkpoint:  playerCheckpointEvent.CheckpointInRace,
	}

	ll.Server.Info.LiveInfo.ActiveRound.Players[playerCheckpointEvent.Login] = playerWaypoint

	handlers.BroadcastLive(ll.Server.Id, map[string]structs.ActiveRound{
		"checkpoint": ll.Server.Info.LiveInfo.ActiveRound,
	})
}

func (ll *LiveListener) onStartRound(_ struct{}) {
	ll.Server.Info.LiveInfo.ActiveRound = structs.ActiveRound{
		Players: make(map[string]structs.PlayerWaypoint),
	}

	handlers.BroadcastLive(ll.Server.Id, map[string]structs.ActiveRound{
		"beginRound": ll.Server.Info.LiveInfo.ActiveRound,
	})
}

func (ll *LiveListener) onEndRound(endRoundEvent events.ScoresEventArgs) {
	for _, team := range endRoundEvent.Teams {
		t := ll.Server.Info.LiveInfo.Teams[team.ID]
		t.RoundPoints = team.RoundPoints
		t.MatchPoints = team.MatchPoints
		ll.Server.Info.LiveInfo.Teams[team.ID] = t
	}

	for _, player := range endRoundEvent.Players {
		p := ll.Server.Info.LiveInfo.Players[player.Login]
		p.RoundPoints = player.RoundPoints
		p.MatchPoints = player.MatchPoints
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

	ll.Server.Info.LiveInfo.ActiveRound = structs.ActiveRound{
		Players: make(map[string]structs.PlayerWaypoint),
	}

	handlers.BroadcastLive(ll.Server.Id, map[string]*structs.LiveInfo{
		"warmUpStartRound": ll.Server.Info.LiveInfo,
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
	} else {
		ll.Server.Info.LiveInfo.PointsLimit = &pointsLimit
	}

	// Set rounds limit
	roundsLimit, ok := scriptSettings["S_RoundsPerMap"].(int)
	if !ok {
		zap.L().Debug("RoundsLimit not found in script settings", zap.Int("server_id", ll.Server.Id))
	} else {
		ll.Server.Info.LiveInfo.RoundsLimit = &roundsLimit
	}

	// Set map limit
	mapLimit, ok := scriptSettings["S_MapsPerMatch"].(int)
	if !ok {
		zap.L().Debug("MapLimit not found in script settings", zap.Int("server_id", ll.Server.Id))
	} else {
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

	ll.Server.Info.LiveInfo.ActiveRound = structs.ActiveRound{
		Players: make(map[string]structs.PlayerWaypoint),
	}
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

	server.Info.LiveInfo.Teams = make(map[int]structs.Team)
	for _, team := range scores.Teams {
		server.Info.LiveInfo.Teams[team.Id] = structs.Team{
			Id:          team.Id,
			Name:        team.Name,
			RoundPoints: team.RoundPoints,
			MatchPoints: team.MatchPoints,
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
			RoundPoints:     player.RoundPoints,
			MatchPoints:     player.MatchPoints,
			BestTime:        player.BestRaceTime,
			BestCheckpoints: player.BestCheckpoints,
			PrevTime:        player.PrevRaceTime,
			PrevCheckpoints: player.PrevCheckpoints,
		}
	}
}
