package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	ID        int       `json:"id"`
	PostID    int       `json:"post_id"`
	UserID    uuid.UUID `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	User      User      `json:"user"`
}

type CommentsPostgreStore struct {
	db *sql.DB
}

func (s *CommentsPostgreStore) GetByPostID(ctx context.Context, postId int64) ([]Comment, error) {
	query := `
		SELECT c.id, c.post_id, c.user_id, c.content, c.created_at, users.username, users.id FROM comments c
		JOIN users on users.id = c.user_id 
		WHERE c.post_id = $1
		ORDER BY c.created_at DESC;
	`
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, postId)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	comments := []Comment{}

	for rows.Next() {
		var comment Comment
		comment.User = User{}

		err := rows.Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Content, &comment.CreatedAt, &comment.User.Username, &comment.User.ID)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)

	}
	return comments, nil
}

func (s *CommentsPostgreStore) CreateComment(ctx context.Context, comment *Comment) error {
	query := `
        WITH post_exists AS (
            SELECT EXISTS(SELECT 1 FROM posts WHERE id = $1) AS exists
        )
        INSERT INTO comments(post_id, user_id, content)
        SELECT $1, $2, $3
        FROM post_exists
        WHERE exists = TRUE
        RETURNING id, created_at
    `
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := s.db.QueryRowContext(
		ctx,
		query,
		comment.PostID,
		comment.UserID,
		comment.Content,
	).Scan(
		&comment.ID,
		&comment.CreatedAt,
	)
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
