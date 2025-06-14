package main

import (
	"bytes"
	"encoding/binary"
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
	Version    string `json:"version"`
	Players    int    `json:"players"`
	MaxPlayers int    `json:"max_players"`
	Passworded bool   `json:"passworded"`
	IsOmp      bool   `json:"is_omp"`
	Error      string `json:"error,omitempty"`
}

type Player struct {
	Name  string `json:"name"`
	Score int    `json:"score"`
	Ping  int    `json:"ping"`
}

// SendQuery sends a UDP packet to the SA-MP/Open.MP server and returns the response
func SendQuery(address string, opcode byte) ([]byte, error) {
	conn, err := net.DialTimeout("udp", address, 3*time.Second)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	packet := []byte{'S', 'A', 'M', 'P'}
	host, portStr, _ := net.SplitHostPort(address)
	port := parsePort(portStr)

	ip := net.ParseIP(host).To4()
	packet = append(packet, ip[0], ip[1], ip[2], ip[3])
	packet = append(packet, byte(port&0xFF), byte((port>>8)&0xFF))
	packet = append(packet, opcode)

	_, err = conn.Write(packet)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 2048)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}

	return buf[:n], nil
}

// QueryServerInfo gets info from a SA-MP or Open.MP server
func QueryServerInfo(address string) (ServerInfo, error) {
	resp, err := SendQuery(address, 'i')
	if err != nil {
		return ServerInfo{IP: address, Error: err.Error()}, err
	}

	info := ServerInfo{IP: address}

	// Skip "SAMP" header
	r := bytes.NewReader(resp[11:])

	info.Hostname = readString(r)
	info.Gamemode = readString(r)
	info.Mapname = readString(r)
	binary.Read(r, binary.LittleEndian, &info.Players)
	binary.Read(r, binary.LittleEndian, &info.MaxPlayers)
	password := byte(0)
	binary.Read(r, binary.LittleEndian, &password)
	info.Passworded = password == 1

	// Open.MP detection
	if strings.Contains(strings.ToLower(info.Gamemode), "open.mp") ||
		strings.Contains(strings.ToLower(info.Hostname), "open.mp") {
		info.IsOmp = true
	}

	return info, nil
}

// QueryPlayers gets player names, scores, and pings
func QueryPlayers(address string) ([]Player, error) {
	resp, err := SendQuery(address, 'd')
	if err != nil {
		return nil, err
	}

	r := bytes.NewReader(resp[11:]) // skip header
	var count uint16
	binary.Read(r, binary.LittleEndian, &count)

	var players []Player
	for i := 0; i < int(count); i++ {
		name := readString(r)
		var score int32
		binary.Read(r, binary.LittleEndian, &score)
		var ping int32
		binary.Read(r, binary.LittleEndian, &ping)

		players = append(players, Player{
			Name:  name,
			Score: int(score),
			Ping:  int(ping),
		})
	}

	return players, nil
}

func readString(r *bytes.Reader) string {
	var len byte
	binary.Read(r, binary.LittleEndian, &len)
	str := make([]byte, len)
	r.Read(str)
	return string(str)
}

func parsePort(p string) uint16 {
	var port uint16
	fmt.Sscanf(p, "%d", &port)
	return port
}

// API Handlers

func serverHandler(w http.ResponseWriter, r *http.Request) {
	ip := r.URL.Query().Get("ip")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if ip == "" || !strings.Contains(ip, ":") {
		http.Error(w, `{"error":"Missing or invalid 'ip'. Use ?ip=127.0.0.1:7777"}`, 400)
		return
	}

	info, err := QueryServerInfo(ip)
	if err != nil {
		info.Error = err.Error()
	}

	json.NewEncoder(w).Encode(info)
}

func playersHandler(w http.ResponseWriter, r *http.Request) {
	ip := r.URL.Query().Get("ip")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if ip == "" || !strings.Contains(ip, ":") {
		http.Error(w, `{"error":"Missing or invalid 'ip'. Use ?ip=127.0.0.1:7777"}`, 400)
		return
	}

	players, err := QueryPlayers(ip)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, 500)
		return
	}

	json.NewEncoder(w).Encode(players)
}

func main() {
	http.HandleFunc("/api/server", serverHandler)
	http.HandleFunc("/api/players", playersHandler)

	log.Println("✅ Running at http://localhost:3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
