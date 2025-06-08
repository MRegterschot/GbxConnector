package listeners

import (
	"strconv"
	"strings"
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

	server.Client.OnEcho = append(server.Client.OnEcho, gbxclient.GbxCallbackStruct[events.EchoEventArgs]{
		Key:  "gbxconnector",
		Call: ll.onEcho,
	})

	server.Client.OnElimination = append(server.Client.OnElimination, gbxclient.GbxCallbackStruct[events.EliminationEventArgs]{
		Key:  "gbxconnector",
		Call: ll.onElimination,
	})

	return ll
}

func (ll *LiveListener) onPlayerFinish(playerFinishEvent events.PlayerWayPointEventArgs) {
	pw := ll.Server.Info.LiveInfo.ActiveRound.Players[playerFinishEvent.Login]

	pw.Time = playerFinishEvent.RaceTime
	pw.HasFinished = true
	pw.Checkpoint = playerFinishEvent.CheckpointInRace + 1

	ll.Server.Info.LiveInfo.ActiveRound.Players[playerFinishEvent.Login] = pw

	handlers.BroadcastLive(ll.Server.Uuid, map[string]structs.ActiveRound{
		"finish": ll.Server.Info.LiveInfo.ActiveRound,
	})

	if ll.Server.Info.LiveInfo.Type != "timeattack" {
		return
	}

	p := ll.Server.Info.LiveInfo.Players[playerFinishEvent.Login]
	if p.BestTime > 0 && p.BestTime <= playerFinishEvent.RaceTime {
		return
	}

	p.BestTime = playerFinishEvent.RaceTime
	ll.Server.Info.LiveInfo.Players[playerFinishEvent.Login] = p

	handlers.BroadcastLive(ll.Server.Uuid, map[string]*structs.LiveInfo{
		"personalBest": ll.Server.Info.LiveInfo,
	})
}

func (ll *LiveListener) onPlayerCheckpoint(playerCheckpointEvent events.PlayerWayPointEventArgs) {
	pw := ll.Server.Info.LiveInfo.ActiveRound.Players[playerCheckpointEvent.Login]

	pw.Time = playerCheckpointEvent.RaceTime
	pw.Checkpoint = playerCheckpointEvent.CheckpointInRace + 1
	pw.HasFinished = false
	pw.HasGivenUp = false

	ll.Server.Info.LiveInfo.ActiveRound.Players[playerCheckpointEvent.Login] = pw

	handlers.BroadcastLive(ll.Server.Uuid, map[string]structs.ActiveRound{
		"checkpoint": ll.Server.Info.LiveInfo.ActiveRound,
	})
}

func (ll *LiveListener) onStartRound(_ struct{}) {
	playerList, err := ll.Server.Client.GetPlayerList(1000, 0)
	if err != nil {
		zap.L().Error("Failed to get player list", zap.String("server_uuid", ll.Server.Uuid), zap.Error(err))
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
				IsFinalist:  isFinalist(playerInfo.MatchPoints, ll.Server.Info.LiveInfo.PointsLimit),
				Checkpoint:  0,
			}

			ll.Server.Info.LiveInfo.ActiveRound.Players[player.Login] = playerWaypoint
		}
	}

	handlers.BroadcastLive(ll.Server.Uuid, map[string]structs.ActiveRound{
		"beginRound": ll.Server.Info.LiveInfo.ActiveRound,
	})
}

