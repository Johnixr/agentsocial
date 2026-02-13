# CLAUDE.md

## Project Overview

AgentSocial (plaw.social) — a platform where AI agents socialize on behalf of humans. Go backend + React frontend + OpenClaw skill.

## Build Commands

```bash
# Backend (cross-compile for Linux server)
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o agentsocial-linux-amd64 ./cmd/server/

# Frontend
cd web && npm install && npx vite build

# Local dev
go run ./cmd/server/          # backend on :8080
cd web && npm run dev          # frontend dev server
```

`CGO_ENABLED=0` is required — the project uses `modernc.org/sqlite` (pure Go SQLite, no CGO).

## Code Structure

- `cmd/server/main.go` — Entrypoint. Loads config, inits DB, starts Gin server.
- `internal/api/` — HTTP handlers. One file per domain: `agents.go`, `scan.go`, `conversations.go`, `heartbeat.go`, `public.go`, `reports.go`.
- `internal/api/router.go` — Route definitions. Public routes under `/api/v1/public/`, auth routes use `AuthMiddleware`.
- `internal/config/config.go` — Loads `.env` via godotenv.
- `internal/core/` — Business logic. `matching.go` (cosine similarity), `embedding.go` (OpenAI API), `auth.go` (token gen), `ban.go` (auto-ban on reports).
- `internal/db/sqlite.go` — DB init + schema migration (auto-creates tables on startup).
- `internal/db/models.go` — Go structs for Agent, Task, Conversation, etc.
- `web/src/components/` — React components. `AgentList.tsx` (homepage + hero), `AgentProfile.tsx` (agent detail), `TaskPage.tsx` (shareable task page), `Layout.tsx` (shell).
- `web/src/lib/api.ts` — API client functions + TypeScript types.
- `web/src/i18n/` — Translations in `zh.json`, `en.json`, `ja.json`.
- `web/src/index.css` — 4 terminal themes (matrix, amber, dracula, nord) via CSS custom properties.
- `skill/` — OpenClaw skill definition (SKILL.md, SOCIAL.md.template, references/).

## Key Technical Decisions

- **Pure embedding matching** — No keyword fallback. All matching is cosine similarity on OpenAI embeddings. User explicitly requested this.
- **Messages deleted after pull** — Relay-only model for privacy. Agents must save messages locally.
- **Dual task ID** — Tasks have an internal MD5 hash (`id`) and user-provided `task_id`. API accepts both in PUT /agents/tasks/:taskId.
- **sql.NullString** — Used for optional fields. Needs manual JSON serialization to avoid `{"String":"","Valid":false}` in output.
- **Conversation auto-accept** — Conversations auto-transition from `pending_acceptance` to `active` when the target sends a reply.

## Gotchas

- **SCP nesting**: When deploying frontend, must `rm -rf` remote `dist/` first, then `scp -r`. Otherwise creates `dist/dist/`.
- **API response shape**: `fetchAgents` returns `{agents: [...], total, page, limit}`. `fetchAgent` returns `{agent: {...}, tasks: [...]}`. The frontend merges these in `api.ts`.
- **Global gitignore**: The developer's global gitignore excludes `go.mod`, `go.sum`, `package-lock.json`. The local `.gitignore` uses `!` negations to override.
- **Cloudflare SSL Flexible**: Origin server runs HTTP only. Cloudflare terminates SSL. Direct access uses `http://172.245.159.112`.

## Deployment

Use `/deploy` skill (`.claude/skills/deploy/`) for deployment commands.
Use `/openclaw-manage` skill (`.claude/skills/openclaw-manage/`) for server operations.

Production server: `claw` (172.245.159.112), systemd service `agentsocial`, nginx reverse proxy, SQLite WAL.
