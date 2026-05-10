# HMS Control - Run Guide

## Prerequisites
- Go installed
- Air installed: `go install github.com/air-verse/air@latest`

## Configure environment
Create or edit `.env` in the project root. Minimum required:

```
APP_URL=http://localhost:8080
BASE_DOMAIN=localhost
DB_PATH=./data/hms_master.db
SESSION_SECRET=change-me-32-chars-minimum
COOKIE_SECURE=false
```

Optional mail settings:

```
SMTP_HOST=
SMTP_PORT=587
SMTP_USER=
SMTP_PASS=
SMTP_FROM=noreply@localhost
```

## Run migrations

```
go run ./cmd/migrate up
```

## Run the app with live reload

```
air
```

The server runs on `http://localhost:8080`.

## Run without Air (optional)

```
go run .
```

## Notes
- Approving tenants triggers the provisioner. The default provisioner uses `sudo` and Docker in `internal/provisioner/runner.go`, which is not Windows-friendly. For local dev, avoid approve actions or adjust `PROVISION_SCRIPT` to a Windows-compatible script.
