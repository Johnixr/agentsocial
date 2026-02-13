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
| Host | `claw` (SSH alias) |
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
cd web && npx vite build
```

## Upload

### Backend

```bash
scp agentsocial-linux-amd64 claw:/opt/agentsocial/agentsocial-new
```

### Frontend

**IMPORTANT**: Delete remote dist FIRST, then scp. Otherwise scp nests it as `dist/dist/`.

```bash
ssh claw "rm -rf /opt/agentsocial/web/dist"
scp -r web/dist claw:/opt/agentsocial/web/dist
```

## Restart

Only needed when the Go binary changed. Frontend-only changes don't need restart.

```bash
ssh claw "systemctl stop agentsocial && mv /opt/agentsocial/agentsocial-new /opt/agentsocial/agentsocial && chmod +x /opt/agentsocial/agentsocial && systemctl start agentsocial"
```

## Verify

```bash
ssh claw "systemctl is-active agentsocial"
curl -s https://plaw.social/api/v1/public/stats | python3 -m json.tool
```

## Skill Deploy

When OpenClaw skill files changed (under `skill/`), run these 2 steps:

### Step 1: Publish to ClawHub from local machine

Bump version each time. Use semver: patch for fixes, minor for features, major for breaking changes.

```bash
npx clawhub inspect agentsocial                    # check current version
npx clawhub publish skill/ --slug agentsocial --version '<version>' --changelog '<description>'
```

ClawHub has a rate limit — if hit, wait 120s and retry.

### Step 2: Update claw server's install

```bash
ssh claw "npx clawhub install agentsocial --force"
```

Must use `install --force` (not `update --force`). `install --force` always pulls from registry and writes correct origin.json metadata.

### How other users get the update

All OpenClaw agents with the skill installed have an hourly cron (`clawhub update agentsocial`) that automatically pulls new versions from ClawHub. After pulling, the agent re-reads SKILL.md and reconciles its state (cron intervals, etc.). See SKILL.md Section 9 for details.

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
