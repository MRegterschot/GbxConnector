package app

import (
	"net/http"

	"github.com/MRegterschot/GbxConnector/handlers"
	"github.com/MRegterschot/GbxConnector/middleware"
	"github.com/gorilla/mux"
)

func SetupRoutes(r *mux.Router) {
	// Health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	adminOnly := middleware.RequireRoles("admin")
	moderatorOrAdmin := middleware.RequireRoles("moderator", "admin")

	r.HandleFunc("/auth", handlers.HandleAuth).Methods("POST")

	r.Handle("/ws/servers", moderatorOrAdmin(http.HandlerFunc(handlers.HandleServersConnection))).Methods("GET")
	r.Handle("/servers", moderatorOrAdmin(http.HandlerFunc(handlers.HandleGetServers))).Methods("GET")
	r.Handle("/servers", adminOnly(http.HandlerFunc(handlers.HandleAddServer))).Methods("POST")
	r.Handle("/servers/{id:[0-9]+}", adminOnly(http.HandlerFunc(handlers.HandleDeleteServer))).Methods("DELETE")
	r.Handle("/servers/{id:[0-9]+}", adminOnly(http.HandlerFunc(handlers.HandleUpdateServer))).Methods("PUT")
	r.Handle("/servers/order", adminOnly(http.HandlerFunc(handlers.HandleOrderServers))).Methods("PUT")

	r.Handle("/ws/map/{id:[0-9]+}", moderatorOrAdmin(http.HandlerFunc(handlers.HandleMapConnection))).Methods("GET")
	r.Handle("/ws/players/{id:[0-9]+}", moderatorOrAdmin(http.HandlerFunc(handlers.HandlePlayersConnection))).Methods("GET")
	r.Handle("/ws/live/{id:[0-9]+}", moderatorOrAdmin(http.HandlerFunc(handlers.HandleLiveConnection))).Methods("GET")
}
