package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"
)

var (
	ErrDuplicateEmail    = errors.New("a user with this email already exists")
	ErrDuplicateUsername = errors.New("a user with this username already exists")
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsActive  bool      `json:"is_active"`
	RoleID    int64     `json:"role_id"`
	Role      Role      `json:"role"`
}

type password struct {
	text *string
	hash []byte
}

func (p *password) Set(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	p.text = &password
	p.hash = hash

	return nil
}

func (p *password) Compare(password string) error {
	return bcrypt.CompareHashAndPassword(p.hash, []byte(password))
}

type UsersPostgresStore struct {
	db *sql.DB
}

func (s *UsersPostgresStore) Create(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `
		INSERT INTO users (username, email, password, role_id)
		VALUES ($1, $2, $3, (SELECT id FROM roles WHERE name = $4))
		RETURNING id, created_at, updated_at
		`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	role := user.Role.Name
	if role == "" {
		role = "user"
	}

	err := tx.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.Email,
		user.Password.hash,
		role,
	).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		switch {
		case err.Error() == "pq: duplicate key value violates unique constraint \"users_email_key\"":
			return ErrDuplicateEmail
		case err.Error() == "pq:duplicate key value violates unique constraint \"users_username_key\"":
			return ErrDuplicateUsername
		default:
			return err
		}
	}
	return nil
}

// CreateAndInvite creates a new user and sends an invitation email.
// If something fails during the process, the user will be deleted.
//
// ctx: The context for the operation.
// user: The user to be created. The user's password should be set before calling this function.
// token: The invitation token to be sent in the email.
// invitationExp: The expiration duration of the invitation token.
//
// Returns an error if the operation fails, or nil if successful.
func (s *UsersPostgresStore) CreateAndInvite(ctx context.Context, user *User, token string, invitationExp time.Duration) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		if err := s.Create(ctx, tx, user); err != nil {
			return err
		}

		err := s.createUserInvitation(ctx, tx, token, invitationExp, user.ID)
		if err != nil {
			return err
		}

		return nil
	})
}

// Activate activates a user account using an invitation token.
// It retrieves the user from the invitation token, sets the user's status to active,
// and deletes the corresponding invitation record.
//
// ctx: The context for the operation.
// token: The invitation token provided by the user.
//
// Returns an error if the operation fails, or nil if successful.
func (s *UsersPostgresStore) Activate(ctx context.Context, token string) error {
	// Video 45 7:50
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		user, err := s.getUserFromInvitation(ctx, tx, token)
		if err != nil {
			return err
		}
		user.IsActive = true

		if err := s.update(ctx, tx, user); err != nil {
			return err
		}

		if err := s.deleteUserInvitation(ctx, tx, user.ID); err != nil {
			return err
		}
		return nil
	})
}

func (s *UsersPostgresStore) GetAllUsers(ctx context.Context) ([]*User, error) {
	rows, err := s.db.Query(`SELECT * FROM users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*User

	for rows.Next() {
		user := &User{}
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Password.hash,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.IsActive,
			&user.RoleID,
		)

		if err != nil {
			return nil, err
		}

		result = append(result, user)
	}

	return result, nil
}

func (s *UsersPostgresStore) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	query := `
		SELECT users.id, username, email, password, created_at, updated_at, is_active, roles.id, roles.name, roles.level, roles.description
		FROM users
		JOIN roles ON (users.role_id = roles.id)
		WHERE users.id = $1
		`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var user User

	err := s.db.QueryRowContext(
		ctx,
		query,
		id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password.hash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.IsActive,
		&user.Role.ID,
		&user.Role.Name,
		&user.Role.Level,
		&user.Role.Description,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}

func (s *UsersPostgresStore) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	query := `
		SELECT id, username, email, password, created_at, updated_at 
		FROM users
		WHERE username = $1 AND is_active = true
		`
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	user := &User{}
	err := s.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password.hash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return user, nil
}
func (s *UsersPostgresStore) UpdateUser(ctx context.Context, user *User) error {
	query := `
		UPDATE users
		SET username = $1, email = $2, updated_at = $4
		WHERE id = $3
		`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	now := time.Now()
	err := s.db.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.Email,
		user.ID,
		now,
	).Scan()
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrNotFound
		default:
			return err
		}
	}
	return nil
}

func (s *UsersPostgresStore) DeleteUser(ctx context.Context, id uuid.UUID) error {
	query := `
		DELETE FROM users WHERE id = $1
		`
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	res, err := s.db.ExecContext(ctx, query, id)

	rows, err := res.RowsAffected()
	if err != nil {
		return errors.New("failed to get affected rows")
	}

	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *UsersPostgresStore) createUserInvitation(ctx context.Context, tx *sql.Tx, token string, exp time.Duration, userID uuid.UUID) error {
	query := `
		INSERT INTO user_invitations (token, id, expiry)
		VALUES ($1, $2, $3)
		`
	ctx, cancel := context.WithTimeout(ctx, exp)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, token, userID, time.Now().Add(exp))
	if err != nil {
		return err
	}
	return nil
}

func (s *UsersPostgresStore) getUserFromInvitation(ctx context.Context, tx *sql.Tx, token string) (*User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.created_at, u.is_active
		FROM users u 
		JOIN user_invitations ui ON u.id = ui.id
		WHERE ui.token = $1 and ui.expiry > $2
		`
	
	hash := sha256.Sum256([]byte(token))
	hashToken := hex.EncodeToString(hash[:])

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	user := &User{}
	err := tx.QueryRowContext(ctx, query, hashToken, time.Now()).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
		&user.IsActive,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return user, nil
}

// Updates the user to being active after e-mail
func (s *UsersPostgresStore) update(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `
		UPDATE users SET username = $1, email = $2, is_active = $3
		WHERE id = $4
		`
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, user.Username, user.Email, user.IsActive, user.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *UsersPostgresStore) deleteUserInvitation(ctx context.Context, tx *sql.Tx, id uuid.UUID) error {
	query := `
		DELETE FROM user_invitations WHERE id = $1
		`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	return nil
}
