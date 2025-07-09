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

	adminOnly := middleware.RequireRoles(true)

	r.HandleFunc("/auth", handlers.HandleAuth).Methods("POST")

	r.Handle("/ws/servers", adminOnly(http.HandlerFunc(handlers.HandleServersConnection))).Methods("GET")
	r.Handle("/servers", adminOnly(http.HandlerFunc(handlers.HandleGetServers))).Methods("GET")
	r.Handle("/servers", adminOnly(http.HandlerFunc(handlers.HandleAddServer))).Methods("POST")
	r.Handle("/servers/{uuid:[0-9a-fA-F-]{36}}", adminOnly(http.HandlerFunc(handlers.HandleDeleteServer))).Methods("DELETE")
	r.Handle("/servers/{uuid:[0-9a-fA-F-]{36}}", adminOnly(http.HandlerFunc(handlers.HandleUpdateServer))).Methods("PUT")
	r.Handle("/servers/order", adminOnly(http.HandlerFunc(handlers.HandleOrderServers))).Methods("PUT")
	r.Handle("/chat/{uuid:[0-9a-fA-F-]{36}}/config", adminOnly(http.HandlerFunc(handlers.HandleGetChatConfig))).Methods("GET")
	r.Handle("/chat/{uuid:[0-9a-fA-F-]{36}}/config", adminOnly(http.HandlerFunc(handlers.HandleUpdateChatConfig))).Methods("PUT")

	r.Handle("/ws/map/{uuid:[0-9a-fA-F-]{36}}", adminOnly(http.HandlerFunc(handlers.HandleMapConnection))).Methods("GET")
	r.Handle("/ws/players/{uuid:[0-9a-fA-F-]{36}}", adminOnly(http.HandlerFunc(handlers.HandlePlayersConnection))).Methods("GET")
	r.Handle("/ws/live/{uuid:[0-9a-fA-F-]{36}}", adminOnly(http.HandlerFunc(handlers.HandleLiveConnection))).Methods("GET")
}
