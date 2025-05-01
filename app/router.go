package app

import (
	"net/http"

	"github.com/MRegterschot/GbxConnector/handlers"
	"github.com/gorilla/mux"
)

func SetupRoutes(r *mux.Router) {
	// Health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	r.HandleFunc("/ws/servers", handlers.HandleServersConnection).Methods("GET")
	r.HandleFunc("/servers", handlers.HandleGetServers).Methods("GET")

	r.HandleFunc("/ws/listeners/{serverId:[0-9]+}", handlers.HandleListenerConnection).Methods("GET")
}
