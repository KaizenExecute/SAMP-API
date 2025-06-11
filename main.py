import socket
import struct
from fastapi import FastAPI, HTTPException, Query
from fastapi.middleware.cors import CORSMiddleware

app = FastAPI(title="SA-MP/Open.MP API")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

@app.get("/")
def root():
    return {"message": "SA-MP/Open.MP API is running."}


def samp_query(ip: str, port: int, opcode: bytes) -> bytes:
    packet = b'SAMP' + bytes(map(int, ip.split('.'))) + struct.pack('<H', port) + opcode
    with socket.socket(socket.AF_INET, socket.SOCK_DGRAM) as s:
        s.settimeout(2)
        try:
            s.sendto(packet, (ip, port))
            data, _ = s.recvfrom(4096)
            if len(data) < 11:
                raise ValueError("Invalid packet length.")
            return data
        except socket.timeout:
            raise ValueError("No response from server (timeout).")
        except Exception as e:
            raise ValueError(f"Socket error: {str(e)}")


def parse_string(data: bytes, offset: int) -> tuple[str, int]:
    if offset >= len(data):
        return "", offset
    length = data[offset]
    offset += 1
    if offset + length > len(data):
        return "", offset + length
    raw = data[offset:offset + length]
    text = raw.decode('utf-8', errors='ignore')
    clean = ''.join(c for c in text if 32 <= ord(c) <= 126)
    return clean.strip(), offset + length


@app.get("/api/server")
def get_server_info(ip: str = Query(...), port: int = Query(7777)):
    try:
        data = samp_query(ip, port, b'i')
        offset = 11

        hostname, offset = parse_string(data, offset)
        gamemode, offset = parse_string(data, offset)
        mapname, offset = parse_string(data, offset)

        players = max_players = 0
        if offset + 4 <= len(data):
            players, max_players = struct.unpack_from('<HH', data, offset)

        if not (0 <= players <= max_players <= 5000):
            players = max_players = 0

        return {
            "hostname": hostname or "Unknown",
            "gamemode": gamemode or "Unknown",
            "mapname": mapname or "Unknown",
            "players": players,
            "max_players": max_players
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Query failed: {str(e)}")


@app.get("/api/players")
def get_players(ip: str = Query(...), port: int = Query(7777)):
    try:
        data = samp_query(ip, port, b'd')
        offset = 11
        players = []

        if offset >= len(data):
            raise ValueError("No player data.")

        count = data[offset]
        offset += 1

        for _ in range(count):
            name, offset = parse_string(data, offset)
            if offset + 4 > len(data):
                break
            score = struct.unpack_from('<I', data[offset:offset + 4])[0]
            offset += 4
            players.append({"name": name, "score": score})

        return {"players": players}
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Query failed: {str(e)}")


@app.get("/api/status")
def get_status(ip: str = Query(...), port: int = Query(7777)):
    try:
        _ = samp_query(ip, port, b'i')
        return {"online": True}
    except:
        return {"online": False}
