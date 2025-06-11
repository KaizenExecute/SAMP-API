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


@app.get("/api/server")
def get_server_info(ip: str = Query(...), port: int = Query(7777)):
    try:
        data = samp_query(ip, port, b'i')
        offset = 11
        hostname = gamemode = mapname = "Unknown"
        players = max_players = 0

        try:
            hostname_len = data[offset]
            offset += 1
            hostname = data[offset:offset + hostname_len].decode(errors='ignore')
            offset += hostname_len
        except: pass

        try:
            gamemode_len = data[offset]
            offset += 1
            gamemode = data[offset:offset + gamemode_len].decode(errors='ignore')
            offset += gamemode_len
        except: pass

        try:
            mapname_len = data[offset]
            offset += 1
            mapname = data[offset:offset + mapname_len].decode(errors='ignore')
            offset += mapname_len
        except: pass

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
                name_len = data[offset]
                offset += 1
                name = data[offset:offset + name_len].decode(errors='ignore')
                offset += name_len
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
