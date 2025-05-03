package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/MRegterschot/GbxConnector/lib"
	"github.com/MRegterschot/GbxConnector/structs"
	"go.uber.org/zap"
)

// Handle GET request to retrieve server information
func HandleAuth(w http.ResponseWriter, r *http.Request) {
	var user structs.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		zap.L().Error("Failed to decode user", zap.Error(err))
		http.Error(w, "Failed to decode user", http.StatusBadRequest)
		return
	}

	token, err := lib.GenerateJWT(user)
	if err != nil {
		zap.L().Error("Failed to generate JWT token", zap.Error(err))
		http.Error(w, "Failed to generate JWT token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]string{"token": token}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		zap.L().Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
