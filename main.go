package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	sampquery "github.com/markqiu/go-sampquery"
)

// ServerInfo is the JSON response struct
type ServerInfo struct {
	Hostname   string `json:"hostname"`
	Gamemode   string `json:"gamemode"`
	Mapname    string `json:"mapname"`
	Players    int    `json:"players"`
	MaxPlayers int    `json:"max_players"`
	Passworded bool   `json:"passworded"`
	Language   string `json:"language"`
}

// getServerInfo queries a SA-MP server and returns structured data
func getServerInfo(ip string) (*ServerInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	info, err := sampquery.GetServerInfo(ctx, ip, true)
	if err != nil {
		return nil, err
	}

	return &ServerInfo{
		Hostname:   info.Hostname,
		Gamemode:   info.Gamemode,
		Mapname:    info.Mapname,
		Players:    info.Players,
		MaxPlayers: info.MaxPlayers,
		Passworded: info.Password,
		Language:   info.Language,
	}, nil
}

func main() {
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("🎮 SA-MP Monitor API is running")
	})

	app.Get("/api/server", func(c *fiber.Ctx) error {
		ip := c.Query("ip")
		if ip == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "Missing ?ip=IP:PORT",
			})
		}

		info, err := getServerInfo(ip)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(info)
	})

	fmt.Println("🚀 API running at http://localhost:3000")
	app.Listen(":3000")
}
