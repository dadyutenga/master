package models

import (
	"database/sql"
	"time"
)

type PaymentMethod struct {
	ID            int64
	Name          string
	MethodType    string // card, lipa_namba, mobile
	APIKey        string
	APISecret     string
	WebhookSecret string
	CallbackURL   string
	LipaNamba     string
	IsActive      bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type PaymentMethodStore struct {
	db *sql.DB
}

func NewPaymentMethodStore(db *sql.DB) *PaymentMethodStore {
	return &PaymentMethodStore{db: db}
}

func (s *PaymentMethodStore) List() ([]PaymentMethod, error) {
	rows, err := s.db.Query(
		`SELECT id, name, method_type, api_key, api_secret, webhook_secret, callback_url, lipa_namba, is_active, created_at, updated_at
		 FROM payment_methods ORDER BY is_active DESC, name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var methods []PaymentMethod
	for rows.Next() {
		var m PaymentMethod
		if err := rows.Scan(&m.ID, &m.Name, &m.MethodType, &m.APIKey, &m.APISecret, &m.WebhookSecret, &m.CallbackURL, &m.LipaNamba, &m.IsActive, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		methods = append(methods, m)
	}
	return methods, nil
}

func (s *PaymentMethodStore) ListActive() ([]PaymentMethod, error) {
	rows, err := s.db.Query(
		`SELECT id, name, method_type, api_key, api_secret, webhook_secret, callback_url, lipa_namba, is_active, created_at, updated_at
		 FROM payment_methods WHERE is_active = 1 ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var methods []PaymentMethod
	for rows.Next() {
		var m PaymentMethod
		if err := rows.Scan(&m.ID, &m.Name, &m.MethodType, &m.APIKey, &m.APISecret, &m.WebhookSecret, &m.CallbackURL, &m.LipaNamba, &m.IsActive, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		methods = append(methods, m)
	}
	return methods, nil
}

func (s *PaymentMethodStore) GetByID(id int64) (*PaymentMethod, error) {
	var m PaymentMethod
	err := s.db.QueryRow(
		`SELECT id, name, method_type, api_key, api_secret, webhook_secret, callback_url, lipa_namba, is_active, created_at, updated_at
		 FROM payment_methods WHERE id = ?`, id,
	).Scan(&m.ID, &m.Name, &m.MethodType, &m.APIKey, &m.APISecret, &m.WebhookSecret, &m.CallbackURL, &m.LipaNamba, &m.IsActive, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *PaymentMethodStore) Create(name, methodType, apiKey, apiSecret, webhookSecret, callbackURL, lipaNamba string) (*PaymentMethod, error) {
	result, err := s.db.Exec(
		`INSERT INTO payment_methods (name, method_type, api_key, api_secret, webhook_secret, callback_url, lipa_namba, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		name, methodType, apiKey, apiSecret, webhookSecret, callbackURL, lipaNamba,
	)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return s.GetByID(id)
}

func (s *PaymentMethodStore) Update(id int64, name, methodType, apiKey, apiSecret, webhookSecret, callbackURL, lipaNamba string) error {
	_, err := s.db.Exec(
		`UPDATE payment_methods
		 SET name = ?, method_type = ?, api_key = ?, api_secret = ?, webhook_secret = ?, callback_url = ?, lipa_namba = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		name, methodType, apiKey, apiSecret, webhookSecret, callbackURL, lipaNamba, id,
	)
	return err
}

func (s *PaymentMethodStore) ToggleActive(id int64, active bool) error {
	_, err := s.db.Exec(`UPDATE payment_methods SET is_active = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, active, id)
	return err
}

func (s *PaymentMethodStore) Delete(id int64) error {
	_, err := s.db.Exec(`DELETE FROM payment_methods WHERE id = ?`, id)
	return err
}
