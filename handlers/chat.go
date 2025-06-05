package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/MRegterschot/GbxConnector/config"
	"github.com/MRegterschot/GbxConnector/structs"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func HandleGetChatConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverIDStr := vars["id"]

	serverId, err := strconv.Atoi(serverIDStr)
	if err != nil {
		zap.L().Error("Invalid server ID", zap.Error(err))
		http.Error(w, "Invalid server ID", http.StatusBadRequest)
		return
	}

	for _, server := range config.AppEnv.Servers {
		if server.Id == serverId {
			if err := json.NewEncoder(w).Encode(server.Info.Chat); err != nil {
				zap.L().Error("Failed to encode chat config", zap.Error(err))
				http.Error(w, "Failed to encode chat config", http.StatusInternalServerError)
			}
			return
		}
	}

	zap.L().Error("Server not found", zap.Int("server_id", serverId))
	http.Error(w, "Server not found", http.StatusNotFound)
}

func HandleUpdateChatConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverIDStr := vars["id"]

	serverId, err := strconv.Atoi(serverIDStr)
	if err != nil {
		zap.L().Error("Invalid server ID", zap.Error(err))
		http.Error(w, "Invalid server ID", http.StatusBadRequest)
		return
	}

	var chatConfig structs.ChatConfig
	if err := json.NewDecoder(r.Body).Decode(&chatConfig); err != nil {
		zap.L().Error("Failed to decode chat config", zap.Error(err))
		http.Error(w, "Failed to decode chat config", http.StatusBadRequest)
		return
	}

	for _, server := range config.AppEnv.Servers {
		if server.Id == serverId {
			if err := server.Client.ChatEnableManualRouting(chatConfig.ManualRouting, true); err != nil {
				zap.L().Error("Failed to set manual routing", zap.Error(err))
				http.Error(w, "Failed to set manual routing", http.StatusInternalServerError)
				return
			}

			server.Info.Chat = chatConfig
			zap.L().Info("Updated chat config", zap.Int("server_id", server.Id), zap.Any("chat_config", chatConfig))
			if err := json.NewEncoder(w).Encode(server.Info.Chat); err != nil {
				zap.L().Error("Failed to encode updated chat config", zap.Error(err))
				http.Error(w, "Failed to encode updated chat config", http.StatusInternalServerError)
			}
			return
		}
	}

	zap.L().Error("Server not found", zap.Int("server_id", serverId))
	http.Error(w, "Server not found", http.StatusNotFound)
}
