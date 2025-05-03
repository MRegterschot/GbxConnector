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
	r.HandleFunc("/servers", handlers.HandleAddServer).Methods("POST")
	r.HandleFunc("/servers/{id:[0-9]+}", handlers.HandleDeleteServer).Methods("DELETE")
	r.HandleFunc("/servers/{id:[0-9]+}", handlers.HandleUpdateServer).Methods("PUT")
	r.HandleFunc("/servers/order", handlers.HandleOrderServers).Methods("PUT")

	r.HandleFunc("/ws/listeners/{id:[0-9]+}", handlers.HandleListenerConnection).Methods("GET")
}
