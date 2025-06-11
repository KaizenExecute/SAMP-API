import socket
from fastapi import FastAPI, HTTPException, Query
from fastapi.middleware.cors import CORSMiddleware
import struct

app = FastAPI(title="SA-MP/Open.MP API")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Core SA-MP Query
def samp_query(ip, port, opcode):
    prefix = b'SAMP' + bytes(map(int, ip.split('.'))) + struct.pack('<H', port)
    packet = prefix + opcode
    with socket.socket(socket.AF_INET, socket.SOCK_DGRAM) as s:
        s.settimeout(2)
        s.sendto(packet, (ip, port))
        return s.recvfrom(4096)[0]

# /api/server endpoint
@app.get("/api/server")
def get_server_info(ip: str = Query(...), port: int = Query(7777)):
    try:
        data = samp_query(ip, port, b'i')
        offset = 11
        hostname_len = data[offset]
        offset += 1
        hostname = data[offset:offset + hostname_len].decode('utf-8', errors='ignore')
        offset += hostname_len
        gamemode_len = data[offset]
        offset += 1
        gamemode = data[offset:offset + gamemode_len].decode('utf-8', errors='ignore')
        offset += gamemode_len
        mapname_len = data[offset]
        offset += 1
        mapname = data[offset:offset + mapname_len].decode('utf-8', errors='ignore')
        offset += mapname_len
        players, max_players = struct.unpack_from('<HH', data[offset:])
        return {
            "hostname": hostname,
            "gamemode": gamemode,
            "mapname": mapname,
            "players": players,
            "max_players": max_players
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Query failed: {e}")

# /api/players endpoint
@app.get("/api/players")
def get_players(ip: str = Query(...), port: int = Query(7777)):
    try:
        data = samp_query(ip, port, b'd')
        offset = 11
        player_count = data[offset]
        offset += 1
        players = []
        for _ in range(player_count):
            name_len = data[offset]
            offset += 1
            name = data[offset:offset + name_len].decode('utf-8', errors='ignore')
            offset += name_len
            score = struct.unpack_from('<I', data[offset:offset + 4])[0]
            offset += 4
            players.append({"name": name, "score": score})
        return {"players": players}
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Query failed: {e}")
