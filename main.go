package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	sampquery "github.com/Southclaws/go-samp-query"
)

type ServerInfo struct {
	IP         string `json:"ip"`
	Hostname   string `json:"hostname"`
	Gamemode   string `json:"gamemode"`
	Mapname    string `json:"mapname"`
	Players    int    `json:"players"`
	MaxPlayers int    `json:"max_players"`
	Passworded bool   `json:"passworded"`
	Error      string `json:"error,omitempty"`
}

func queryServer(ip string, port int) (ServerInfo, error) {
	info, err := sampquery.GetServerInfo(ip, port)
	if err != nil {
		return ServerInfo{}, fmt.Errorf("info fetch failed: %v", err)
	}

	return ServerInfo{
		IP:         fmt.Sprintf("%s:%d", ip, port),
		Hostname:   info.Hostname,
		Gamemode:   info.Gamemode,
		Mapname:    info.MapName,
		Players:    info.Players,
		MaxPlayers: info.MaxPlayers,
		Passworded: info.Passworded,
	}, nil
}

func serverHandler(w http.ResponseWriter, r *http.Request) {
	ipPort := r.URL.Query().Get("ip")
	if ipPort == "" || !strings.Contains(ipPort, ":") {
		http.Error(w, `{"error":"Missing or invalid 'ip'. Use ?ip=127.0.0.1:7777"}`, http.StatusBadRequest)
		return
	}

	parts := strings.Split(ipPort, ":")
	ip := parts[0]
	var port int
	fmt.Sscanf(parts[1], "%d", &port)

	info, err := queryServer(ip, port)
	if err != nil {
		info.Error = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

func main() {
	http.HandleFunc("/api/server", serverHandler)
	port := "3000"
	log.Printf("✅ SA-MP API running at http://localhost:%s/api/server?ip=127.0.0.1:7777\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
