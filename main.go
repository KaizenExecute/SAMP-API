package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	sampquery "github.com/Southclaws/go-sampquery"
)

type ServerInfo struct {
	Hostname   string `json:"hostname"`
	Gamemode   string `json:"gamemode"`
	Mapname    string `json:"mapname"`
	Players    int    `json:"players"`
	MaxPlayers int    `json:"max_players"`
	Passworded bool   `json:"passworded"`
}

type Player struct {
	Name  string `json:"name"`
	Score int    `json:"score"`
}

func queryServer(ip string) (*sampquery.ServerInfo, error) {
	hostParts := strings.Split(ip, ":")
	if len(hostParts) != 2 {
		return nil, fmt.Errorf("invalid IP format")
	}

	client, err := sampquery.NewClient(hostParts[0], hostParts[1])
	if err != nil {
		return nil, err
	}
	defer client.Close()

	info, err := client.GetServerInfo()
	if err != nil {
		return nil, err
	}
	return &info, nil
}

func queryPlayers(ip string) ([]sampquery.Player, error) {
	hostParts := strings.Split(ip, ":")
	if len(hostParts) != 2 {
		return nil, fmt.Errorf("invalid IP format")
	}

	client, err := sampquery.NewClient(hostParts[0], hostParts[1])
	if err != nil {
		return nil, err
	}
	defer client.Close()

	players, err := client.GetPlayers()
	if err != nil {
		return nil, err
	}
	return players, nil
}

func serverInfoHandler(w http.ResponseWriter, r *http.Request) {
	ip := r.URL.Query().Get("ip")
	if ip == "" {
		http.Error(w, "Missing 'ip' parameter", http.StatusBadRequest)
		return
	}

	info, err := queryServer(ip)
	if err != nil {
		http.Error(w, "Server not reachable or invalid: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := ServerInfo{
		Hostname:   info.Hostname,
		Gamemode:   info.Gamemode,
		Mapname:    info.MapName,
		Players:    info.Players,
		MaxPlayers: info.MaxPlayers,
		Passworded: info.Passworded,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func playersHandler(w http.ResponseWriter, r *http.Request) {
	ip := r.URL.Query().Get("ip")
	if ip == "" {
		http.Error(w, "Missing 'ip' parameter", http.StatusBadRequest)
		return
	}

	players, err := queryPlayers(ip)
	if err != nil {
		http.Error(w, "Failed to get players: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var playerList []Player
	for _, p := range players {
		playerList = append(playerList, Player{Name: p.Name, Score: p.Score})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(playerList)
}

func main() {
	http.HandleFunc("/api/server", serverInfoHandler)
	http.HandleFunc("/api/players", playersHandler)

	srv := &http.Server{
		Addr:         ":3000",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Println("🚀 API running at http://localhost:3000")
	if err := srv.ListenAndServe(); err != nil {
		fmt.Println("Server failed:", err)
	}
}
