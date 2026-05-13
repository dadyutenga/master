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
	go e.billingChecker()
	log.Printf("[provisioner] started %d workers and billing checker", e.cfg.WorkerCount)
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

func (e *Engine) billingChecker() {
	const interval = 1 * time.Hour
	log.Printf("[billing-checker] starting, checking every %v", interval)
	for {
		time.Sleep(interval)
		e.checkOverdueTenants()
		e.checkOverdueInstances()
	}
}

func (e *Engine) checkOverdueTenants() {
	ctx := context.Background()
	q := generated.New(e.db)

	tenants, err := q.ListOverdueTenants(ctx)
	if err != nil {
		log.Printf("[billing-checker] failed to list overdue tenants: %v", err)
		return
	}

	runner := NewRunner(e.cfg)
	for _, tenant := range tenants {
		log.Printf("[billing-checker] marking tenant %s as overdue", tenant.Slug)

		if err := q.MarkTenantOverdue(ctx, tenant.ID); err != nil {
			log.Printf("[billing-checker] failed to mark %s overdue: %v", tenant.Slug, err)
			continue
		}

		if _, err := runner.StopTenant(tenant.Slug); err != nil {
			log.Printf("[billing-checker] failed to stop container for %s: %v", tenant.Slug, err)
		} else {
			log.Printf("[billing-checker] stopped container for overdue tenant %s", tenant.Slug)
		}

		if _, err := q.CreateDeployment(ctx, generated.CreateDeploymentParams{
			TenantID: tenant.ID,
			Action:   "billing_overdue",
			Status:   generated.DeploymentStatusStopped,
		}); err != nil {
			log.Printf("[billing-checker] failed to record deployment for %s: %v", tenant.Slug, err)
		}
	}
}

func (e *Engine) checkOverdueInstances() {
	ctx := context.Background()
	q := generated.New(e.db)

	instances, err := q.ListOverdueInstances(ctx)
	if err != nil {
		log.Printf("[billing-checker] failed to list overdue instances: %v", err)
		return
	}

	runner := NewRunner(e.cfg)
	for _, inst := range instances {
		log.Printf("[billing-checker] marking instance %s as overdue", inst.Slug)

		now := time.Now()
		if err := q.UpdateInstanceBilling(ctx, inst.ID, generated.BillingStatusOverdue, inst.LastPaymentAt, &now); err != nil {
			log.Printf("[billing-checker] failed to mark instance %s overdue: %v", inst.Slug, err)
			continue
		}

		if _, err := runner.StopInstance(inst.Slug); err != nil {
			log.Printf("[billing-checker] failed to stop container for instance %s: %v", inst.Slug, err)
		} else {
			log.Printf("[billing-checker] stopped container for overdue instance %s", inst.Slug)
		}

		if _, err := q.CreateInstanceDeployment(ctx, inst.ID, "billing_overdue", generated.DeploymentStatusStopped); err != nil {
			log.Printf("[billing-checker] failed to record deployment for instance %s: %v", inst.Slug, err)
		}
	}
}
