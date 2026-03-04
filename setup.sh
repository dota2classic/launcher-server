#!/usr/bin/env bash
# One-time server bootstrap. After this, all updates are: git pull && docker compose up -d
set -euo pipefail

CERTBOT_EMAIL="admin@dotaclassic.ru"
DOMAIN="launcher.dotaclassic.ru"
REPO="https://github.com/dota2classic/launcher-host.git"
PROJECT_DIR="/opt/launcher"

# ── Docker ────────────────────────────────────────────────────────────────────
echo "==> Installing Docker..."
curl -fsSL https://get.docker.com | sh
systemctl enable --now docker

# ── Clone repo ────────────────────────────────────────────────────────────────
echo "==> Cloning repo to $PROJECT_DIR..."
git clone "$REPO" "$PROJECT_DIR"
cd "$PROJECT_DIR"

# ── Files directory (not in git) ──────────────────────────────────────────────
mkdir -p files

# ── Bootstrap: start nginx with HTTP-only config so certbot can run ───────────
echo "==> Starting nginx (HTTP only for ACME challenge)..."
mkdir -p nginx/conf.d
cat > nginx/conf.d/launcher.conf <<'BOOTSTRAP'
server {
    listen 80;
    server_name launcher.dotaclassic.ru;

    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }

    location / {
        return 301 https://$host$request_uri;
    }
}
BOOTSTRAP

docker compose up -d nginx

# ── Obtain TLS certificate ────────────────────────────────────────────────────
echo "==> Obtaining TLS certificate for $DOMAIN..."
docker compose run --rm --entrypoint certbot certbot certonly \
    --webroot -w /var/www/certbot \
    -d "$DOMAIN" \
    --email "$CERTBOT_EMAIL" \
    --agree-tos --no-eff-email

# ── Restore real nginx config from repo and start everything ──────────────────
echo "==> Restoring nginx config from repo..."
git checkout nginx/conf.d/launcher.conf

echo "==> Starting all services..."
docker compose up -d

echo ""
echo "Done. https://$DOMAIN is live."
echo "Future updates: git pull && docker compose up -d && docker compose exec nginx nginx -s reload"
