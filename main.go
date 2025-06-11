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
	"strings"
	"time"
)

type ServerInfo struct {
	Hostname   string `json:"hostname"`
	Gamemode   string `json:"gamemode"`
	Mapname    string `json:"mapname"`
	Players    int    `json:"players"`
	MaxPlayers int    `json:"maxplayers"`
	Passworded bool   `json:"passworded"`
	Version    string `json:"version"`
}

func readString(r io.Reader) string {
	var lengthByte [1]byte
	_, err := r.Read(lengthByte[:])
	if err != nil {
		return ""
	}
	length := lengthByte[0]

	strBytes := make([]byte, length)
	_, err = r.Read(strBytes)
	if err != nil {
		return ""
	}

	return string(strBytes)
}

func queryServer(ip string, port string) (*ServerInfo, error) {
	addr := fmt.Sprintf("%s:%s", ip, port)
	conn, err := net.DialTimeout("udp", addr, 2*time.Second)
	if err != nil {
		return nil, fmt.Errorf("unable to connect: %v", err)
	}
	defer conn.Close()

	// Build base packet header
	ipParts := strings.Split(ip, ".")
	if len(ipParts) != 4 {
		return nil, fmt.Errorf("invalid IP format")
	}

	packetHeader := []byte{'S', 'A', 'M', 'P'}
	for _, part := range ipParts {
		var b byte
		fmt.Sscanf(part, "%d", &b)
		packetHeader = append(packetHeader, b)
	}
	var portNum uint16
	fmt.Sscanf(port, "%d", &portNum)
	portBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(portBytes, portNum)
	packetHeader = append(packetHeader, portBytes...)

	// ===== Query: Info =====
	infoPacket := append(append([]byte{}, packetHeader...), 'i')
	_, err = conn.Write(infoPacket)
	if err != nil {
		return nil, fmt.Errorf("info query failed: %v", err)
	}

	buf := make([]byte, 512)
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, err := conn.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("no info response: %v", err)
	}

	r := bytes.NewReader(buf[:n])
	r.Seek(11, io.SeekStart) // Skip SAMP header

	var password byte
	binary.Read(r, binary.LittleEndian, &password)

	var players, maxPlayers uint16
	binary.Read(r, binary.LittleEndian, &players)
	binary.Read(r, binary.LittleEndian, &maxPlayers)

	hostname := readString(r)
	gamemode := readString(r)
	mapname := readString(r)

	// ===== Query: Rules =====
	rulesPacket := append(append([]byte{}, packetHeader...), 'r')
	_, err = conn.Write(rulesPacket)
	if err != nil {
		return nil, fmt.Errorf("rules query failed: %v", err)
	}

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, err = conn.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("no rules response: %v", err)
	}

	r2 := bytes.NewReader(buf[11:n]) // skip header
	var ruleCount byte
	binary.Read(r2, binary.LittleEndian, &ruleCount)

	version := "unknown"
	for i := 0; i < int(ruleCount); i++ {
		key := readString(r2)
		value := readString(r2)
		if strings.ToLower(key) == "version" {
			version = value
			break
		}
	}

	return &ServerInfo{
		Hostname:   hostname,
		Gamemode:   gamemode,
		Mapname:    mapname,
		Players:    int(players),
		MaxPlayers: int(maxPlayers),
		Passworded: password == 1,
		Version:    version,
	}, nil
}

func serverHandler(w http.ResponseWriter, r *http.Request) {
	ip := r.URL.Query().Get("ip")
	port := r.URL.Query().Get("port")
	if ip == "" || port == "" {
		http.Error(w, "Missing ip or port", http.StatusBadRequest)
		return
	}

	info, err := queryServer(ip, port)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

func main() {
	http.HandleFunc("/api/server", serverHandler)
	log.Println("API running at http://localhost:8080/api/server")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
