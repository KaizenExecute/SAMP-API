package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
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
	address := fmt.Sprintf("%s:%d", ip, port)
	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return ServerInfo{}, fmt.Errorf("resolve error: %v", err)
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return ServerInfo{}, fmt.Errorf("dial error: %v", err)
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(2 * time.Second))

	packet := []byte{'S', 'A', 'M', 'P'}
	for _, part := range strings.Split(ip, ".") {
		var b byte
		fmt.Sscanf(part, "%d", &b)
		packet = append(packet, b)
	}
	packet = append(packet, byte(port&0xFF), byte((port>>8)&0xFF))
	packet = append(packet, 'i') // info opcode

	_, err = conn.Write(packet)
	if err != nil {
		return ServerInfo{}, fmt.Errorf("write error: %v", err)
	}

	buffer := make([]byte, 512)
	n, _, err := conn.ReadFromUDP(buffer)
	if err != nil {
		return ServerInfo{}, fmt.Errorf("read error: %v", err)
	}

	if n < 11 {
		return ServerInfo{}, fmt.Errorf("invalid response")
	}

	offset := 11 // skip header

	readString := func() string {
		length := int(buffer[offset])
		offset++
		s := string(buffer[offset : offset+length])
		offset += length
		return s
	}

	hostname := readString()
	gamemode := readString()
	mapname := readString()
	players := int(buffer[offset])
	offset++
	maxPlayers := int(buffer[offset])
	offset++

	return ServerInfo{
		IP:         fmt.Sprintf("%s:%d", ip, port),
		Hostname:   hostname,
		Gamemode:   gamemode,
		Mapname:    mapname,
		Players:    players,
		MaxPlayers: maxPlayers,
		Passworded: false,
	}, nil
}

func serverHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	ipParam := r.URL.Query().Get("ip")
	if ipParam == "" || !strings.Contains(ipParam, ":") {
		http.Error(w, `{"error":"Missing or invalid 'ip'. Use ?ip=127.0.0.1:7777"}`, http.StatusBadRequest)
		return
	}

	parts := strings.Split(ipParam, ":")
	ip := parts[0]
	var port int
	fmt.Sscanf(parts[1], "%d", &port)

	info, err := queryServer(ip, port)
	if err != nil {
		info = ServerInfo{
			IP:    fmt.Sprintf("%s:%d", ip, port),
			Error: err.Error(),
		}
	}

	json.NewEncoder(w).Encode(info)
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "online",
		"message": "SA-MP/Open.MP API is running!",
		"usage":   "/api/server?ip=IP:PORT",
	})
}

func main() {
	http.HandleFunc("/api/server", serverHandler)
	http.HandleFunc("/api", statusHandler) // ✅ Add this line

	port := "3000"
	log.Printf("✅ API listening on http://0.0.0.0:%s/api", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
