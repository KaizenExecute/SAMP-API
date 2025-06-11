package main

import (
	"encoding/binary"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "SA-MP API is running."})
	})

	app.Get("/api/server", func(c *fiber.Ctx) error {
		ip := c.Query("ip")
		portStr := c.Query("port", "7777")

		port, err := strconv.Atoi(portStr)
		if err != nil || port <= 0 || port > 65535 {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid port"})
		}

		data, err := queryServer(ip, port)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		hostname, gamemode, mapname, players, maxPlayers := parseInfo(data)

		return c.JSON(fiber.Map{
			"hostname":    hostname,
			"gamemode":    gamemode,
			"mapname":     mapname,
			"players":     players,
			"max_players": maxPlayers,
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Fatal(app.Listen(":" + port))
}

func queryServer(ip string, port int) ([]byte, error) {
	addr := net.JoinHostPort(ip, strconv.Itoa(port))

	// Build packet: SAMP + IP bytes + port + opcode
	packet := []byte{'S', 'A', 'M', 'P'}
	for _, part := range strings.Split(ip, ".") {
		b, _ := strconv.Atoi(part)
		packet = append(packet, byte(b))
	}
	packet = append(packet, byte(port&0xFF), byte((port>>8)&0xFF))
	packet = append(packet, 'i') // Info opcode

	conn, err := net.DialTimeout("udp", addr, 2*time.Second)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(2 * time.Second))
	_, err = conn.Write(packet)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}

	return buf[:n], nil
}

func parseInfo(data []byte) (string, string, string, uint16, uint16) {
	if len(data) < 11 {
		return "Unknown", "Unknown", "Unknown", 0, 0
	}

	offset := 11
	hostname, ok := readString(data, &offset)
	if !ok {
		return "Unknown", "Unknown", "Unknown", 0, 0
	}

	gamemode, ok := readString(data, &offset)
	if !ok {
		return hostname, "Unknown", "Unknown", 0, 0
	}

	mapname, ok := readString(data, &offset)
	if !ok {
		return hostname, gamemode, "Unknown", 0, 0
	}

	if offset+4 > len(data) {
		return hostname, gamemode, mapname, 0, 0
	}

	players := binary.LittleEndian.Uint16(data[offset : offset+2])
	maxPlayers := binary.LittleEndian.Uint16(data[offset+2 : offset+4])

	return hostname, gamemode, mapname, players, maxPlayers
}

func readString(data []byte, offset *int) (string, bool) {
	if *offset >= len(data) {
		return "", false
	}
	length := int(data[*offset])
	*offset++
	if *offset+length > len(data) {
		return "", false
	}
	raw := data[*offset : *offset+length]
	*offset += length

	// Filter non-printable characters
	filtered := make([]byte, 0, len(raw))
	for _, b := range raw {
		if b >= 32 && b <= 126 {
			filtered = append(filtered, b)
		}
	}

	return string(filtered), true
}
