# 🌐 SAMP-API

A simple, lightweight, and fast API to query SA-MP (San Andreas Multiplayer) or Open.MP servers and retrieve live server details and player data in JSON format.

## 🔧 Features

- Query any public SA-MP or Open.MP server
- Get server information (hostname, gamemode, version, players, etc.)
- REST API using Go (Golang)
- Supports IP:Port input format

## 📦 API Endpoints

### `GET https://ainsoft.xyz/api/server/{ip}:{port}`

Returns server info.

**Example:**

**Response:**
```json
{
  "ip": "127.0.0.0:7777",
  "hostname": "Grand Life Roleplay",
  "gamemode": "Los Santos",
  "version": "omp 1.4.0.2783",
  "players": 46,
  "max_players": 300,
  "passworded": false,
  "isOmp": true
}
```
---

❗ Notes

Works with both SA-MP and Open.MP servers

Make sure the server IP is correct and public

Port must be open and accessible for queries



---

🤝 License

MIT License © 2025 KaizenExecute
