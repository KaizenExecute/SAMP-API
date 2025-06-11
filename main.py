from fastapi import FastAPI, HTTPException, Query
from fastapi.middleware.cors import CORSMiddleware
from samp_query import SampQuery

app = FastAPI(title="SA-MP/Open.MP Server Status API")

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

@app.get("/api/server")
def get_server_info(
    ip: str = Query(..., description="IP address of the SA-MP/Open.MP server"),
    port: int = Query(7777, description="Port of the server")
):
    try:
        with SampQuery(ip, port) as query:
            info = query.get_server_info()
            rules = query.get_rules()

            return {
                "hostname": info.hostname,
                "gamemode": info.gamemode,
                "mapname": info.mapname,
                "passworded": info.passworded,
                "players": info.players,
                "max_players": info.max_players,
                "rules": rules
            }

    except Exception as e:
        raise HTTPException(status_code=400, detail=f"Failed to query server: {e}")

@app.get("/api/players")
def get_player_list(
    ip: str = Query(..., description="IP address of the SA-MP/Open.MP server"),
    port: int = Query(7777, description="Port of the server")
):
    try:
        with SampQuery(ip, port) as query:
            players = query.get_players()

            return {
                "players": [
                    {"name": player.name, "score": player.score}
                    for player in players
                ]
            }

    except Exception as e:
        raise HTTPException(status_code=400, detail=f"Failed to query players: {e}")
