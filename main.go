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

	// Safe version check
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

func serverPathHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/server/")
	if path == "" || !strings.Contains(path, ":") {
		http.Error(w, `{"error":"Missing or invalid IP. Use /api/server/127.0.0.1:7777"}`, http.StatusBadRequest)
		return
	}

	// Reuse the serverHandler by injecting the IP as a query param
	q := r.URL.Query()
	q.Set("ip", path)
	r.URL.RawQuery = q.Encode()

	serverHandler(w, r)
}

func main() {
	http.HandleFunc("/api/server/", serverPathHandler) // path-style endpoint
	http.HandleFunc("/api/server", serverHandler)      // query-style endpoint
	log.Println("✅ API running on http://0.0.0.0:3000/api/server/127.0.0.1:7777 or ?ip=127.0.0.1:7777")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
