package listeners

import (
	"github.com/MRegterschot/GbxConnector/handlers"
	"github.com/MRegterschot/GbxConnector/lib"
	"github.com/MRegterschot/GbxConnector/structs"
	"go.uber.org/zap"
)

type MapListener struct {
	Server *structs.Server
}

func AddMapListeners(server *structs.Server) {
	ml := &MapListener{Server: server}
	server.Client.AddScriptCallback("Maniaplanet.EndMap_Start", "EndMapListener", ml.onEndMap)
	server.Client.AddScriptCallback("Maniaplanet.StartMap_Start", "StartMapListener", ml.onStartMap)
}

func (ml *MapListener) onEndMap(data any) {
	var event structs.MapEvent
	if err := lib.ConvertCallbackData(data, &event); err != nil {
		zap.L().Error("Failed to cast data to MapEvent", zap.Error(err))
		return
	}

	ml.Server.ActiveMap = event.Map.UId

	handlers.BroadcastListener(ml.Server.Id, map[string]string{
		"endMap": event.Map.UId,
	})
}

func (ml *MapListener) onStartMap(data any) {
	var event structs.MapEvent
	if err := lib.ConvertCallbackData(data, &event); err != nil {
		zap.L().Error("Failed to cast data to MapEvent", zap.Error(err))
		return
	}

	ml.Server.ActiveMap = event.Map.UId

	handlers.BroadcastListener(ml.Server.Id, map[string]string{
		"startMap": event.Map.UId,
	})
}
