package main

import (
	"encoding/binary"
	"log"
	"net"
	"os"
	"strconv"   // ✅ This line fixes the error
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

func queryServer(ip string, port int) ([]byte, error) {
	addr := net.JoinHostPort(ip, stringPort(port))
	packet := []byte{'S', 'A', 'M', 'P'}
	for _, b := range strings.Split(ip, ".") {
		packet = append(packet, byte(atoi(b)))
	}
	packet = append(packet, byte(port&0xFF), byte((port>>8)&0xFF))
	packet = append(packet, 'i') // 'i' = info

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

func parseString(data []byte, offset int) (string, int, error) {
	if offset >= len(data) {
		return "", offset, fiber.ErrBadRequest
	}
	length := int(data[offset])
	offset++
	if offset+length > len(data) {
		return "", offset, fiber.ErrBadRequest
	}
	raw := data[offset : offset+length]
	clean := make([]rune, 0, length)
	for _, b := range raw {
		if b >= 32 && b <= 126 {
			clean = append(clean, rune(b))
		}
	}
	return string(clean), offset + length, nil
}

func stringPort(port int) string {
	return strconv.Itoa(port)
}

func atoi(s string) int {
	n := 0
	for _, r := range s {
		n = n*10 + int(r-'0')
	}
	return n
}

func main() {
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "SA-MP API is running."})
	})

	app.Get("/api/server", func(c *fiber.Ctx) error {
		ip := c.Query("ip")
		port := atoi(c.Query("port"))
		if ip == "" || port == 0 {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid IP or port"})
		}

		data, err := queryServer(ip, port)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		offset := 11 // skip SAMP header

		hostname, offset, err := parseString(data, offset)
		if err != nil {
			hostname = "Unknown"
		}

		gamemode, offset, err := parseString(data, offset)
		if err != nil {
			gamemode = "Unknown"
		}

		mapname, offset, err := parseString(data, offset)
		if err != nil {
			mapname = "Unknown"
		}

		var players, maxPlayers uint16
		if offset+4 <= len(data) {
			players = binary.LittleEndian.Uint16(data[offset : offset+2])
			maxPlayers = binary.LittleEndian.Uint16(data[offset+2 : offset+4])
		}

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
