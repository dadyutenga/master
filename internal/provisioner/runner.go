package provisioner

import (
	"bytes"
	"os/exec"
	"strings"

	"github.com/dadyutenga/hms-control/internal/config"
	"github.com/dadyutenga/hms-control/internal/db/generated"
)

type Runner struct {
	cfg *config.Config
}

func NewRunner(cfg *config.Config) *Runner {
	return &Runner{cfg: cfg}
}

func (r *Runner) Run(t generated.Tenant) (string, error) {
	args := []string{
		t.Slug,
		t.Domain,
		t.DbName,
		t.DbUser,
		t.DbPassword,
		r.cfg.TenantDir,
		r.cfg.HMSSource,
	}

	cmd := exec.Command("sudo", append([]string{"bash", r.cfg.ProvisionScript}, args...)...)

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	err := cmd.Run()
	return buf.String(), err
}

func (r *Runner) GetAppKey(slug string) (string, error) {
	out, err := exec.Command(
		"docker", "exec", "hms_"+slug,
		"php", "artisan", "key:generate", "--show",
	).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}