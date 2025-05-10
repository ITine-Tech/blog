package store

import (
	"context"
	"database/sql"
)

type Role struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Level       int    `json:"level"`
	Description string `json:"description"`
}

type RolePostgreStore struct {
	db *sql.DB
}

func (s *RolePostgreStore) GetByName(ctx context.Context, slug string) (*Role, error) {
	query := `
SELECT id, name, level, description
FROM roles
WHERE name = $1`

	role := new(Role)
	err := s.db.QueryRowContext(ctx, query, slug).Scan(
		&role.ID,
		&role.Name,
		&role.Level,
		&role.Description,
	)
	if err != nil {
		return nil, err
	}
	return role, err
}
