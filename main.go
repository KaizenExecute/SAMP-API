package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type ServerInfo struct {
	IP         string `json:"ip"`
	Port       int    `json:"port"`
	Hostname   string `json:"hostname"`
	Gamemode   string `json:"gamemode"`
	Language   string `json:"language"`
	Players    int    `json:"players"`
	MaxPlayers int    `json:"maxPlayers"`
	Passworded bool   `json:"passworded"`
	Version    string `json:"version"`
}

func serverHandler(w http.ResponseWriter, r *http.Request) {
	ip := r.URL.Query().Get("ip")
	port := r.URL.Query().Get("port")
	if ip == "" || port == "" {
		http.Error(w, "Missing ip or port", http.StatusBadRequest)
		return
	}

	apiURL := fmt.Sprintf("https://api.open.mp/servers/%s:%s", ip, port)

	resp, err := http.Get(apiURL)
	if err != nil || resp.StatusCode != 200 {
		http.Error(w, "Server not found or Open.MP API error", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var info ServerInfo
	if err := json.Unmarshal(body, &info); err != nil {
		http.Error(w, "Failed to parse Open.MP response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

func main() {
	http.HandleFunc("/api/server", serverHandler)
	log.Println("Proxy API running at http://localhost:8080/api/server")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
