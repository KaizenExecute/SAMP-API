package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	sampquery "github.com/Southclaws/go-samp-query"
)

type ServerInfo struct {
	IP         string `json:"ip"`
	Hostname   string `json:"hostname"`
	Gamemode   string `json:"gamemode"`
	Version    string `json:"version"`
	Players    int    `json:"players"`
	MaxPlayers int    `json:"max_players"`
	Passworded bool   `json:"passworded"`
	IsOmp      bool   `json:"isOmp"`
	Error      string `json:"error,omitempty"`
}

type Player struct {
	Name  string `json:"name"`
	Score int    `json:"score"`
	Ping  int    `json:"ping"`
}

// Custom function to query player list using sampquery
func GetPlayers(ctx context.Context, address string, detectOmp bool) ([]Player, error) {
	client := sampquery.NewClient()
	resp, err := client.Query(ctx, address, sampquery.PACKET_TYPE_PLAYER, detectOmp)
	if err != nil {
		return nil, err
	}

	players := make([]Player, len(resp.Players))
	for i, p := range resp.Players {
		players[i] = Player{
			Name:  p.Name,
			Score: p.Score,
			Ping:  p.Ping,
		}
	}

	return players, nil
}

func serverHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	ip := r.URL.Query().Get("ip")
	if ip == "" || !strings.Contains(ip, ":") {
		http.Error(w, `{"error":"Missing or invalid 'ip'. Use ?ip=127.0.0.1:7777"}`, http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	server, err := sampquery.GetServerInfo(ctx, ip, true)
	info := ServerInfo{IP: ip}

	if err != nil {
		info.Error = err.Error()
		json.NewEncoder(w).Encode(info)
		return
	}

	version := ""
	if v, ok := server.Rules["version"]; ok {
		version = v
	}

	info.Hostname = server.Hostname
	info.Gamemode = server.Gamemode
	info.Version = version
	info.Players = server.Players
	info.MaxPlayers = server.MaxPlayers
	info.Passworded = server.Password
	info.IsOmp = server.IsOmp

	json.NewEncoder(w).Encode(info)
}

func playersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	ip := r.URL.Query().Get("ip")
	if ip == "" || !strings.Contains(ip, ":") {
		http.Error(w, `{"error":"Missing or invalid 'ip'. Use ?ip=127.0.0.1:7777"}`, http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	players, err := GetPlayers(ctx, ip, true)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(players)
}

func main() {
	http.HandleFunc("/api/server", serverHandler)
	http.HandleFunc("/api/players", playersHandler)

	log.Println("✅ API running at http://0.0.0.0:3000/api/server?ip=127.0.0.1:7777")
	log.Println("✅ Player list at http://0.0.0.0:3000/api/players?ip=127.0.0.1:7777")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
