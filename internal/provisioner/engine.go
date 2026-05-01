package provisioner

import (
	"context"
	"database/sql"
	"log"

	"github.com/dadyutenga/hms-control/internal/config"
	"github.com/dadyutenga/hms-control/internal/db/generated"
	"github.com/dadyutenga/hms-control/internal/mailer"
	"github.com/google/uuid"
)

type Engine struct {
	cfg   *config.Config
	db   *sql.DB
	mail  *mailer.Mailer
	queue chan uuid.UUID
}

func NewEngine(cfg *config.Config, db *sql.DB, mail *mailer.Mailer) *Engine {
	return &Engine{
		cfg:   cfg,
		db:    db,
		mail:  mail,
		queue: make(chan uuid.UUID, 50),
	}
}

func (e *Engine) Start() {
	for i := range e.cfg.WorkerCount {
		go e.worker(i)
	}
	log.Printf("[provisioner] started %d workers", e.cfg.WorkerCount)
}

func (e *Engine) Enqueue(tenantID uuid.UUID) {
	e.queue <- tenantID
}

func (e *Engine) worker(id int) {
	log.Printf("[worker-%d] ready", id)
	for tenantID := range e.queue {
		log.Printf("[worker-%d] provisioning tenant %s", id, tenantID)
		e.provision(tenantID)
	}
}

func (e *Engine) provision(tenantID uuid.UUID) {
	ctx := context.Background()
	q := generated.New(e.db)

	tenant, err := q.GetTenantByID(ctx, tenantID)
	if err != nil {
		log.Printf("[provisioner] tenant %s not found: %v", tenantID, err)
		return
	}

	runner := NewRunner(e.cfg)
	logOutput, err := runner.Run(tenant)

	if err != nil {
		log.Printf("[provisioner] FAILED tenant %s: %v", tenant.Slug, err)
		q.SetTenantFailed(ctx, generated.SetTenantFailedParams{
			ID:           tenantID,
			ProvisionLog: &logOutput,
		})
		return
	}

	appKey, _ := runner.GetAppKey(tenant.Slug)

	q.SetTenantActive(ctx, generated.SetTenantActiveParams{
		ID:           tenantID,
		AppKey:       &appKey,
		ProvisionLog: &logOutput,
	})

	user, _ := q.GetUserByID(ctx, tenant.UserID)
	go e.mail.SendTenantReady(user.Email, user.Name, "https://"+tenant.Domain)

	log.Printf("[provisioner] SUCCESS tenant %s live at https://%s", tenant.Slug, tenant.Domain)
}