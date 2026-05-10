#!/bin/bash
set -euo pipefail

SLUG=$1
DOMAIN=$2
DB_NAME=$3
DB_USER=$4
DB_PASS=$5
TENANT_DIR=$6
HMS_SOURCE=$7

DIR="${TENANT_DIR}/${SLUG}"

echo "[provision] Starting tenant: ${SLUG}"
echo "[provision] Domain:          ${DOMAIN}"

mkdir -p "${DIR}/storage/app/public"
mkdir -p "${DIR}/storage/framework/cache"
mkdir -p "${DIR}/storage/framework/sessions"
mkdir -p "${DIR}/storage/framework/views"
mkdir -p "${DIR}/storage/logs"

cat > "${DIR}/.env" <<EOF
APP_NAME="HMS - ${SLUG}"
APP_ENV=production
APP_KEY=
APP_DEBUG=false
APP_URL=https://${DOMAIN}

DB_CONNECTION=pgsql
DB_HOST=postgres_${SLUG}
DB_PORT=5432
DB_DATABASE=${DB_NAME}
DB_USERNAME=${DB_USER}
DB_PASSWORD=${DB_PASS}

CACHE_DRIVER=file
SESSION_DRIVER=file
QUEUE_CONNECTION=sync
EOF

TEMPLATE="/opt/hms-control/docker-templates/docker-compose.template.yml"
sed \
  -e "s|{{SLUG}}|${SLUG}|g" \
  -e "s|{{DOMAIN}}|${DOMAIN}|g" \
  -e "s|{{DB_NAME}}|${DB_NAME}|g" \
  -e "s|{{DB_USER}}|${DB_USER}|g" \
  -e "s|{{DB_PASS}}|${DB_PASS}|g" \
  "${TEMPLATE}" > "${DIR}/docker-compose.yml"

cd "${DIR}"
docker compose up -d --build

echo "[provision] Waiting for PostgreSQL..."
for i in $(seq 1 30); do
    docker exec "postgres_${SLUG}" pg_isready -U "${DB_USER}" && break
    sleep 2
done

docker exec "hms_${SLUG}" php artisan key:generate --force
docker exec "hms_${SLUG}" php artisan migrate --force
docker exec "hms_${SLUG}" php artisan db:seed --class=DefaultDataSeeder --force
docker exec "hms_${SLUG}" php artisan config:cache
docker exec "hms_${SLUG}" php artisan route:cache
docker exec "hms_${SLUG}" php artisan storage:link

echo "[provision] DONE — https://${DOMAIN} is live"