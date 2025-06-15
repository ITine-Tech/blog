package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

func NewMockStore() Storage {
	return Storage{
		Users: &MockUserStore{},
	}
}

type MockUserStore struct {
}

func (m *MockUserStore) Create(ctx context.Context, tx *sql.Tx, u *User) error {
	return nil
}

func (m *MockUserStore) GetAllUsers(context.Context) ([]*User, error) {
	return nil, errors.New("database connection failed")
}

func (m *MockUserStore) CreateAndInvite(context.Context, *User, string, time.Duration) error {
	return nil
}

func (m *MockUserStore) Activate(context.Context, string) error {
	return nil
}

func (m *MockUserStore) GetUserByID(context.Context, uuid.UUID) (*User, error) {
	return &User{}, nil

}

func (m *MockUserStore) GetUserByUsername(context.Context, string) (*User, error) {
	return nil, nil
}

func (m *MockUserStore) UpdateUser(context.Context, *User) error {
	return nil
}
func (m *MockUserStore) DeleteUser(context.Context, uuid.UUID) error {
	return nil
}
