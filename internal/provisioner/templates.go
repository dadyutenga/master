package provisioner

import (
	"fmt"
	"os"
	"strings"
)

type TenantVars struct {
	Slug   string
	Domain string
	DBName string
	DBUser string
	DBPass string
}

func GenerateEnv(dir string, v TenantVars) error {
	content := fmt.Sprintf(`APP_NAME="HMS - %s"
APP_ENV=production
APP_KEY=
APP_DEBUG=false
APP_URL=https://%s

DB_CONNECTION=pgsql
DB_HOST=postgres_%s
DB_PORT=5432
DB_DATABASE=%s
DB_USERNAME=%s
DB_PASSWORD=%s

CACHE_DRIVER=file
SESSION_DRIVER=file
QUEUE_CONNECTION=sync
`, v.Slug, v.Domain, v.Slug, v.DBName, v.DBUser, v.DBPass)

	return os.WriteFile(dir+"/.env", []byte(content), 0600)
}

func GenerateCompose(templatePath, outputPath string, v TenantVars) error {
	raw, err := os.ReadFile(templatePath)
	if err != nil {
		return err
	}

	result := strings.NewReplacer(
		"{{SLUG}}", v.Slug,
		"{{DOMAIN}}", v.Domain,
		"{{DB_NAME}}", v.DBName,
		"{{DB_USER}}", v.DBUser,
		"{{DB_PASS}}", v.DBPass,
	).Replace(string(raw))

	return os.WriteFile(outputPath, []byte(result), 0644)
}