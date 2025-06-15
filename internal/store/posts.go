package store

//database handling

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq" //Database driver for Postgres
)

type Post struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Text      string    `json:"text"`
	UserID    uuid.UUID `json:"user_id"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Version   int       `json:"version"`
	Comments  []Comment `json:"comments"`
}

type PostsPostgreStore struct {
	db *sql.DB
}

func (s *PostsPostgreStore) CreatePost(ctx context.Context, post *Post) error {
	query := `
	INSERT INTO posts (title, text, user_id, tags)
	VALUES ($1, $2, $3, $4)
	RETURNING id, created_at, updated_at
	`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	err := s.db.QueryRowContext(
		ctx, //What does this context do?
		query,
		post.Title,
		post.Text,
		post.UserID,
		pq.Array(post.Tags),
	).Scan( //this Scan part is used for the automatically generated data types
		&post.ID,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *PostsPostgreStore) GetAllPosts(ctx context.Context) ([]*Post, error) {
	rows, err := s.db.Query(`SELECT * FROM posts`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var result []*Post

	for rows.Next() {
		post := &Post{}
		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Text,
			&post.UserID,
			pq.Array(&post.Tags),
			&post.CreatedAt,
			&post.UpdatedAt,
			&post.Version,
		)

		if err != nil {
			return nil, err
		}

		result = append(result, post)
	}

	return result, nil
}

func (s *PostsPostgreStore) GetPostByID(ctx context.Context, id int64) (*Post, error) {
	query := `
	SELECT id, title, text, user_id, tags, created_at, updated_at, version 
	FROM posts
	WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var post Post

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&post.ID,
		&post.Title,
		&post.Text,
		&post.UserID,
		pq.Array(&post.Tags),
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return &post, nil
}

func (s *PostsPostgreStore) UpdatePost(ctx context.Context, post *Post) error {
	query := `
    UPDATE posts
    SET title = $1, text = $2, updated_at = $3, version = version + 1
    WHERE id = $4 AND version = $5
    RETURNING version
`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	now := time.Now()
	err := s.db.QueryRowContext(
		ctx,
		query,
		post.Title,
		post.Text,
		now,
		post.ID,
		post.Version,
	).Scan(&post.Version)
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

func (s *PostsPostgreStore) DeletePost(ctx context.Context, PostId int64) error {
	query := `
DELETE FROM posts WHERE id = $1
`
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	//This is Ecec because I don't want to return anything
	res, err := s.db.ExecContext(ctx, query, PostId)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
