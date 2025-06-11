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
	r.Seek(11, io.SeekStart) // Skip header

	var password byte
	binary.Read(r, binary.LittleEndian, &password)

	var players, maxPlayers uint16
	binary.Read(r, binary.LittleEndian, &players)
	binary.Read(r, binary.LittleEndian, &maxPlayers)

	readString := func(r io.Reader) string {
		var length byte
		binary.Read(r, binary.LittleEndian, &length)
		data := make([]byte, length)
		r.Read(data)
		return string(data)
	}

	hostname := readString(r)
	gamemode := readString(r)
	mapname := readString(r)

	// ===== Query: Rules (to get version) =====
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

	r2 := bytes.NewReader(buf[11:n])
	var ruleCount byte
	binary.Read(r2, binary.LittleEndian, &ruleCount)

	version := "unknown"
	for i := 0; i < int(ruleCount); i++ {
		nameLen, _ := r2.ReadByte()
		nameBytes := make([]byte, nameLen)
		r2.Read(nameBytes)

		valLen, _ := r2.ReadByte()
		valBytes := make([]byte, valLen)
		r2.Read(valBytes)

		name := string(nameBytes)
		val := string(valBytes)

		if strings.ToLower(name) == "version" {
			version = val
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
	log.Println("API running on http://localhost:8080/api/server")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
