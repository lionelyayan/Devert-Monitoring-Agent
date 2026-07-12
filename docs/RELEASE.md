# 📦 Tutorial: Push Tag & Release ke GitHub Package

Panduan lengkap dari awal hingga image tersedia di GHCR.

---

## Prasyarat

- Git sudah terinstall dan terkonfigurasi
- Akun GitHub
- Repository sudah dibuat di GitHub

---

## Langkah 1 — Setup Repository (Sekali Saja)

### Inisialisasi Git & Push Pertama

```bash
# Masuk ke folder project
cd "Devert Monitoring Agent"

# Init git (jika belum)
git init

# Tambahkan semua file
git add .

# Commit pertama
git commit -m "feat: initial commit — Devert Monitor Agent v1.0.0"

# Hubungkan ke GitHub
git remote add origin https://github.com/lionelyayan/Devert-Monitoring-Agent.git

# Push ke branch main
git push -u origin main
```

---

## Langkah 2 — Aktifkan GHCR di Repository

1. Buka repository di GitHub
2. Klik **Settings** → **Actions** → **General**
3. Scroll ke bawah ke **Workflow permissions**
4. Pilih **Read and write permissions** ✅
5. Klik **Save**

> Ini memastikan workflow bisa push image ke GHCR menggunakan `GITHUB_TOKEN`.

---

## Langkah 3 — Push Tag Pertama (Release)

### Format Tag

Tag harus menggunakan format **Semantic Versioning**:

```
v{MAJOR}.{MINOR}.{PATCH}

Contoh:
  v1.0.0   → rilis pertama
  v1.1.0   → fitur baru (minor)
  v1.1.1   → bugfix (patch)
  v2.0.0   → breaking change (major)
```

### Cara Push Tag

```bash
# 1. Pastikan working tree bersih
git status

# 2. Commit semua perubahan terlebih dahulu
git add .
git commit -m "release: v1.0.0"

# 3. Push commit ke main
git push origin main

# 4. Buat tag dengan pesan
git tag -a v1.0.0 -m "Release v1.0.0 — Initial release"

# 5. Push TAG ke GitHub (ini yang memicu workflow!)
git push origin v1.0.0
```

✅ Setelah `git push origin v1.0.0`, workflow otomatis berjalan di GitHub Actions.

---

## Langkah 4 — Pantau Workflow di GitHub

1. Buka `https://github.com/lionelyayan/Devert-Monitoring-Agent/actions`
2. Klik workflow **"Build & Publish Docker Image"** yang sedang berjalan
3. Tunggu hingga semua step selesai (~3–5 menit untuk build multi-arch)

Setelah selesai, image tersedia di:
```
ghcr.io/lionelyayan/devert-monitoring-agent:latest
ghcr.io/lionelyayan/devert-monitoring-agent:v1.0.0
ghcr.io/lionelyayan/devert-monitoring-agent:1.0.0
ghcr.io/lionelyayan/devert-monitoring-agent:1.0
ghcr.io/lionelyayan/devert-monitoring-agent:1
```

---

## Langkah 5 — Setup di Server (Sekali Saja)

```bash
# Masuk ke server via SSH
ssh user@your-server-ip

# Buat direktori
mkdir -p /opt/devert && cd /opt/devert

# Download hanya 2 file yang dibutuhkan
curl -O https://raw.githubusercontent.com/lionelyayan/Devert-Monitoring-Agent/main/docker-compose.prod.yml
curl -O https://raw.githubusercontent.com/lionelyayan/Devert-Monitoring-Agent/main/.env.example

# Buat file .env
cp .env.example .env
nano .env
```

### Isi `.env` di Server

```env
# GHCR
GITHUB_USERNAME=lionelyayan
GITHUB_REPO=devert-monitoring-agent
IMAGE_TAG=latest           # atau v1.0.0 untuk pin ke versi spesifik

# Server
SERVER_NAME=production-server
TIMEZONE=Asia/Jakarta
HTTP_PORT=8080
API_TOKEN=isi-dengan-token-rahasia-yang-panjang

# PostgreSQL
POSTGRES_DSN=postgres://devert:password_aman@postgres:5432/devert_monitor?sslmode=disable
POSTGRES_USER=devert
POSTGRES_PASSWORD=password_aman
POSTGRES_DB=devert_monitor

# n8n Webhook (opsional)
N8N_WEBHOOK_URL=https://n8n.domain.com/webhook/docker-events
N8N_WEBHOOK_SECRET=secret-hmac
N8N_WEBHOOK_ENABLED=true

LOG_LEVEL=info
```

### Login GHCR & Jalankan

```bash
# Login ke GHCR (gunakan GitHub Personal Access Token dengan scope read:packages)
echo YOUR_GITHUB_TOKEN | docker login ghcr.io -u lionelyayan --password-stdin

# Pull & jalankan
docker compose -f docker-compose.prod.yml up -d

# Verifikasi
curl http://localhost:8080/health
```

---

## Workflow Harian — Cara Release Versi Baru

```bash
# 1. Lakukan perubahan kode
# ...edit file...

# 2. Commit perubahan
git add .
git commit -m "feat: tambah fitur monitoring Redis"

# 3. Push ke main
git push origin main

# 4. Buat dan push tag baru
git tag -a v1.1.0 -m "Release v1.1.0 — tambah Redis monitoring"
git push origin v1.1.0

# ✅ Workflow otomatis build & push image baru ke GHCR
```

### Update di Server

```bash
cd /opt/devert

# Pull image terbaru
docker compose -f docker-compose.prod.yml pull

# Restart dengan image baru (zero-downtime rolling update)
docker compose -f docker-compose.prod.yml up -d

# Verifikasi versi yang berjalan
docker inspect devert-monitor-agent | grep -i version
```

---

## Cheatsheet Git Tag

```bash
# Lihat semua tag yang ada
git tag -l

# Lihat detail tag
git show v1.0.0

# Hapus tag lokal (jika salah buat)
git tag -d v1.0.0

# Hapus tag di GitHub (jika sudah terlanjur push)
git push origin --delete v1.0.0

# Buat ulang tag di commit tertentu
git tag -a v1.0.0 <commit-hash> -m "Release v1.0.0"
git push origin v1.0.0
```

---

## Struktur File di Server

```
/opt/devert/
├── .env                    ← Konfigurasi (JANGAN di-commit!)
└── docker-compose.prod.yml ← Definisi service (pull dari GitHub)
```

**Tidak perlu menyimpan source code di server sama sekali.**
Image sudah berisi binary yang siap jalan. ✅

---

## Troubleshooting

### Image tidak bisa di-pull (unauthorized)

```bash
# Login ulang ke GHCR
echo TOKEN | docker login ghcr.io -u lionelyayan --password-stdin
```

### Package masih private di GHCR

1. Buka `https://github.com/lionelyayan?tab=packages`
2. Klik package **devert-monitoring-agent**
3. **Package settings** → **Change visibility** → **Public**

Atau tetap private dan gunakan PAT dengan scope `read:packages`.

### Workflow tidak terpicu

Pastikan format tag benar: harus diawali `v` dan memiliki 3 bagian angka.
```bash
# ✅ Benar
git tag v1.0.0
git tag v2.1.3

# ❌ Salah (tidak akan memicu workflow)
git tag 1.0.0
git tag release-1.0
git tag v1.0
```
