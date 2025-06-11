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


def samp_query(ip, port, opcode):
    try:
        prefix = b'SAMP' + bytes(map(int, ip.split('.'))) + struct.pack('<H', port)
        packet = prefix + opcode
        with socket.socket(socket.AF_INET, socket.SOCK_DGRAM) as s:
            s.settimeout(2)
            s.sendto(packet, (ip, port))
            data, _ = s.recvfrom(4096)
            if not data or len(data) < 11:
                raise ValueError("Empty or invalid response from server")
            return data
    except socket.timeout:
        raise ValueError("No response from server (timeout)")
    except Exception as e:
        raise ValueError(f"UDP query failed: {e}")


@app.get("/api/server")
def get_server_info(ip: str = Query(...), port: int = Query(7777)):
    try:
        data = samp_query(ip, port, b'i')
        offset = 11

        if len(data) < offset + 1:
            raise ValueError("Incomplete server info response")

        hostname_len = data[offset]
        offset += 1
        if len(data) < offset + hostname_len + 1:
            raise ValueError("Hostname length exceeds buffer")

        hostname = data[offset:offset + hostname_len].decode('utf-8', errors='ignore')
        offset += hostname_len

        gamemode_len = data[offset]
        offset += 1
        if len(data) < offset + gamemode_len + 1:
            raise ValueError("Gamemode length exceeds buffer")

        gamemode = data[offset:offset + gamemode_len].decode('utf-8', errors='ignore')
        offset += gamemode_len

        mapname_len = data[offset]
        offset += 1
        if len(data) < offset + mapname_len + 4:
            raise ValueError("Mapname length exceeds buffer")

        mapname = data[offset:offset + mapname_len].decode('utf-8', errors='ignore')
        offset += mapname_len

        if len(data) < offset + 4:
            raise ValueError("Missing player data")

        players, max_players = struct.unpack_from('<HH', data[offset:])
        return {
            "hostname": hostname,
            "gamemode": gamemode,
            "mapname": mapname,
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
        if len(data) < offset + 1:
            raise ValueError("Invalid player response")

        player_count = data[offset]
        offset += 1
        players = []

        for _ in range(player_count):
            if offset >= len(data):
                break
            name_len = data[offset]
            offset += 1
            if len(data) < offset + name_len + 4:
                break
            name = data[offset:offset + name_len].decode('utf-8', errors='ignore')
            offset += name_len
            score = struct.unpack_from('<I', data[offset:offset + 4])[0]
            offset += 4
            players.append({"name": name, "score": score})

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
