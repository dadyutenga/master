#!/bin/bash

echo "=== Rebuilding HMS image ==="
docker build -t hms-app:latest /opt/hms-source

echo "=== Rolling update across all tenants ==="
for dir in /opt/tenants/*/; do
    slug=$(basename "$dir")
    echo "  → ${slug}"
    cd "$dir"
    docker compose up -d --build
    docker exec "hms_${slug}" php artisan migrate --force
    docker exec "hms_${slug}" php artisan config:cache
done

echo "=== All tenants updated ==="