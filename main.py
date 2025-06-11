
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


def samp_query(ip, port, opcode, attempts=3):
    prefix = b'SAMP' + bytes(map(int, ip.split('.'))) + struct.pack('<H', port)
    packet = prefix + opcode
    last_error = None
    for _ in range(attempts):
        try:
            with socket.socket(socket.AF_INET, socket.SOCK_DGRAM) as s:
                s.settimeout(3)
                s.sendto(packet, (ip, port))
                data, _ = s.recvfrom(4096)
                if data and len(data) >= 11:
                    return data
        except Exception as e:
            last_error = e
    raise ValueError(f"No response from server (timeout or blocked) — {last_error}")


def safe_read_string(data: bytes, offset: int) -> tuple[str, int]:
    if offset >= len(data):
        return "", offset
    try:
        length = data[offset]
        offset += 1
        if offset + length > len(data):
            return "", offset
        raw = data[offset:offset + length]
        text = raw.decode('utf-8', errors='ignore')
        text = ''.join(c for c in text if 32 <= ord(c) <= 126)
        return text.strip(), offset + length
    except:
        return "", offset

@app.get("/api/server")
def get_server_info(ip: str = Query(...), port: int = Query(7777)):
    try:
        data = samp_query(ip, port, b'i')
        offset = 11

        def safe_read(data: bytes, offset: int) -> tuple[str, int]:
            if offset >= len(data):
                return "", offset
            try:
                length = data[offset]
                offset += 1
                if length == 0 or offset + length > len(data):
                    return "", offset
                raw = data[offset:offset + length]
                decoded = raw.decode('utf-8', errors='ignore')
                clean = ''.join(c for c in decoded if 32 <= ord(c) <= 126)
                return clean.strip(), offset + length
            except:
                return "", offset

        hostname, offset = safe_read(data, offset)
        gamemode, offset = safe_read(data, offset)
        mapname, offset = safe_read(data, offset)

        players = max_players = 0
        if offset + 4 <= len(data):
            try:
                players, max_players = struct.unpack_from('<HH', data, offset)
            except:
                players = max_players = 0

        # Sanity check
        if not (0 <= players <= 5000) or not (0 <= max_players <= 5000):
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

        try:
            count = data[offset]
            offset += 1
            for _ in range(count):
                name, offset = safe_read_string(data, offset)
                if offset + 4 > len(data):
                    break
                score = struct.unpack_from('<I', data[offset:offset + 4])[0]
                offset += 4
                players.append({"name": name, "score": score})
        except:
            pass

        return {"players": players}
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Query failed: {str(e)}")


@app.get("/api/status")
def get_status(ip: str, port: int = 7777):
    try:
        _ = samp_query(ip, port, b'i')
        return {"online": True}
    except:
        return {"online": False}
