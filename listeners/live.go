package listeners

import (
	"github.com/MRegterschot/GbxConnector/handlers"
	"github.com/MRegterschot/GbxConnector/structs"
	"github.com/MRegterschot/GbxRemoteGo/events"
	"github.com/MRegterschot/GbxRemoteGo/gbxclient"
)

type LiveListener struct {
	Server *structs.Server
}

func AddLiveListeners(server *structs.Server) {
	ll := &LiveListener{Server: server}

	server.Client.OnPlayerFinish = append(server.Client.OnPlayerFinish, gbxclient.GbxCallbackStruct[events.PlayerWayPointEventArgs]{
		Key:  "gbxconnector",
		Call: ll.onPlayerFinish,
	})

	server.Client.OnPlayerCheckpoint = append(server.Client.OnPlayerCheckpoint, gbxclient.GbxCallbackStruct[events.PlayerWayPointEventArgs]{
		Key:  "gbxconnector",
		Call: ll.onPlayerCheckpoint,
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

	handlers.BroadcastLive(ll.Server.Id, map[string]structs.PlayerWaypoint{
		"finish": playerWaypoint,
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

	handlers.BroadcastLive(ll.Server.Id, map[string]structs.PlayerWaypoint{
		"checkpoint": playerWaypoint,
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
	ll.Server.ResetLiveInfo()

	handlers.BroadcastLive(ll.Server.Id, map[string]*structs.LiveInfo{
		"beginMatch": ll.Server.Info.LiveInfo,
	})
}

func (ll *LiveListener) onPlayerGiveUp(playerGiveUpEvent events.PlayerGiveUpEventArgs) {
	handlers.BroadcastLive(ll.Server.Id, map[string]string{
		"giveUp": playerGiveUpEvent.Login,
	})
}
