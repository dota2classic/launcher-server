#!/usr/bin/env bash
set -euo pipefail

CERTBOT_EMAIL="admin@dotaclassic.ru"
DOMAIN="launcher.dotaclassic.ru"
PROJECT_DIR="/opt/launcher"

# ── Docker ────────────────────────────────────────────────────────────────────
echo "==> Installing Docker..."
curl -fsSL https://get.docker.com | sh
systemctl enable --now docker

# ── Project structure ─────────────────────────────────────────────────────────
echo "==> Creating project structure in $PROJECT_DIR..."
mkdir -p "$PROJECT_DIR"/{files,nginx/conf.d}
cd "$PROJECT_DIR"

# ── docker-compose.yml ────────────────────────────────────────────────────────
cat > docker-compose.yml <<'EOF'
services:
  launcher:
    image: dota2classic/launcher-server:latest
    restart: unless-stopped
    environment:
      LAUNCHER_FILES_PATH: /data/files
      LAUNCHER_ADDR: :8080
    volumes:
      - ./files:/data/files
    networks:
      - internal

  nginx:
    image: nginx:alpine
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/conf.d:/etc/nginx/conf.d:ro
      - certbot_certs:/etc/letsencrypt:ro
      - certbot_www:/var/www/certbot
    depends_on:
      - launcher
    networks:
      - internal

  certbot:
    image: certbot/certbot
    volumes:
      - certbot_certs:/etc/letsencrypt
      - certbot_www:/var/www/certbot
    entrypoint: >
      /bin/sh -c "trap exit TERM;
        while :; do
          certbot renew;
          sleep 12h & wait $${!};
        done"

volumes:
  certbot_certs:
  certbot_www:

networks:
  internal:
EOF

# ── Temporary HTTP-only nginx config (certbot needs port 80 before certs exist)
cat > nginx/conf.d/launcher.conf <<'EOF'
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
EOF

# ── Start nginx so certbot can complete the ACME http-01 challenge ─────────────
echo "==> Starting nginx (HTTP only)..."
docker compose up -d nginx

# ── Obtain initial certificate ────────────────────────────────────────────────
echo "==> Obtaining TLS certificate for $DOMAIN..."
docker compose run --rm certbot certonly \
    --webroot -w /var/www/certbot \
    -d "$DOMAIN" \
    --email "$CERTBOT_EMAIL" \
    --agree-tos --no-eff-email

# ── Full nginx config with TLS ────────────────────────────────────────────────
cat > nginx/conf.d/launcher.conf <<'EOF'
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

server {
    listen 443 ssl;
    server_name launcher.dotaclassic.ru;

    ssl_certificate     /etc/letsencrypt/live/launcher.dotaclassic.ru/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/launcher.dotaclassic.ru/privkey.pem;

    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_prefer_server_ciphers off;

    location / {
        proxy_pass         http://launcher:8080;
        proxy_set_header   Host              $host;
        proxy_set_header   X-Real-IP         $remote_addr;
        proxy_set_header   X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto $scheme;
        proxy_read_timeout 300s;
    }
}
EOF

# ── Reload nginx with TLS config, then bring up everything ───────────────────
echo "==> Reloading nginx with TLS config..."
docker compose exec nginx nginx -s reload

echo "==> Starting all services..."
docker compose up -d

echo ""
echo "Done. https://$DOMAIN is live."
