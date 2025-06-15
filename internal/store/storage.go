package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

var (
	ErrNotFound = errors.New("record not found")
)

type Posts interface {
	CreatePost(context.Context, *Post) error
	GetAllPosts(context.Context) ([]*Post, error)
	GetPostByID(context.Context, int64) (*Post, error)
	UpdatePost(context.Context, *Post) error
	DeletePost(context.Context, int64) error
}

type Users interface {
	Create(context.Context, *sql.Tx, *User) error
	GetAllUsers(context.Context) ([]*User, error)
	CreateAndInvite(context.Context, *User, string, time.Duration) error
	Activate(context.Context, string) error
	GetUserByID(context.Context, uuid.UUID) (*User, error)
	GetUserByUsername(context.Context, string) (*User, error)
	UpdateUser(context.Context, *User) error
	DeleteUser(context.Context, uuid.UUID) error
}

type Roles interface {
	GetByName(context.Context, string) (*Role, error)
}

type Comments interface {
	GetByPostID(context.Context, int64) ([]Comment, error)
	CreateComment(context.Context, *Comment) error
}

type Storage struct {
	Posts    Posts
	Users    Users
	Comments Comments
	Roles    Roles
}

func NewPostgresStorage(db *sql.DB) Storage {
	return Storage{
		Posts:    &PostsPostgreStore{db},
		Users:    &UsersPostgresStore{db},
		Comments: &CommentsPostgreStore{db},
		Roles:    &RolePostgreStore{db},
	}
}

// withTx is a wrapper function that creates a transaction for the provided database.
// It takes a context, a database connection, and a function as parameters.
// The function is executed within the transaction. If the function returns an error,
// the transaction is rolled back. Otherwise, the transaction is committed.
//
// Parameters:
// - db: A pointer to a sql.DB object representing the database connection.
// - ctx: A context.Context object that provides a deadline, cancellation signal, and other options for the transaction.
// - fn: A function that takes a *sql.Tx as a parameter and returns an error. This function represents the work to be done within the transaction.
//
// Return value:
//   - An error if any error occurs during the transaction or if the provided function returns an error.
//     If the transaction is successfully committed, nil is returned.
func withTx(db *sql.DB, ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
