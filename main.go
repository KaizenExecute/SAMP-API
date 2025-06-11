package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
)

// ServerCore holds the basic server info
type ServerCore struct {
	Address    string `json:"address"`
	Hostname   string `json:"hostname"`
	Players    int    `json:"players"`
	MaxPlayers int    `json:"max_players"`
	Gamemode   string `json:"gamemode"`
	Language   string `json:"language"`
	Password   bool   `json:"password"`
}

// Server wraps core info, rules and extra fields
type Server struct {
	Core        ServerCore        `json:"core"`
	Rules       map[string]string `json:"rules"`
	Description string            `json:"description,omitempty"`
	Banner      string            `json:"banner,omitempty"`
	Active      bool              `json:"active,omitempty"`
}

// In-memory store
var (
	servers = make(map[string]Server)
	mutex   = &sync.Mutex{}
)

func main() {
	app := fiber.New()

	app.Post("/v2/server", postServer)
	app.Patch("/v2/server", patchServer)
	app.Get("/v2/server/:address", getServer)

	log.Fatal(app.Listen(":8080"))
}

func postServer(c *fiber.Ctx) error {
	var srv Server
	if err := c.BodyParser(&srv); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid server format"})
	}

	addr := normalizeAddress(srv.Core.Address)
	mutex.Lock()
	servers[addr] = srv
	mutex.Unlock()

	return c.JSON(srv)
}

func patchServer(c *fiber.Ctx) error {
	address := c.FormValue("address")
	if address == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "missing address field"})
	}

	normalized := normalizeAddress(address)
	mutex.Lock()
	srv, ok := servers[normalized]
	if !ok {
		mutex.Unlock()
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "server not found"})
	}
	srv.Active = true
	servers[normalized] = srv
	mutex.Unlock()

	return c.JSON(fiber.Map{"status": "updated"})
}

func getServer(c *fiber.Ctx) error {
	address := c.Params("address")
	normalized := normalizeAddress(address)

	mutex.Lock()
	srv, ok := servers[normalized]
	mutex.Unlock()

	if !ok {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "server not found"})
	}

	return c.JSON(srv)
}

func normalizeAddress(address string) string {
	return strings.ToLower(strings.TrimSpace(address))
}
