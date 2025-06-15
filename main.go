package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net"
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

// getAccuratePlayerCount queries detailed player list (opcode 'd') and returns actual player count.
func getAccuratePlayerCount(ctx context.Context, address string) (int, error) {
	conn, err := net.DialTimeout("udp", address, defaultTimeout)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	buf := bytes.NewBuffer([]byte("SAMP"))

	host, port, _ := net.SplitHostPort(address)
	ip := net.ParseIP(host).To4()
	if ip == nil {
		return 0, fmt.Errorf("invalid IP")
	}

	buf.Write(ip)

	var portNum uint16
	if _, err := fmt.Sscanf(port, "%d", &portNum); err != nil {
		return 0, err
	}
	_ = binary.Write(buf, binary.LittleEndian, portNum)

	buf.WriteByte('d') // opcode for detailed player list

	if _, err := conn.Write(buf.Bytes()); err != nil {
		return 0, err
	}

	_ = conn.SetReadDeadline(time.Now().Add(defaultTimeout))
	resp := make([]byte, 2048)
	n, err := conn.Read(resp)
	if err != nil {
		return 0, err
	}
	if n < 12 {
		return 0, fmt.Errorf("invalid response length")
	}

	return int(resp[11]), nil
}

// serverHandler handles both /api/server and ?ip= format
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
	info.MaxPlayers = server.MaxPlayers
	info.Passworded = server.Password
	info.IsOmp = server.IsOmp

	// Accurate player count override
	if playerCount, err := getAccuratePlayerCount(ctx, ip); err == nil {
		info.Players = playerCount
	} else {
		info.Players = server.Players // fallback
	}

	_ = json.NewEncoder(w).Encode(info)
}

// serverPathHandler handles /api/server/{ip:port}
func serverPathHandler(w http.ResponseWriter, r *http.Request) {
	ip := strings.TrimPrefix(r.URL.Path, apiPrefix)
	if !isValidIP(ip) {
		http.Error(w, `{"error":"Missing or invalid IP. Use /api/server/127.0.0.1:7777"}`, http.StatusBadRequest)
		return
	}

	// Forward to query-style handler
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