func (ll *LiveListener) onEndRound(endRoundEvent events.ScoresEventArgs) {
	if endRoundEvent.UseTeams {
		for _, team := range endRoundEvent.Teams {
			t := ll.Server.Info.LiveInfo.Teams[team.ID]
			if ll.Server.Info.LiveInfo.Type == "tmwc" || ll.Server.Info.LiveInfo.Type == "tmwt" {
				t.RoundPoints = team.MapPoints
			} else {
				t.RoundPoints = team.RoundPoints
			}
			t.MatchPoints = team.MatchPoints
			ll.Server.Info.LiveInfo.Teams[team.ID] = t
		}
	}

	for _, player := range endRoundEvent.Players {
		p := ll.Server.Info.LiveInfo.Players[player.Login]
		p.RoundPoints = player.RoundPoints
		p.MatchPoints = player.MatchPoints
		p.Finalist = isFinalist(player.MatchPoints, ll.Server.Info.LiveInfo.PointsLimit)
		p.Winner = isWinner(player.MatchPoints, ll.Server.Info.LiveInfo.PointsLimit)
		p.BestTime = player.BestRaceTime
		p.BestCheckpoints = player.BestRaceCheckpoints
		p.PrevTime = player.PrevRaceTime
		p.PrevCheckpoints = player.PrevRaceCheckpoints
		ll.Server.Info.LiveInfo.Players[player.Login] = p
	}

	ll.Server.Client.TriggerModeScriptEventArray("Maniaplanet.Pause.GetStatus", []string{"gbxconnector"})

	time.Sleep(300 * time.Millisecond) // Wait a bit for callbacks to be set

	handlers.BroadcastLive(ll.Server.Uuid, map[string]*structs.LiveInfo{
		"endRound": ll.Server.Info.LiveInfo,
	})
}

func (ll *LiveListener) onBeginMap(beginMapEvent events.MapEventArgs) {
	ll.Server.Info.LiveInfo.CurrentMap = beginMapEvent.Map.Uid

	handlers.BroadcastLive(ll.Server.Uuid, map[string]string{
		"beginMap": beginMapEvent.Map.Uid,
	})
}

func (ll *LiveListener) onEndMap(endMapEvent events.MapEventArgs) {
	handlers.BroadcastLive(ll.Server.Uuid, map[string]string{
		"endMap": endMapEvent.Map.Uid,
	})
}

func (ll *LiveListener) onBeginMatch(_ struct{}) {
	SyncLiveInfo(ll.Server)

	time.Sleep(300 * time.Millisecond) // Wait a bit for callbacks to be set

	handlers.BroadcastLive(ll.Server.Uuid, map[string]*structs.LiveInfo{
		"beginMatch": ll.Server.Info.LiveInfo,
	})
}

func (ll *LiveListener) onPlayerGiveUp(playerGiveUpEvent events.PlayerGiveUpEventArgs) {
	r := ll.Server.Info.LiveInfo.ActiveRound.Players[playerGiveUpEvent.Login]
	r.HasGivenUp = true
	ll.Server.Info.LiveInfo.ActiveRound.Players[playerGiveUpEvent.Login] = r

	handlers.BroadcastLive(ll.Server.Uuid, map[string]structs.ActiveRound{
		"giveUp": ll.Server.Info.LiveInfo.ActiveRound,
	})
}

func (ll *LiveListener) onWarmUpStart(_ struct{}) {
	ll.Server.Info.LiveInfo.IsWarmUp = true

	handlers.BroadcastLive(ll.Server.Uuid, map[string]*structs.LiveInfo{
		"warmUpStart": ll.Server.Info.LiveInfo,
	})
}

