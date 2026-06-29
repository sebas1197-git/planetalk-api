# Planetalk — Backend API

Backend for **Planetalk**, a mobile app for private 1‑on‑1 video chat with people
around the world, featuring real‑time message translation and lightweight social
features (profiles, interests, friends).

This repository contains the **Go API**. The Flutter mobile app lives in a
separate repository.

> ⚠️ Work in progress — built step by step. See **Roadmap** below for status.

## Tech stack
- **Go + Gin** — HTTP API
- **PostgreSQL** — primary data store (users, matches, messages, posts…)
- **Redis** — matchmaking queue, presence, and WebSocket pub/sub
- **Agora** — real‑time video (server mints short‑lived tokens; the App
  Certificate never leaves the backend)
- **Twilio** — OTP SMS · **Google Translate** — message translation
- **JWT** — stateless auth (access + refresh tokens)
- **Docker Compose** — local Postgres/Redis/API · **Air** — hot reload in dev
- **golang-migrate** — versioned SQL migrations

## API design
- Versioned under **`/api/v1`** (e.g. `POST /api/v1/auth/otp/verify`).
  `GET /health` stays unversioned.
- **Success** responses return the resource directly; **lists** use
  `{ "data": [...], "meta": {...} }`.
- **Errors** use a consistent envelope with a stable, machine‑readable code:
  ```json
  { "error": { "code": "invalid_code", "message": "the code is incorrect" } }
  ```

## Project layout
```
cmd/api/          # entrypoint (wires everything together)
internal/
  config/         # env -> Config struct
  db/             # Postgres pool + migration runner
  redis/          # Redis client
  httputil/       # consistent success/error response helpers
  middleware/     # JWT auth middleware
  auth/           # phone OTP + JWT
  user/           # profiles, interests, friends
  realtime/       # WebSocket hub (presence + Redis pub/sub fan-out)
  match/          # Quick Match (Redis queue + pairing)
  call/           # Agora video token issuing
  chat/           # messages + translation
pkg/
  twilio/         # SMS sender (console stub in dev, Twilio in prod)
  agora/          # RTC token builder
  translate/      # translator (Google v2 REST, stub in dev)
migrations/       # SQL migrations
```
Each feature module follows the same layers: **handler → service → repository**.

## Getting started

### Prerequisites
- Go 1.25+
- Docker Desktop

### Run everything (with hot reload)
```bash
cp .env.example .env            # fill values; dev defaults work out of the box
docker compose up -d            # starts Postgres, Redis, and the API (Air)
curl http://localhost:8080/health   # -> {"status":"ok"}
```
Edit any `.go` file and Air rebuilds the API automatically — no `--build` needed.

> Datastores run on non‑default host ports to avoid clashes: **Postgres `5434`**,
> **Redis `6380`**.

### Production-like build (no hot reload)
```bash
docker compose -f docker-compose.yml up --build
```

### Run the API in your IDE
Keep `postgres` + `redis` containers up and run `cmd/api` directly; it connects
via `localhost:5434` / `localhost:6380` from `.env`.

## Auth flow (phone OTP)
1. `POST /api/v1/auth/otp/request` `{ "phone": "+1..." }` — sends a 6‑digit code
   (printed to logs in dev via the console SMS stub).
2. `POST /api/v1/auth/otp/verify` `{ "phone", "code" }` — returns
   `access_token` + `refresh_token`; creates the user on first login.
3. Send `Authorization: Bearer <access_token>` on protected routes.
4. `POST /api/v1/auth/refresh` — exchange a refresh token for a new access token.

## Endpoints (so far)
| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/health` | – | Liveness probe |
| POST | `/api/v1/auth/otp/request` | – | Send an OTP code |
| POST | `/api/v1/auth/otp/verify` | – | Verify code → tokens |
| POST | `/api/v1/auth/refresh` | – | Refresh access token |
| GET | `/api/v1/interests` | – | List available interests |
| GET | `/api/v1/me` | ✅ | Current user's profile |
| PATCH | `/api/v1/me` | ✅ | Update profile |
| PUT | `/api/v1/me/interests` | ✅ | Set your interests |
| GET | `/api/v1/users/:id` | ✅ | Another user's profile |
| GET | `/api/v1/friends` | ✅ | List friends |
| GET | `/api/v1/friends/requests` | ✅ | Incoming requests |
| POST | `/api/v1/friends/:id` | ✅ | Send friend request |
| POST | `/api/v1/friends/:id/accept` | ✅ | Accept request |
| DELETE | `/api/v1/friends/:id` | ✅ | Remove friend |
| GET (WS) | `/api/v1/ws?token=…` | ✅ | WebSocket: live events |
| GET | `/api/v1/presence` | ✅ | List online users |
| GET | `/api/v1/presence/:id` | ✅ | Is a user online? |
| POST | `/api/v1/realtime/echo` | ✅ | Send a test event to yourself |
| POST | `/api/v1/match/enter` | ✅ | Join Quick Match queue |
| POST | `/api/v1/match/leave` | ✅ | Leave the queue |
| GET | `/api/v1/matches/:id` | ✅ | Match details (participants) |
| POST | `/api/v1/matches/:id/end` | ✅ | End a match |
| POST | `/api/v1/matches/:id/token` | ✅ | Get an Agora video token |
| POST | `/api/v1/matches/:id/messages` | ✅ | Send a chat message |
| GET | `/api/v1/matches/:id/messages` | ✅ | Chat history (paginated) |

### Realtime events (over the WebSocket)
| Event `type` | When |
|---|---|
| `match_found` | You were paired in Quick Match |
| `match_ended` | The other participant ended the match |
| `message` | You received a chat message |
| `echo` | Test event (from `/realtime/echo`) |

## Roadmap
- [x] Project skeleton & config
- [x] Database layer & migrations
- [x] Auth (phone OTP + JWT)
- [x] Users, profiles, interests, friends
- [x] Realtime hub (WebSockets: presence + live events)
- [x] Quick Match (Redis matchmaking queue)
- [x] Video call tokens (Agora)
- [x] Chat + real‑time translation
- [ ] Feed (posts / personal homepage)
- [ ] Moderation (reports, blocks)
- [ ] Flutter mobile app (separate repo)

## License
Proprietary — all rights reserved (for now).
