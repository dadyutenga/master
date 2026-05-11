package provisioner

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/dadyutenga/hms-control/internal/config"
	"github.com/dadyutenga/hms-control/internal/db/generated"
	"github.com/dadyutenga/hms-control/internal/mailer"
	"github.com/google/uuid"
)

type Engine struct {
	cfg   *config.Config
	db    *sql.DB
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
		now := time.Now()
		errMsg := err.Error()
		var logPtr *string
		if logOutput != "" {
			logPtr = &logOutput
		}
		var errPtr *string
		if errMsg != "" {
			errPtr = &errMsg
		}
		if err := q.UpdateLatestDeploymentStatus(ctx, generated.UpdateLatestDeploymentStatusParams{
			TenantID:     tenantID,
			Status:       generated.DeploymentStatusFailed,
			Log:          logPtr,
			ErrorMessage: errPtr,
			CompletedAt:  &now,
		}); err != nil {
			log.Printf("[provisioner] FAILED to update deployment status for %s: %v", tenant.Slug, err)
		}
		if err := q.SetTenantFailed(ctx, generated.SetTenantFailedParams{
			ID:           tenantID,
			ProvisionLog: &logOutput,
		}); err != nil {
			log.Printf("[provisioner] FAILED to mark tenant %s as failed: %v", tenant.Slug, err)
		}
		return
	}

	appKey, err := runner.GetAppKey(tenant.Slug)
	if err != nil {
		log.Printf("[provisioner] WARNING: failed to get app key for %s: %v", tenant.Slug, err)
		appKey = ""
	}

	if err := q.SetTenantActive(ctx, generated.SetTenantActiveParams{
		ID:           tenantID,
		AppKey:       &appKey,
		ProvisionLog: &logOutput,
	}); err != nil {
		log.Printf("[provisioner] FAILED to mark tenant %s active: %v", tenant.Slug, err)
		return
	}

	now := time.Now()
	var logPtr *string
	if logOutput != "" {
		logPtr = &logOutput
	}
	if err := q.UpdateLatestDeploymentStatus(ctx, generated.UpdateLatestDeploymentStatusParams{
		TenantID:    tenantID,
		Status:      generated.DeploymentStatusActive,
		Log:         logPtr,
		CompletedAt: &now,
	}); err != nil {
		log.Printf("[provisioner] FAILED to update deployment status for %s: %v", tenant.Slug, err)
	}

	user, err := q.GetUserByID(ctx, tenant.UserID)
	if err != nil {
		log.Printf("[provisioner] WARNING: failed to load user for %s: %v", tenant.Slug, err)
		return
	}
	go e.mail.SendTenantReady(user.Email, user.Name, "https://"+tenant.Domain)

	log.Printf("[provisioner] SUCCESS tenant %s live at https://%s", tenant.Slug, tenant.Domain)
}
