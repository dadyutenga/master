package models

import (
	"database/sql"
	"time"
)

type DockerTemplate struct {
	ID        int64
	Name      string
	Body      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type DockerTemplateStore struct{ db *sql.DB }

func NewDockerTemplateStore(db *sql.DB) *DockerTemplateStore {
	return &DockerTemplateStore{db: db}
}

func (s *DockerTemplateStore) List() ([]DockerTemplate, error) {
	rows, err := s.db.Query(`
		SELECT id, name, template_body, created_at, updated_at
		FROM docker_templates
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []DockerTemplate
	for rows.Next() {
		var t DockerTemplate
		var createdAt, updatedAt string
		if err := rows.Scan(&t.ID, &t.Name, &t.Body, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		t.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		t.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
		templates = append(templates, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return templates, nil
}

func (s *DockerTemplateStore) GetByID(id int64) (DockerTemplate, error) {
	var t DockerTemplate
	var createdAt, updatedAt string
	err := s.db.QueryRow(`
		SELECT id, name, template_body, created_at, updated_at
		FROM docker_templates WHERE id = ?`, id,
	).Scan(&t.ID, &t.Name, &t.Body, &createdAt, &updatedAt)
	if err != nil {
		return t, err
	}
	t.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	t.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
	return t, nil
}

func (s *DockerTemplateStore) GetByName(name string) (DockerTemplate, error) {
	var t DockerTemplate
	var createdAt, updatedAt string
	err := s.db.QueryRow(`
		SELECT id, name, template_body, created_at, updated_at
		FROM docker_templates WHERE name = ?`, name,
	).Scan(&t.ID, &t.Name, &t.Body, &createdAt, &updatedAt)
	if err != nil {
		return t, err
	}
	t.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	t.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
	return t, nil
}

func (s *DockerTemplateStore) Create(name, body string) (DockerTemplate, error) {
	res, err := s.db.Exec(`
		INSERT INTO docker_templates (name, template_body, created_at, updated_at)
		VALUES (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`, name, body)
	if err != nil {
		return DockerTemplate{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return DockerTemplate{}, err
	}
	return s.GetByID(id)
}

func (s *DockerTemplateStore) Update(id int64, name, body string) (DockerTemplate, error) {
	_, err := s.db.Exec(`
		UPDATE docker_templates
		SET name = ?, template_body = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`, name, body, id)
	if err != nil {
		return DockerTemplate{}, err
	}
	return s.GetByID(id)
}

func (s *DockerTemplateStore) Delete(id int64) error {
	_, err := s.db.Exec(`DELETE FROM docker_templates WHERE id = ?`, id)
	return err
}
