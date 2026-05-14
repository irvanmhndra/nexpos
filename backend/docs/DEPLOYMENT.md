# Nexpos Deployment Guide

## Subdomains
- **nexpos.irvanmahendra.com** → nexpos-web (React static files)
- **api.nexpos.irvanmahendra.com** → nexpos-api (Go API via Docker, port 8080)

---

## 1. VPS Initial Setup

### As root
```bash
mkdir -p /var/www/nexpos-web
chown -R deploy:deploy /var/www/nexpos-web
```

### As deploy user
```bash
mkdir -p ~/nexpos
```

Create `~/nexpos/.env`:
```bash
tee ~/nexpos/.env > /dev/null <<'EOF'
# Server
SERVER_PORT=8080
SERVER_ENV=production

# PostgreSQL
POSTGRES_HOST=nexpos_db
POSTGRES_PORT=5432
POSTGRES_USER=nexpos_user
POSTGRES_PASSWORD=your_postgres_password
POSTGRES_DB=nexpos_db
POSTGRES_SSLMODE=disable

# MongoDB
MONGO_USER=nexpos_mongo_user
MONGO_PASSWORD=your_mongo_password
MONGO_URI=mongodb://nexpos_mongo_user:your_mongo_password@nexpos_mongo:27017
MONGO_DB=nexpos

# JWT
JWT_SECRET=your_jwt_secret_here
JWT_ACCESS_EXPIRES_HOURS=2
JWT_REFRESH_EXPIRES_DAYS=7
EOF
chmod 600 ~/nexpos/.env
```

> Note: `MONGO_URI` uses the container name `nexpos_mongo` as the host since both services run on the same Docker network.

Copy `docker-compose.yml` to VPS:
```bash
scp docker-compose.yml deploy@<VPS_IP>:~/nexpos/
```

Login to GHCR on VPS (one-time):
```bash
echo <GHCR_TOKEN> | docker login ghcr.io -u irvanmhndra --password-stdin
```

Start services:
```bash
cd ~/nexpos
docker compose up -d
```

---

## 2. Nginx Config (as root)

```bash
tee /etc/nginx/sites-available/nexpos > /dev/null <<'EOF'
# nexpos-web
server {
    listen 80;
    server_name nexpos.irvanmahendra.com;

    root /var/www/nexpos-web;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }
}

# nexpos-api
server {
    listen 80;
    server_name api.nexpos.irvanmahendra.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
EOF

ln -s /etc/nginx/sites-available/nexpos /etc/nginx/sites-enabled/nexpos
nginx -t && systemctl reload nginx
```

---

## 3. HTTPS with Certbot (as root)

```bash
certbot --nginx -d nexpos.irvanmahendra.com -d api.nexpos.irvanmahendra.com
```

---

## 4. GitHub Actions Secrets

Add to **both** repos (Settings → Secrets → Actions):

| Secret | Value |
|--------|-------|
| `VPS_HOST` | your VPS IP |
| `VPS_USER` | `deploy` |
| `VPS_SSH_KEY` | private SSH key for deploy user |
| `GHCR_TOKEN` | GitHub personal access token (with `write:packages`) |

---

## 5. DNS Records

```
nexpos.irvanmahendra.com     A   <VPS_IP>
api.nexpos.irvanmahendra.com A   <VPS_IP>
```

---

## Useful Commands

```bash
# API logs
docker logs nexpos_api -f

# PostgreSQL logs
docker logs nexpos_db -f

# MongoDB logs
docker logs nexpos_mongo -f

# Restart API
docker compose -f ~/nexpos/docker-compose.yml restart nexpos-api

# Manual PostgreSQL migration
docker exec nexpos_api sh -c '/app/migrate -path /app/migrations -database "postgres://$POSTGRES_USER:$POSTGRES_PASSWORD@$POSTGRES_HOST:$POSTGRES_PORT/$POSTGRES_DB?sslmode=$POSTGRES_SSLMODE" up'

# MongoDB shell
docker exec -it nexpos_mongo mongosh -u $MONGO_USER -p $MONGO_PASSWORD --authenticationDatabase admin
```
