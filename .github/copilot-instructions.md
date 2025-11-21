# Arma-Reforger API - AI Agent Instructions

## Architecture Overview
This is a Go-based proxy/gateway service for Arma Reforger game servers. It authenticates Steam users locally and forwards requests to multiple external game APIs.

**Key Components:**
- `handlers/` - HTTP endpoints (auth, servers, workshop)
- `models/` - Data structures and SQLite operations
- `utils/` - Cross-cutting concerns (API forwarding, helpers)

**Data Flow:**
1. Steam auth via `handlers/auth.go` → local user/token stored
2. Remote API auth via `utils/integration.go` → tokens cached per API
3. Room join requests → token translation → forwarded to target API

## Critical Patterns

### Token Management
- Local tokens: stored in `users.access_token`
- Remote tokens: stored in `user_tokens` table with `api_name`
- Always use `models.CreateOrUpdateAccessToken(userID, token, apiName)` for remote tokens
- Use `models.GetAccessTokenByUserAndAPI(userID, apiName)` to retrieve

### API Forwarding
- `utils/integration.go` contains all cross-API logic
- `JoinOnOtherAPI()` automatically translates local tokens to remote tokens
- `AuthOnOtherAPI()` caches remote tokens per user per API

### Database Schema
- SQLite with foreign key constraints enabled
- Users → user_tokens (1:many)
- Servers with `api_name` field for routing
- Use existing `models/` functions, don't write raw SQL

## Development Workflow
```bash
# Build
go build ./...

# Run with config.json in project root (listens on port from config.API.PORT)
./arma-reforger-api

# Test endpoints (examples)
curl -X POST http://localhost:8080/game-identity/api/v1.1/identities/reforger/auth -H "Content-Type: application/json" -d '{"platform":"Arma Reforger PC","token":"steam_token","platformOpts":{"appId":"480"}}'
```

## API Endpoints Reference

### Authentication
- `POST /game-identity/api/v1.1/identities/reforger/auth` - Steam authentication and user creation
- `GET /game-identity/api/v1.0/health` - Health check

### Game API - S2S (Server-to-Server)
- `POST /game-api/s2s-api/v1.0/lobby/rooms/acceptPlayer` - Accept player on server
- `POST /game-api/s2s-api/v1.0/lobby/dedicatedServers/registerUnmanagedServer` - Register unmanaged server
- `POST /game-api/s2s-api/v1.0/lobby/rooms/register` - Register room
- `POST /game-api/s2s-api/v1.0/lobby/rooms/remove` - Remove server
- `POST /game-api/s2s-api/v1.0/lobby/dedicatedServers/heartBeat` - Server heartbeat
- `POST /game-api/s2s-api/v1.0/lobby/rooms/listActiveBans` - List active bans
- `POST /game-api/s2s-api/v1.0/lobby/rooms/removePlayer` - Remove player
- `POST /game-api/s2s-api/v1.0/lobby/rooms/createBan` - Create ban
- `POST /game-api/s2s-api/v1.0/lobby/rooms/removeBans` - Remove bans
- `POST /game-api/s2s-api/v1.0/sendTdEvents` - Send telemetry events

### Game API - Public
- `POST /game-api/api/v1.0/lobby/rooms/search` - Search available rooms/servers
- `POST /game-api/api/v1.0/lobby/rooms/getRoomsByIds` - Get rooms by IDs
- `POST /game-api/api/v1.0/lobby/rooms/join` - Join room (main join endpoint)
- `POST /game-api/api/v1.0/lobby/rooms/verifyPassword` - Verify room password
- `POST /game-api/api/v1.0/lobby/rooms/register` - Register room
- `POST /game-api/api/v1.0/lobby/rooms/heartBeat` - Room heartbeat
- `POST /game-api/api/v1.0/lobby/rooms/remove` - Remove room
- `POST /game-api/api/v1.0/lobby/rooms/acceptPlayer` - Accept player
- `POST /game-api/api/v1.0/lobby/rooms/listPlayers` - List room players
- `POST /game-api/api/v1.0/lobby/dedicatedServers/createOwnerToken` - Create owner token
- `POST /game-api/api/v1.0/lobby/getPingSites` - Get ping sites
- `POST /game-api/api/v1.0/session/login` - Session login
- `POST /game-api/api/v1.0/blockList/listBlocked` - List blocked users
- `POST /game-api/api/v1.0/sendTdEvents` - Send telemetry events
- `GET /game-api/api/v1.0/world` - Get world info
- `GET /game-api/api/v1.0/dummy` - Dummy endpoint
- `GET /game-api/health` - Health check

### Workshop API
- `POST /workshop-api/api/v3.0/assets/list` - List workshop assets

### Utility Endpoints
- `POST /ping` - Ping endpoint
- `GET /api/news` - Get news
- `POST /GetWorkshopToken` - Get workshop token
- `POST /getLobby` - Get lobby info

## Key Files for Modifications
- `utils/integration.go` - Cross-API communication logic
- `models/database.go` - All database operations
- `handlers/auth.go` - Authentication flow
- `config.json` - API endpoints configuration

## Important Conventions
- Use `log.Printf()` for async operations (AuthOnOtherAPI)
- Always close HTTP response bodies with `defer resp.Body.Close()`
- Room join requests automatically get token translation
- API URLs configured in `config.json.ServersAPI`
- Return errors in HTTP responses using `c.JSON(http.StatusInternalServerError, gin.H{"error": "message"})` or appropriate status codes
- Use `fmt.Errorf()` for error wrapping to preserve context