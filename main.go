package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
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

func buildInfoPacket(ip string, port int) ([]byte, error) {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid IP address")
	}

	buf := []byte{'S', 'A', 'M', 'P'}
	for _, p := range parts {
		n, _ := strconv.Atoi(p)
		buf = append(buf, byte(n))
	}
	portBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(portBytes, uint16(port))
	buf = append(buf, portBytes...)
	buf = append(buf, 'i') // info packet
	return buf, nil
}

func queryServer(ip string, port int) (*ServerInfo, error) {
	addr := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("udp", addr, 2*time.Second)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	packet, err := buildInfoPacket(ip, port)
	if err != nil {
		return nil, err
	}
	_, err = conn.Write(packet)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 2048)
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}

	r := bytes.NewReader(buf[11:n]) // skip 11-byte header

	var players, maxPlayers byte
	binary.Read(r, binary.LittleEndian, &players)
	binary.Read(r, binary.LittleEndian, &maxPlayers)

	hostname := readString(r)
	gamemode := readString(r)
	language := readString(r)

	return &ServerInfo{
		IP:         fmt.Sprintf("%s:%d", ip, port),
		Port:       port,
		Hostname:   hostname,
		Gamemode:   gamemode,
		Language:   language,
		Players:    int(players),
		MaxPlayers: int(maxPlayers),
		Passworded: false, // not in info packet
		Version:    "unknown", // version not available in 'i' packet
	}, nil
}

func readString(r *bytes.Reader) string {
	var length uint32
	binary.Read(r, binary.LittleEndian, &length)
	data := make([]byte, length)
	io.ReadFull(r, data)
	return string(data)
}

func serverHandler(w http.ResponseWriter, r *http.Request) {
	ip := r.URL.Query().Get("ip")
	portStr := r.URL.Query().Get("port")
	if ip == "" || portStr == "" {
		http.Error(w, "Missing ip or port", http.StatusBadRequest)
		return
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		http.Error(w, "Invalid port", http.StatusBadRequest)
		return
	}

	info, err := queryServer(ip, port)
	if err != nil {
		http.Error(w, "Failed to query server: "+err.Error(), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

func main() {
	http.HandleFunc("/api/server", serverHandler)
	log.Println("API running at http://localhost:8080/api/server")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
