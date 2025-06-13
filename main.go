package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	sampquery "github.com/Southclaws/go-samp-query"
)

type ServerInfo struct {
	Hostname   string `json:"hostname"`
	Gamemode   string `json:"gamemode"`
	Mapname    string `json:"mapname"`
	Players    int    `json:"players"`
	MaxPlayers int    `json:"max_players"`
	Passworded bool   `json:"passworded"`
	Language   string `json:"language"`
}

func queryServer(ip string) (*ServerInfo, error) {
	if !strings.Contains(ip, ":") {
		return nil, fmt.Errorf("invalid IP format. Use IP:PORT")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	info, err := sampquery.GetServerInfo(ctx, ip, true)
	if err != nil {
		return nil, err
	}

	return &ServerInfo{
		Hostname:   info.Hostname,
		Gamemode:   info.GameMode,
		Mapname:    info.MapName,
		Players:    info.Players,
		MaxPlayers: info.MaxPlayers,
		Passworded: info.Passworded,
		Language:   info.Language,
	}, nil
}

func serverHandler(w http.ResponseWriter, r *http.Request) {
	ip := r.URL.Query().Get("ip")
	if ip == "" {
		http.Error(w, "Missing `ip` query param. Use ?ip=IP:PORT", http.StatusBadRequest)
		return
	}

	info, err := queryServer(ip)
	if err != nil {
		http.Error(w, "Server error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

func main() {
	http.HandleFunc("/api/server", serverHandler)

	fmt.Println("🚀 SA-MP Monitor API running at http://localhost:3000")
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		fmt.Println("Server failed to start:", err)
	}
}
