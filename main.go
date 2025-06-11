package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/gofiber/fiber/v2"
)

type ServerInfo struct {
	Hostname   string `json:"hostname"`
	Gamemode   string `json:"gamemode"`
	Mapname    string `json:"mapname"`
	Players    uint16 `json:"players"`
	MaxPlayers uint16 `json:"max_players"`
}

func main() {
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "SA-MP/Open.MP API is running."})
	})

	app.Get("/api/server", func(c *fiber.Ctx) error {
		ip := c.Query("ip")
		port := c.QueryInt("port", 7777)

		if ip == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Missing 'ip' query parameter"})
		}

		info, err := queryServerInfo(ip, port)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(info)
	})

	log.Fatal(app.Listen(":3000"))
}

func queryServerInfo(ip string, port int) (*ServerInfo, error) {
	addr := net.UDPAddr{
		IP:   net.ParseIP(ip),
		Port: port,
	}
	conn, err := net.DialUDP("udp", nil, &addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(2 * time.Second))

	// Build SAMP packet
	packet := []byte("SAMP")
	for _, b := range net.ParseIP(ip).To4() {
		packet = append(packet, b)
	}
	packet = append(packet, byte(port&0xFF), byte((port>>8)&0xFF))
	packet = append(packet, 'i')

	_, err = conn.Write(packet)
	if err != nil {
		return nil, err
	}

	// Receive response
	buf := make([]byte, 4096)
	n, _, err := conn.ReadFromUDP(buf)
	if err != nil {
		return nil, err
	}
	if n < 11 {
		return nil, fmt.Errorf("invalid response length")
	}

	offset := 11

	readString := func() (string, error) {
		if offset >= n {
			return "", fmt.Errorf("offset out of range")
		}
		length := int(buf[offset])
		offset++
		if offset+length > n {
			return "", fmt.Errorf("string out of bounds")
		}
		s := string(buf[offset : offset+length])
		offset += length
		return s, nil
	}

	hostname, err := readString()
	if err != nil {
		return nil, err
	}
	gamemode, err := readString()
	if err != nil {
		return nil, err
	}
	mapname, err := readString()
	if err != nil {
		return nil, err
	}

	if offset+4 > n {
		return nil, fmt.Errorf("missing player counts")
	}
	players := binary.LittleEndian.Uint16(buf[offset : offset+2])
	maxPlayers := binary.LittleEndian.Uint16(buf[offset+2 : offset+4])

	return &ServerInfo{
		Hostname:   hostname,
		Gamemode:   gamemode,
		Mapname:    mapname,
		Players:    players,
		MaxPlayers: maxPlayers,
	}, nil
}
