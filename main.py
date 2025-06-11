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


def safe_read_string(data, offset):
    if offset >= len(data):
        return "", offset
    try:
        length = data[offset]
        offset += 1
        if length == 0 or offset + length > len(data):
            return "", offset
        string = data[offset:offset + length].decode('utf-8', errors='ignore')
        string = ''.join(c for c in string if 32 <= ord(c) <= 126)
        offset += length
        return string.strip(), offset
    except:
        return "", offset


@app.get("/api/server")
def get_server_info(ip: str = Query(...), port: int = Query(7777)):
    try:
        data = samp_query(ip, port, b'i')
        offset = 11

        hostname, offset = safe_read_string(data, offset)
        gamemode, offset = safe_read_string(data, offset)
        mapname, offset = safe_read_string(data, offset)

        players = max_players = 0
        try:
            players, max_players = struct.unpack_from('<HH', data[offset:])
        except: pass

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
