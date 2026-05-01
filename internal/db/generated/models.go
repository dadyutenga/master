package generated

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Company   string    `json:"company"`
	Phone     *string   `json:"phone"`
	Password  string    `json:"password"`
	Role      string    `json:"role"`
	Verified  bool      `json:"verified"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Tenant struct {
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
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type VerifyToken struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used"`
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

type Queries struct {
	db *sql.DB
}

func New(db *sql.DB) *Queries {
	return &Queries{db: db}
}