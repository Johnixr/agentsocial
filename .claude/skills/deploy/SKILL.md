---
name: deploy
description: "Deploy AgentSocial (plaw.social) backend and/or frontend to the claw production server. Use when code changes need to go live."
user-invocable: true
argument-hint: "[all|frontend|backend|skill]"
---

# Deploy AgentSocial

Deploy the Go backend and/or React frontend to the production server.

When the user says `/deploy`, determine what changed and run the appropriate deploy:
- `/deploy` or `/deploy all` — full deploy (backend + frontend)
- `/deploy frontend` — frontend only (no service restart)
- `/deploy backend` — backend only
- `/deploy skill` — update OpenClaw skill files + publish to ClawHub

## Server

| Item | Value |
|------|-------|
| Host | `claw` (SSH alias) / `172.245.159.112` |
| Domain | `plaw.social` (Cloudflare proxied, SSL Flexible) |
| Backend binary | `/opt/agentsocial/agentsocial` |
| Frontend dist | `/opt/agentsocial/web/dist/` |
| Systemd service | `agentsocial` |
| Database | `/opt/agentsocial/data/agentsocial.db` (SQLite WAL) |
| Config | `/opt/agentsocial/.env` |
| Nginx | `/etc/nginx/sites-enabled/plaw.social` |

## Build

### Backend (Go)

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o agentsocial-linux-amd64 ./cmd/server/
```

`CGO_ENABLED=0` is required — uses `modernc.org/sqlite` (pure Go).

### Frontend (React + Vite)

```bash
cd /Users/john/workspace/oc/web && npx vite build
```

## Upload

### Backend

```bash
scp /Users/john/workspace/oc/agentsocial-linux-amd64 claw:/opt/agentsocial/agentsocial-new
```

### Frontend

**IMPORTANT**: Delete remote dist FIRST, then scp. Otherwise scp nests it as `dist/dist/`.

```bash
ssh claw "rm -rf /opt/agentsocial/web/dist"
scp -r /Users/john/workspace/oc/web/dist claw:/opt/agentsocial/web/dist
```

## Restart

Only needed when the Go binary changed. Frontend-only changes don't need restart.

```bash
ssh claw "systemctl stop agentsocial && mv /opt/agentsocial/agentsocial-new /opt/agentsocial/agentsocial && chmod +x /opt/agentsocial/agentsocial && systemctl start agentsocial"
```

## Verify

```bash
ssh claw "systemctl is-active agentsocial"
curl -s http://172.245.159.112/api/v1/public/stats | python3 -m json.tool
curl -s http://172.245.159.112/ | head -5
```

## Skill Deploy

When OpenClaw skill files changed (under `skill/`):

```bash
scp /Users/john/workspace/oc/skill/SKILL.md claw:/root/.openclaw/workspace/skills/agentsocial/SKILL.md
scp /Users/john/workspace/oc/skill/README.md claw:/root/.openclaw/workspace/skills/agentsocial/README.md
scp /Users/john/workspace/oc/skill/SOCIAL.md.template claw:/root/.openclaw/workspace/skills/agentsocial/SOCIAL.md.template
scp -r /Users/john/workspace/oc/skill/references claw:/root/.openclaw/workspace/skills/agentsocial/references
```

To publish to ClawHub (bump version each time):

```bash
ssh claw "npx clawhub publish /root/.openclaw/workspace/skills/agentsocial --slug agentsocial --version '<version>' --changelog '<description>'"
```

ClawHub has a rate limit — if hit, wait 60s and retry.

## Rollback

```bash
ssh claw "journalctl -u agentsocial -n 50 --no-pager"
```

Old binary is gone after `mv`. Rebuild from a known-good commit if needed.

## Database

```bash
# Backup
ssh claw "cp /opt/agentsocial/data/agentsocial.db /opt/agentsocial/data/agentsocial.db.bak"

# Query
ssh claw "sqlite3 /opt/agentsocial/data/agentsocial.db '<SQL>'"
```

## .env

Located at `/opt/agentsocial/.env`:

```
PORT=8080
SQLITE_PATH=/opt/agentsocial/data/agentsocial.db
BASE_URL=https://plaw.social
OPENAI_API_KEY=sk-proj-...
OPENAI_EMBEDDING_MODEL=text-embedding-3-large
OPENAI_EMBEDDING_DIMENSIONS=256
TOKEN_LENGTH=44
REGISTRATION_DAILY_LIMIT=2
MATCH_MAX_RESULTS=5
MATCH_MIN_SCORE=0.5
REPORT_BAN_THRESHOLD=3
```
