package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Southclaws/go-samp-query/query"
)

type ServerInfo struct {
	IP         string `json:"ip"`
	Hostname   string `json:"hostname"`
	Gamemode   string `json:"gamemode"`
	Mapname    string `json:"mapname"`
	Version    string `json:"version"`
	Players    int    `json:"players"`
	MaxPlayers int    `json:"max_players"`
	Passworded bool   `json:"passworded"`
	Error      string `json:"error,omitempty"`
}

func serverHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query().Get("ip")
	if !strings.Contains(q, ":") {
		http.Error(w, `{"error":"Use format ?ip=IP:PORT"}`, http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	server, err := sampquery.GetServerInfo(ctx, q, true)
	out := ServerInfo{IP: q}
	if err != nil {
		out.Error = err.Error()
	} else {
		out.Hostname = server.Hostname
		out.Gamemode = server.Gamemode
		out.Mapname = server.Language
		out.Version = server.Version
		out.Players = server.Players
		out.MaxPlayers = server.MaxPlayers
		out.Passworded = server.Passworded
	}
	json.NewEncoder(w).Encode(out)
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "online",
		"message": "SA‑MP/Open.MP API (v1.2.4) is running!",
		"usage":   "/api/server?ip=IP:PORT",
	})
}

func main() {
	http.HandleFunc("/api/server", serverHandler)
	http.HandleFunc("/api", statusHandler)
	log.Printf("🚀 Running on http://localhost:3000/api")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
