package models

import (
	"database/sql"
	"time"
)

type DockerTemplate struct {
	ID           int64
	Name         string
	Description  string
	TemplateBody string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type DockerTemplateStore struct {
	db *sql.DB
}

func NewDockerTemplateStore(db *sql.DB) *DockerTemplateStore {
	return &DockerTemplateStore{db: db}
}

func (s *DockerTemplateStore) List() ([]DockerTemplate, error) {
	rows, err := s.db.Query(
		`SELECT id, name, description, template_body, created_at, updated_at
		 FROM docker_templates ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []DockerTemplate
	for rows.Next() {
		var t DockerTemplate
		if err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.TemplateBody, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		templates = append(templates, t)
	}
	return templates, nil
}

func (s *DockerTemplateStore) ListNames() ([]string, error) {
	rows, err := s.db.Query(`SELECT name FROM docker_templates ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		rows.Scan(&name)
		names = append(names, name)
	}
	return names, nil
}

func (s *DockerTemplateStore) GetByID(id int64) (*DockerTemplate, error) {
	var t DockerTemplate
	err := s.db.QueryRow(
		`SELECT id, name, description, template_body, created_at, updated_at
		 FROM docker_templates WHERE id = ?`, id,
	).Scan(&t.ID, &t.Name, &t.Description, &t.TemplateBody, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *DockerTemplateStore) GetByName(name string) (*DockerTemplate, error) {
	var t DockerTemplate
	err := s.db.QueryRow(
		`SELECT id, name, description, template_body, created_at, updated_at
		 FROM docker_templates WHERE name = ?`, name,
	).Scan(&t.ID, &t.Name, &t.Description, &t.TemplateBody, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *DockerTemplateStore) Create(name, description, templateBody string) (*DockerTemplate, error) {
	result, err := s.db.Exec(
		`INSERT INTO docker_templates (name, description, template_body, created_at, updated_at)
		 VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		name, description, templateBody,
	)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return s.GetByID(id)
}

func (s *DockerTemplateStore) Update(id int64, name, description, templateBody string) error {
	_, err := s.db.Exec(
		`UPDATE docker_templates
		 SET name = ?, description = ?, template_body = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ?`,
		name, description, templateBody, id,
	)
	return err
}

func (s *DockerTemplateStore) Delete(id int64) error {
	_, err := s.db.Exec(`DELETE FROM docker_templates WHERE id = ?`, id)
	return err
}

func (s *DockerTemplateStore) NameExists(name string, excludeID int64) bool {
	var count int
	if excludeID > 0 {
		s.db.QueryRow(`SELECT COUNT(*) FROM docker_templates WHERE name = ? AND id != ?`, name, excludeID).Scan(&count)
	} else {
		s.db.QueryRow(`SELECT COUNT(*) FROM docker_templates WHERE name = ?`, name).Scan(&count)
	}
	return count > 0
}
