from fastapi import FastAPI, Query, HTTPException
from samp_client.client import SampClient
from samp_client.exceptions import SampError
from pydantic import BaseModel
from typing import List

app = FastAPI(title="SA-MP/Open.MP Server Info API")

class PlayerInfo(BaseModel):
    id: int
    name: str
    score: int
    ping: int

class ServerInfo(BaseModel):
    ip: str
    port: int
    hostname: str
    gamemode: str
    language: str
    players: int
    max_players: int
    player_list: List[PlayerInfo]

@app.get("/serverinfo", response_model=ServerInfo)
def get_server_info(ip: str = Query(...), port: int = Query(...)):
    try:
        with SampClient(address=ip, port=port, timeout=2) as client:
            info = client.get_server_info()
            players = client.get_players()

            player_list = [
                PlayerInfo(id=p.id, name=p.name, score=p.score, ping=p.ping)
                for p in players
            ]

            return ServerInfo(
                ip=ip,
                port=port,
                hostname=info.hostname,
                gamemode=info.gamemode,
                language=info.language,
                players=info.players,
                max_players=info.max_players,
                player_list=player_list
            )
    except SampError:
        raise HTTPException(status_code=504, detail="Server not responding or timed out")
