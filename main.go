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

// ServerInfo defines the structure of server response data
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

// Constants
const (
	defaultTimeout = 3 * time.Second
	apiPrefix      = "/api/server/"
	contentType    = "application/json"
)

func serverHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", contentType)

	ip := r.URL.Query().Get("ip")
	if !isValidIP(ip) {
		http.Error(w, `{"error":"Missing or invalid 'ip'. Use ?ip=127.0.0.1:7777"}`, http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	server, err := sampquery.GetServerInfo(ctx, ip, true)
	info := ServerInfo{IP: ip}

	if err != nil {
		info.Error = err.Error()
		_ = json.NewEncoder(w).Encode(info)
		return
	}

	info.Hostname = server.Hostname
	info.Gamemode = server.Gamemode
	info.Version = server.Rules["version"]
	info.Players = server.Players
	info.MaxPlayers = server.MaxPlayers
	info.Passworded = server.Password
	info.IsOmp = server.IsOmp

	_ = json.NewEncoder(w).Encode(info)
}

func serverPathHandler(w http.ResponseWriter, r *http.Request) {
	ip := strings.TrimPrefix(r.URL.Path, apiPrefix)
	if !isValidIP(ip) {
		http.Error(w, `{"error":"Missing or invalid IP. Use /api/server/127.0.0.1:7777"}`, http.StatusBadRequest)
		return
	}

	// Forward to query-style handler by injecting IP into query
	q := r.URL.Query()
	q.Set("ip", ip)
	r.URL.RawQuery = q.Encode()

	serverHandler(w, r)
}

// isValidIP checks if the IP string is in host:port format
func isValidIP(ip string) bool {
	return ip != "" && strings.Contains(ip, ":")
}

func main() {
	http.HandleFunc(apiPrefix, serverPathHandler)
	http.HandleFunc("/api/server", serverHandler)

	log.Println("✅ API running on http://0.0.0.0:3000/api/server/127.0.0.1:7777 or ?ip=127.0.0.1:7777")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
