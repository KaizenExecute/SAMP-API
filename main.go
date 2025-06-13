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

func queryBasic(ip string, port int) (ServerInfo, error) {
	address := fmt.Sprintf("%s:%d", ip, port)
	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return ServerInfo{}, err
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return ServerInfo{}, err
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(2 * time.Second))

	// SA-MP query header: SAMP + ip + port + opcode (0x69 = info)
	req := []byte{'S', 'A', 'M', 'P'}
	for _, b := range strings.Split(ip, ".") {
		var n byte
		fmt.Sscanf(b, "%d", &n)
		req = append(req, n)
	}
	req = append(req, byte(port&0xFF), byte((port>>8)&0xFF))
	req = append(req, 'i') // 'i' = 0x69 = basic info

	_, err = conn.Write(req)
	if err != nil {
		return ServerInfo{}, err
	}

	buffer := make([]byte, 512)
	n, _, err := conn.ReadFromUDP(buffer)
	if err != nil {
		return ServerInfo{}, err
	}

	if n < 11 {
		return ServerInfo{}, fmt.Errorf("invalid response")
	}

	offset := 11 // skip header

	readString := func() string {
		length := int(buffer[offset])
		offset++
		str := string(buffer[offset : offset+length])
		offset += length
		return str
	}

	hostname := readString()
	gamemode := readString()
	mapname := readString()
	players := int(buffer[offset])
	offset++
	maxPlayers := int(buffer[offset])
	offset++
	passworded := false
	if len(buffer) > offset && buffer[offset] == 1 {
		passworded = true
	}

	return ServerInfo{
		IP:         fmt.Sprintf("%s:%d", ip, port),
		Hostname:   hostname,
		Gamemode:   gamemode,
		Mapname:    mapname,
		Players:    players,
		MaxPlayers: maxPlayers,
		Passworded: passworded,
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

	info, err := queryBasic(ip, port)
	if err != nil {
		info.Error = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

func main() {
	http.HandleFunc("/api/server", serverHandler)
	log.Println("✅ Running on http://localhost:3000/api/server?ip=127.0.0.1:7777")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
