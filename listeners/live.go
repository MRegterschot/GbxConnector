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
		Key:  "PlayerFinishListener",
		Call: ll.onPlayerFinish,
	})

	server.Client.OnPlayerCheckpoint = append(server.Client.OnPlayerCheckpoint, gbxclient.GbxCallbackStruct[events.PlayerWayPointEventArgs]{
		Key:  "PlayerCheckpointListener",
		Call: ll.onPlayerCheckpoint,
	})

	server.Client.OnEndRound = append(server.Client.OnEndRound, gbxclient.GbxCallbackStruct[events.ScoresEventArgs]{
		Key:  "EndRoundListener",
		Call: ll.onEndRound,
	})

	server.Client.OnBeginMap = append(server.Client.OnBeginMap, gbxclient.GbxCallbackStruct[events.MapEventArgs]{
		Key:  "BeginMapListener",
		Call: ll.onBeginMap,
	})

	server.Client.OnEndMap = append(server.Client.OnEndMap, gbxclient.GbxCallbackStruct[events.MapEventArgs]{
		Key:  "EndMapListener",
		Call: ll.onEndMap,
	})

	server.Client.OnBeginMatch = append(server.Client.OnBeginMatch, gbxclient.GbxCallbackStruct[struct{}]{
		Key:  "BeginMatchListener",
		Call: ll.onBeginMatch,
	})

	server.Client.OnEndMatch = append(server.Client.OnEndMatch, gbxclient.GbxCallbackStruct[events.EndMatchEventArgs]{
		Key:  "EndMatchListener",
		Call: ll.onEndMatch,
	})

	server.Client.OnPlayerGiveUp = append(server.Client.OnPlayerGiveUp, gbxclient.GbxCallbackStruct[events.PlayerGiveUpEventArgs]{
		Key:  "PlayerGiveUpListener",
		Call: ll.onPlayerGiveUp,
	})

	server.Client.OnStartLine = append(server.Client.OnStartLine, gbxclient.GbxCallbackStruct[events.StartLineEventArgs]{
		Key:  "StartLineListener",
		Call: ll.onStartLine,
	})
}

func (ll *LiveListener) onPlayerFinish(playerFinishEvent events.PlayerWayPointEventArgs) {
	handlers.BroadcastLive(ll.Server.Id, map[string]any{
		"finish": playerFinishEvent,
	})
}

func (ll *LiveListener) onPlayerCheckpoint(playerCheckpointEvent events.PlayerWayPointEventArgs) {
	handlers.BroadcastLive(ll.Server.Id, map[string]any{
		"checkpoint": playerCheckpointEvent,
	})
}

func (ll *LiveListener) onEndRound(endRoundEvent events.ScoresEventArgs) {
	handlers.BroadcastLive(ll.Server.Id, map[string]any{
		"endRound": endRoundEvent,
	})
}

func (ll *LiveListener) onBeginMap(beginMapEvent events.MapEventArgs) {
	handlers.BroadcastLive(ll.Server.Id, map[string]any{
		"beginMap": beginMapEvent,
	})
}

func (ll *LiveListener) onEndMap(endMapEvent events.MapEventArgs) {
	handlers.BroadcastLive(ll.Server.Id, map[string]any{
		"endMap": endMapEvent,
	})
}

func (ll *LiveListener) onBeginMatch(_ struct{}) {
	handlers.BroadcastLive(ll.Server.Id, map[string]any{
		"beginMatch": true,
	})
}

func (ll *LiveListener) onEndMatch(endMatchEvent events.EndMatchEventArgs) {
	handlers.BroadcastLive(ll.Server.Id, map[string]any{
		"endMatch": endMatchEvent,
	})
}

func (ll *LiveListener) onPlayerGiveUp(playerGiveUpEvent events.PlayerGiveUpEventArgs) {
	handlers.BroadcastLive(ll.Server.Id, map[string]any{
		"giveUp": playerGiveUpEvent,
	})
}

func (ll *LiveListener) onStartLine(startLineEvent events.StartLineEventArgs) {
	handlers.BroadcastLive(ll.Server.Id, map[string]any{
		"startLine": startLineEvent,
	})
}
