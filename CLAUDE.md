# CLAUDE.md — Planetalk Backend

Guidance for working in this repo. Read this before making changes.

## What this is
Go (Gin) backend for **Planetalk**: 1v1 stranger video chat with real-time message
translation and social features. PostgreSQL + Redis, Agora for video. The Flutter
mobile app is a separate repo. Module/import path: `github.com/sebas1197-git/planetalk`.

## Working style with the user
- The user is learning. **Explain changes in simple, beginner-friendly terms.**
- From Step 4 onward the user asked Claude to **write the code directly** (not just
  scaffold). Still explain what was written and why.
- Build features **step by step**; after each, give a plain-language summary and
  concrete test steps. Commit per step only when the user asks.

## Architecture & conventions (follow these exactly)
- **Layered modules** under `internal/<module>/`: `handler → service → repository`.
  - `repository.go` — the ONLY place with SQL. Returns Go types + sentinel errors.
  - `service.go` — business rules; no HTTP, no SQL strings.
  - `handler.go` — reads input, calls service, writes response. Thin.
  - `routes.go` — `RegisterRoutes(r gin.IRouter, h *Handler, tokens *auth.TokenManager)`.
  - `models.go` — DTOs/structs returned by the API.
- **API versioning:** all app routes are registered on the `v1 := r.Group("/api/v1")`
  group in `cmd/api/main.go`. `GET /health` stays unversioned. Don't hardcode
  `/api/v1` inside modules — only `main.go` knows the prefix.
- **Responses (Approach A)** via `internal/httputil`:
  - Success: `httputil.OK(c, data)` / `Created` / `NoContent` — raw resource.
  - Lists: `httputil.List(c, items, httputil.Meta{Total,Page,PerPage})` → `{data, meta}`.
  - Errors: `httputil.Error(c, status, code, message)` → `{"error":{"code","message"}}`.
    `code` is a STABLE machine-readable string (the mobile app branches/translates on it).
  - In middleware use `httputil.Abort(...)` (calls `c.Abort()`).
- **Auth:** `middleware.RequireAuth(tokens)` validates the Bearer JWT and stores the
  user id at `middleware.ContextUserID`. Handlers read it via `c.GetString(middleware.ContextUserID)`.
- **Migrations:** numbered pairs in `migrations/NNNNNN_name.{up,down}.sql`. Applied on
  startup by `internal/db.RunMigrations`. uuid PKs use `gen_random_uuid()` (pgcrypto,
  enabled in migration 000001). Add a new migration; never edit an applied one.
- **External providers behind interfaces** with a dev stub + real impl, chosen by
  whether credentials are set: `pkg/twilio` (SMS), `pkg/translate` (translation),
  `pkg/agora` (video tokens). This keeps the app testable with zero credentials.
- **Realtime:** `internal/realtime.Hub.SendToUser(ctx, userID, realtime.Event{Type,Data})`
  publishes to Redis pub/sub; every instance delivers to its local WS clients. Other
  modules (match, chat) push live events through the hub.
- **Module dependencies:** `realtime` is a leaf. `match` depends on `realtime`. `call`
  and `chat` depend on `match` (reuse `match.Service.Get` for participant checks) and
  `realtime`. Keep this acyclic.

## Dev workflow
- **Hot reload:** `docker compose up -d` runs Postgres + Redis + the API via **Air**
  (see `docker-compose.override.yml` + `Dockerfile.dev`). Edit a `.go` file → Air
  rebuilds automatically. No `--build` needed for code changes.
- **Prod-like build:** `docker compose -f docker-compose.yml up --build` (no override).
- **Container names:** `planetalk-api`, `planetalk-db`, `planetalk-redis`.
- **Host ports (non-default to avoid clashes):** Postgres `5434`, Redis `6380`, API `8080`.
- **Local build check (Go not on PATH; use the GoLand SDK):**
  `& "C:\Users\USER\sdk\go1.26.4\bin\go.exe" build ./...`
- **After `go get` a new dependency:** the Air container's module-cache volume won't
  have it yet. Run `docker compose exec -T api sh -c "go mod download"` then
  `docker compose restart api`.
- **Logs / DB:** `docker compose logs -f api` · `docker exec -it planetalk-db psql -U planetalk -d planetalk`.

## Testing approach
- **Unit tests** (fast, no Docker): colocated `*_test.go`, run with `go test ./...`
  (or `make test`). Cover pure logic (jwt, config, agora, translate, refresh helpers).
- **Integration tests** (real Postgres): files tagged `//go:build integration`, run with
  `go test -tags=integration ./...` (or `make test-integration`). `internal/testutil`
  spins up a throwaway Postgres via **testcontainers-go** and applies the **embedded**
  migrations (`migrations/embed.go` → iofs), so they're CWD-independent. Repository
  tests live in `*_repository_integration_test.go`. Needs Docker running.
  - **Windows local gotcha:** if you see "rootless Docker is not supported on Windows",
    set `DOCKER_HOST=npipe:////./pipe/docker_engine` and run with `-p 1` (serial). CI
    (Linux) is unaffected. `make test-integration` already passes `-p 1`.
- **CI:** `.github/workflows/ci.yml` — a `unit` job (gofmt + vet + build + `go test -race`)
  and an `integration` job (`go test -tags=integration`, Docker provided by ubuntu-latest).
- **Dev OTP:** when `APP_ENV=development`, `POST /auth/otp/request` returns `dev_code`
  in the response (NEVER in production). Used to automate login.
- **Throwaway E2E checks:** for quick manual flow checks, write a Go program under
  `tmp/<name>/main.go` (`tmp/` is gitignored + Air-excluded), run it, then delete it.
- **Postman:** collection "PlaneTalk" (managed via the Postman MCP). Two environments,
  `PlaneTalk - User A` / `User B`, are self-contained (own tokens) for two-user flows.
  `{{base_url}}` = `{{host}}/api/v1`. Keep new endpoints' requests in sync there.

## Git
- Work on **`develop`**; `master` is the release branch. Repo: `sebas1197-git/planetalk-api` (public).
- Don't commit `.env` (gitignored). Verify before committing: `git status --short | grep '\.env$'`.
- PowerShell shows git's stderr in red even on success — check the final ref update line.

## Gotchas
- **Gin static-vs-param route conflict:** don't register `/x/:id` and `/x/static` under
  the same method+prefix. We use `/match/...` for queue actions and `/matches/:id/...`
  for the match resource to avoid this.
- **pgx scanning:** nullable columns scan into `*string` / `*time.Time`; uuid columns
  scan into `string` fine. For `date`, select `to_char(col,'YYYY-MM-DD')`.
- **Don't return transport info in the body** (no `status`/`code` mirroring the HTTP
  status). HTTP status is the source of truth; the body carries data or `{error}`.

## Build order / status
Skeleton → DB → Auth → Users/Friends → Realtime → Quick Match → Call (Agora) → Chat
(done). Next: Feed (posts), Moderation (reports/blocks), then the Flutter app.

## Known design gaps (intentional, revisit later)
- Chat is **match-scoped**; there are no friend-to-friend DMs yet. A future refactor
  may generalize to a `conversation` shared by matches and friendships.
- Refresh tokens are stateless (no rotation/revocation). OTP rows aren't purged.
