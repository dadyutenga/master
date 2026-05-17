package generated

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Company     string    `json:"company"`
	Phone       *string   `json:"phone"`
	Password    string    `json:"password"`
	Role        string    `json:"role"`
	Verified    bool      `json:"verified"`
	TIN         *string   `json:"tin"`
	BrelaNumber *string   `json:"brela_number"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Tenant struct {
	ID                 uuid.UUID  `json:"id"`
	UserID             int64      `json:"user_id"`
	CompanyName        string     `json:"company_name"`
	Slug               string     `json:"slug"`
	Domain             string     `json:"domain"`
	DbName             string     `json:"db_name"`
	DbUser             string     `json:"db_user"`
	DbPassword         string     `json:"db_password"`
	AppKey             *string    `json:"app_key"`
	Status             string     `json:"status"`
	ProvisionLog       *string    `json:"provision_log"`
	ApprovedAt         *time.Time `json:"approved_at"`
	ProvisionedAt      *time.Time `json:"provisioned_at"`
	HotelName          *string    `json:"hotel_name"`
	Category           *string    `json:"category"`
	RoomCount          *int64     `json:"room_count"`
	Address            *string    `json:"address"`
	City               *string    `json:"city"`
	Country            *string    `json:"country"`
	RequestedSubdomain *string    `json:"requested_subdomain"`
	AdminName          *string    `json:"admin_name"`
	AdminEmail         *string    `json:"admin_email"`
	AdminPhone         *string    `json:"admin_phone"`
	BillingStatus      string     `json:"billing_status"`
	LastPaymentAt      *time.Time `json:"last_payment_at"`
	NextDueAt          *time.Time `json:"next_due_at"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type Document struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	TenantID     *string   `json:"tenant_id"`
	DocType      string    `json:"doc_type"`
	Filename     string    `json:"filename"`
	OriginalName string    `json:"original_name"`
	MimeType     string    `json:"mime_type"`
	SizeBytes    int64     `json:"size_bytes"`
	CreatedAt    time.Time `json:"created_at"`
}

type VerifyToken struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used"`
}

type ContactDetails struct {
	ID          int64     `json:"id"`
	Location    string    `json:"location"`
	PhoneNumber string    `json:"phone_number"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Deployment struct {
	ID           uuid.UUID  `json:"id"`
	TenantID     uuid.UUID  `json:"tenant_id"`
	Action       string     `json:"action"`
	Status       string     `json:"status"`
	Log          *string    `json:"log"`
	ErrorMessage *string    `json:"error_message"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	CompletedAt  *time.Time `json:"completed_at"`
}

type ListTenantsRow struct {
	ID            uuid.UUID  `json:"id"`
	UserID        int64      `json:"user_id"`
	CompanyName   string     `json:"company_name"`
	Slug          string     `json:"slug"`
	Domain        string     `json:"domain"`
	DbName        string     `json:"db_name"`
	DbUser        string     `json:"db_user"`
	DbPassword    string     `json:"db_password"`
	AppKey        *string    `json:"app_key"`
	Status        string     `json:"status"`
	ProvisionLog  *string    `json:"provision_log"`
	ApprovedAt    *time.Time `json:"approved_at"`
	ProvisionedAt *time.Time `json:"provisioned_at"`
	BillingStatus string     `json:"billing_status"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	UserName      string     `json:"user_name"`
	UserEmail     string     `json:"user_email"`
}

type GetVerifyTokenRow struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used"`
	Uid       int64     `json:"uid"`
}

type BillingTransaction struct {
	ID              int64      `json:"id"`
	TenantID        string     `json:"tenant_id"`
	Amount          float64    `json:"amount"`
	Currency        string     `json:"currency"`
	Description     string     `json:"description"`
	TransactionType string     `json:"transaction_type"`
	Status          string     `json:"status"`
	AdminID         *int64     `json:"admin_id"`
	CreatedAt       time.Time  `json:"created_at"`
}

type Instance struct {
	ID             uuid.UUID  `json:"id"`
	TenantID       uuid.UUID  `json:"tenant_id"`
	HotelName      string     `json:"hotel_name"`
	Slug           string     `json:"slug"`
	Domain         string     `json:"domain"`
	DbName         string     `json:"db_name"`
	DbUser         string     `json:"db_user"`
	DbPassword     string     `json:"db_password"`
	AppKey         *string    `json:"app_key"`
	Status         string     `json:"status"`
	AdminDisabled  bool       `json:"admin_disabled"`
	BillingStatus  string     `json:"billing_status"`
	Price          float64    `json:"price"`
	PackageName    string     `json:"package_name"`
	LastPaymentAt  *time.Time `json:"last_payment_at"`
	NextDueAt      *time.Time `json:"next_due_at"`
	ProvisionLog   *string    `json:"provision_log"`
	ApprovedAt     *time.Time `json:"approved_at"`
	ProvisionedAt  *time.Time `json:"provisioned_at"`
	ArchivedAt     *time.Time `json:"archived_at"`
	DeletedAt      *time.Time `json:"deleted_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type InstanceDeployment struct {
	ID           uuid.UUID  `json:"id"`
	InstanceID   uuid.UUID  `json:"instance_id"`
	Action       string     `json:"action"`
	Status       string     `json:"status"`
	Log          *string    `json:"log"`
	ErrorMessage *string    `json:"error_message"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	CompletedAt  *time.Time `json:"completed_at"`
}

type ListPaymentsRow struct {
	ID              int64     `json:"id"`
	TenantID        string    `json:"tenant_id"`
	Amount          float64   `json:"amount"`
	Currency        string    `json:"currency"`
	Description     string    `json:"description"`
	TransactionType string    `json:"transaction_type"`
	Status          string    `json:"status"`
	AdminID         *int64    `json:"admin_id"`
	CreatedAt       time.Time `json:"created_at"`
	CompanyName     string    `json:"company_name"`
	BillingStatus   string    `json:"billing_status"`
	UserName        string    `json:"user_name"`
	UserEmail       string    `json:"user_email"`
}

type BillingPackage struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Price           float64   `json:"price"`
	Currency        string    `json:"currency"`
	BillingCycle    string    `json:"billing_cycle"`
	IsActive        bool      `json:"is_active"`
	DockerTemplateID *int64   `json:"docker_template_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type Queries struct {
	db *sql.DB
}

func New(db *sql.DB) *Queries {
	return &Queries{db: db}
}
