# Devert Monitor Agent

> Lightweight, event-driven Docker & server monitoring agent written in Go.

## Features

- 🐳 **Real-time Docker Events** — streams Docker Events API (no polling)
- 📦 **Container Monitoring** — full metadata, status, health, ports, volumes, networks
- 📊 **Resource Monitoring** — CPU%, memory, network I/O, disk I/O per container
- 🖼️ **Image Monitoring** — repository, tag, size, dangling images, usage
- 🌐 **Network & Volume Monitoring** — full Docker network/volume inventory
- 🖥️ **Server Monitoring** — CPU, memory, disk, network, OS info
- ⚙️ **Service Monitoring** — systemctl status for Nginx, MySQL, Redis, etc.
- 📨 **n8n Webhook** — async forwarding with HMAC-SHA256 signing + retry
- 🗄️ **PostgreSQL Logging** — all events persisted with full payload
- 🔐 **Security** — Bearer token auth, per-IP rate limiting, CORS
- 🚀 **< 20 MB RAM** target, < 1% CPU under normal load

---

## Deployment via GitHub Package (Direkomendasikan)

Image dibangun otomatis oleh GitHub Actions dan disimpan di **GitHub Container Registry (GHCR)**.
Di server, kamu hanya perlu dua file: `.env` dan `docker-compose.prod.yml`.

### Setup Awal di Server

```bash
# 1. Buat direktori
mkdir -p /opt/devert && cd /opt/devert

# 2. Download docker-compose.prod.yml saja
curl -O https://raw.githubusercontent.com/YOUR_USERNAME/YOUR_REPO/main/docker-compose.prod.yml

# 3. Buat .env dari template
curl -O https://raw.githubusercontent.com/YOUR_USERNAME/YOUR_REPO/main/.env.example
cp .env.example .env
nano .env  # Isi API_TOKEN, POSTGRES_DSN, N8N_WEBHOOK_URL, GITHUB_USERNAME, dll

# 4. Login ke GHCR (hanya sekali)
echo YOUR_GITHUB_TOKEN | docker login ghcr.io -u YOUR_USERNAME --password-stdin

# 5. Jalankan
docker compose -f docker-compose.prod.yml up -d
```

### Update ke Versi Terbaru

```bash
cd /opt/devert
docker compose -f docker-compose.prod.yml pull
docker compose -f docker-compose.prod.yml up -d
```

### Pinned ke Versi Spesifik

Edit `.env`, ubah `IMAGE_TAG`:

```bash
# Gunakan tag semantik untuk production
IMAGE_TAG=v1.2.3

# Atau gunakan latest dari main branch
IMAGE_TAG=latest

# Atau pin ke commit tertentu
IMAGE_TAG=main-a1b2c3d
```

---

## Quick Start (Development / Build Lokal)

```bash
cp .env.example .env && nano .env
docker compose up -d
```

## REST API

All endpoints require `Authorization: Bearer <API_TOKEN>` header.

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check (unauthenticated) |
| GET | `/api/server` | Server info (CPU, memory, disk, network, OS) |
| GET | `/api/containers` | All containers |
| GET | `/api/container/{id}` | Single container details |
| GET | `/api/images` | All Docker images |
| GET | `/api/networks` | All Docker networks |
| GET | `/api/volumes` | All Docker volumes |
| GET | `/api/services` | System service statuses |
| GET | `/api/resources` | Container resource usage |
| GET | `/api/events` | Event log (filterable) |
| POST | `/api/container/{id}/start` | Start container |
| POST | `/api/container/{id}/stop` | Stop container |
| POST | `/api/container/{id}/restart` | Restart container |
| POST | `/api/container/{id}/remove` | Remove container |

### Event Log Filters

```
GET /api/events?server=server-a&container=nginx&action=stop&limit=50&offset=0
```

---

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_NAME` | `devert-server` | Human-readable server name |
| `TIMEZONE` | `UTC` | Server timezone (e.g. `Asia/Jakarta`) |
| `HTTP_PORT` | `8080` | API server port |
| `API_TOKEN` | _(required)_ | Bearer token for API auth |
| `RATE_LIMIT_RPM` | `100` | Max requests/minute per IP |
| `POSTGRES_DSN` | _(required)_ | Full PostgreSQL connection string |
| `DOCKER_SOCKET` | `/var/run/docker.sock` | Docker socket path |
| `N8N_WEBHOOK_URL` | — | n8n webhook URL |
| `N8N_WEBHOOK_SECRET` | — | HMAC secret for webhook signing |
| `N8N_WEBHOOK_ENABLED` | `true` | Enable/disable webhook forwarding |
| `LOG_LEVEL` | `info` | Log level: debug/info/warn/error |
| `SERVER_POLL_INTERVAL` | `30` | Server stats poll interval (seconds) |
| `RESOURCE_POLL_INTERVAL` | `10` | Container resource poll interval (seconds) |
| `SERVICE_POLL_INTERVAL` | `60` | Service status poll interval (seconds) |

---

## Docker Event Output Example

```json
{
  "server": "server-a",
  "container": "hrd-push",
  "image": "go-fcm",
  "action": "start",
  "event_type": "container",
  "status": "start",
  "time": "2026-07-12T09:30:00+07:00"
}
```

---

## Architecture

```
Docker Engine
      │ Docker Events API (streaming)
      ▼
Devert Monitor Agent (Go)
      │
┌─────┼──────────┐
▼     ▼          ▼
PostgreSQL  n8n Webhook  REST API
```

---

## Project Structure

```
├── cmd/agent/main.go          # Entry point
├── internal/
│   ├── config/                # ENV config loader
│   ├── logger/                # zerolog setup
│   ├── database/              # pgx pool + event log repository
│   ├── webhook/               # n8n sender with retry + HMAC
│   ├── docker/                # Docker modules (events, containers, resources, images, volumes, networks)
│   ├── server/                # Server monitoring (CPU, memory, disk, network, OS)
│   ├── services/              # Service status via systemctl
│   └── api/                   # Chi router + handlers + middleware
├── migrations/                # SQL schema
├── Dockerfile                 # Multi-stage build
├── docker-compose.yml
└── .env.example
```

---

## Security

- **Bearer Token** — all API endpoints require `Authorization: Bearer` header
- **Rate Limiting** — 100 req/min per IP by default (configurable)
- **Webhook HMAC** — every webhook POST includes `X-Webhook-Signature: sha256=...`
- **Non-root Docker** — container runs as unprivileged `agent` user
- **Docker socket** — mounted read-only (`ro`)

---

## Roadmap

| Phase | Status |
|-------|--------|
| Phase 1: Docker Events, Webhook, PostgreSQL, Container monitoring, REST API | ✅ Done |
| Phase 2: CPU, Memory, Disk, Network, Service monitoring | ✅ Done |
| Phase 3: RabbitMQ, PostgreSQL DB, Redis, SSL, Domain monitoring | 🔲 Planned |
| Phase 4: Multi-server, WebSocket, Remote management | 🔲 Planned |
| Phase 5: Kubernetes, AI anomaly detection | 🔲 Future |