func (ll *LiveListener) onWarmUpEnd(_ struct{}) {
	ll.Server.Info.LiveInfo.IsWarmUp = false
	handlers.BroadcastLive(ll.Server.Uuid, map[string]*structs.LiveInfo{
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

	handlers.BroadcastLive(ll.Server.Uuid, map[string]*structs.LiveInfo{
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

	handlers.BroadcastLive(ll.Server.Uuid, map[string]structs.ActiveRound{
		"playerInfoChanged": ll.Server.Info.LiveInfo.ActiveRound,
	})
}

func (ll *LiveListener) onPlayerConnect(playerConnectEvent events.PlayerConnectEventArgs) {
	playerInfo, err := ll.Server.Client.GetPlayerInfo(playerConnectEvent.Login)
	if err != nil {
		zap.L().Error("Failed to get player info", zap.String("server_uuid", ll.Server.Uuid), zap.Error(err))
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

	handlers.BroadcastLive(ll.Server.Uuid, map[string]*structs.LiveInfo{
		"playerConnect": ll.Server.Info.LiveInfo,
	})
}

func (ll *LiveListener) onPlayerDisconnect(playerDisconnectEvent events.PlayerDisconnectEventArgs) {
	delete(ll.Server.Info.LiveInfo.ActiveRound.Players, playerDisconnectEvent.Login)

	handlers.BroadcastLive(ll.Server.Uuid, map[string]structs.ActiveRound{
		"playerDisconnect": ll.Server.Info.LiveInfo.ActiveRound,
	})
}

func (ll *LiveListener) onEcho(echoEvent events.EchoEventArgs) {
	if echoEvent.Internal == "UpdatedSettings" {
		setScriptSettings(ll.Server)
		handlers.BroadcastLive(ll.Server.Uuid, map[string]*structs.LiveInfo{
			"updatedSettings": ll.Server.Info.LiveInfo,
		})
	}
}

func (ll *LiveListener) onElimination(eliminationEvent events.EliminationEventArgs) {
	for _, accountId := range eliminationEvent.AccountIds {
		for _, player := range ll.Server.Info.LiveInfo.Players {
			if player.AccountId == accountId {
				player.Eliminated = true
				ll.Server.Info.LiveInfo.Players[player.Login] = player
			}
		}
	}

	handlers.BroadcastLive(ll.Server.Uuid, map[string]*structs.LiveInfo{
		"elimination": ll.Server.Info.LiveInfo,
	})
}

func SyncLiveInfo(server *structs.Server) {
	// Set warmup status
	server.Client.AddScriptCallback("Trackmania.WarmUp.Status", "server", func(event any) {
		onWarmUpStatus(event, server)
	})
	server.Client.TriggerModeScriptEventArray("Trackmania.WarmUp.GetStatus", []string{"gbxconnector"})

	// Set the current game mode
	mode, err := server.Client.GetScriptName()
	if err != nil {
		zap.L().Error("Failed to get script name", zap.String("server_uuid", server.Uuid), zap.Error(err))
	}

	server.Info.LiveInfo.Mode = mode.CurrentValue

	modeLower := strings.ToLower(mode.CurrentValue)

	// Set the type
	switch {
	case strings.Contains(modeLower, "timeattack"):
		server.Info.LiveInfo.Type = "timeattack"
	case strings.Contains(modeLower, "rounds"):
		server.Info.LiveInfo.Type = "rounds"
	case strings.Contains(modeLower, "cup"):
		server.Info.LiveInfo.Type = "cup"
	case strings.Contains(modeLower, "tmwc"):
		server.Info.LiveInfo.Type = "tmwc"
	case strings.Contains(modeLower, "tmwt"):
		server.Info.LiveInfo.Type = "tmwt"
	case strings.Contains(modeLower, "teams"):
		server.Info.LiveInfo.Type = "teams"
	case strings.Contains(modeLower, "knockout"):
		server.Info.LiveInfo.Type = "knockout"
	default:
		server.Info.LiveInfo.Type = "rounds"
	}

	mapInfo, err := server.Client.GetCurrentMapInfo()
	if err != nil {
		zap.L().Error("Failed to get current map info", zap.String("server_uuid", server.Uuid), zap.Error(err))
	}

	server.Info.LiveInfo.CurrentMap = mapInfo.UId

	setScriptSettings(server)

	// Set map list
	mapList, err := server.Client.GetMapList(1000, 0)
	if err != nil {
		zap.L().Error("Failed to get map list", zap.String("server_uuid", server.Uuid), zap.Error(err))
	}

	server.Info.LiveInfo.Maps = make([]string, len(mapList))
	for i, m := range mapList {
		server.Info.LiveInfo.Maps[i] = m.UId
	}

	server.Client.AddScriptCallback("Trackmania.Scores", "server", func(event any) {
		onScores(event, server)
	})
	server.Client.TriggerModeScriptEventArray("Trackmania.GetScores", []string{"gbxconnector"})

	// Set pause status
	server.Client.AddScriptCallback("Maniaplanet.Pause.Status", "server", func(event any) {
		onPauseStatus(event, server)
	})
	server.Client.TriggerModeScriptEventArray("Maniaplanet.Pause.GetStatus", []string{"gbxconnector"})
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
			Finalist:        isFinalist(player.MatchPoints, server.Info.LiveInfo.PointsLimit),
			Winner:          isWinner(player.MatchPoints, server.Info.LiveInfo.PointsLimit),
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
		zap.L().Error("Failed to get player list", zap.String("server_uuid", server.Uuid), zap.Error(err))
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

func onPauseStatus(event any, server *structs.Server) {
	var status structs.Pause
	if err := lib.ConvertCallbackData(event, &status); err != nil {
		zap.L().Error("Failed to get callback data", zap.Error(err))
		return
	}

	if status.ResponseId != "gbxconnector" {
		return
	}

	server.Info.LiveInfo.PauseAvailable = status.Available
	server.Info.LiveInfo.IsPaused = status.Active
}

func setScriptSettings(server *structs.Server) {
	// Get script settings
	scriptSettings, err := server.Client.GetModeScriptSettings()
	if err != nil {
		zap.L().Error("Failed to get script settings", zap.String("server_uuid", server.Uuid), zap.Error(err))
	}

	plVar := "S_PointsLimit"
	mlVar := "S_MapsPerMatch"
	prVar := "S_PointsRepartition"
	if server.Info.LiveInfo.Type == "tmwc" || server.Info.LiveInfo.Type == "tmwt" {
		plVar = "S_MapPointsLimit"
		mlVar = "S_MatchPointsLimit"
	}

	if server.Info.LiveInfo.Type == "knockout" {
		prVar = "S_EliminatedPlayersNbRanks"
	}

	// Set points limit
	pointsLimit, ok := scriptSettings[plVar].(int)
	if !ok {
		zap.L().Debug("PointsLimit not found in script settings", zap.String("server_uuid", server.Uuid))
	} else if pointsLimit > 0 {
		server.Info.LiveInfo.PointsLimit = &pointsLimit
	}

	// Set rounds limit
	roundsLimit, ok := scriptSettings["S_RoundsPerMap"].(int)
	if !ok {
		zap.L().Debug("RoundsLimit not found in script settings", zap.String("server_uuid", server.Uuid))
	} else if roundsLimit > 0 {
		server.Info.LiveInfo.RoundsLimit = &roundsLimit
	}

	// Set map limit
	mapLimit, ok := scriptSettings[mlVar].(int)
	if !ok {
		zap.L().Debug("MapLimit not found in script settings", zap.String("server_uuid", server.Uuid))
	} else if mapLimit > 0 {
		server.Info.LiveInfo.MapLimit = &mapLimit
	}

	// Set number of winners
	nbWinners, ok := scriptSettings["S_NbOfWinners"].(int)
	if !ok {
		zap.L().Debug("NbOfWinners not found in script settings", zap.String("server_uuid", server.Uuid))
	} else if nbWinners > 0 {
		server.Info.LiveInfo.NbWinners = &nbWinners
	}

	// Set points repartition
	pointsRepartition, ok := scriptSettings[prVar].(string)
	if !ok {
		zap.L().Debug("PointsRepartition not found in script settings", zap.String("server_uuid", server.Uuid))
	} else if len(pointsRepartition) > 0 {
		repartitionList := []int{}
		for i := range strings.SplitSeq(pointsRepartition, ",") {
			value, err := strconv.Atoi(strings.TrimSpace(i))
			if err == nil {
				repartitionList = append(repartitionList, value)
			}
		}
		server.Info.LiveInfo.PointsRepartition = repartitionList
	}
}

func isFinalist(matchPoints int, pointsLimit *int) bool {
	if pointsLimit == nil {
		return false
	}

	return matchPoints == *pointsLimit
}

func isWinner(matchPoints int, pointsLimit *int) bool {
	if pointsLimit == nil {
		return false
	}

	return matchPoints > *pointsLimit
}
