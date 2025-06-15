package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ITine-Tech/blog/internal/store"

	_ "github.com/ITine-Tech/blog/docs"
	"github.com/go-chi/chi/v5"
)

type postKey string

const postCtx postKey = "post"

type CreatePost struct {
	Title string   `json:"title"`
	Text  string   `json:"text"`
	Tags  []string `json:"tags"`
}

type UpdatePostPayload struct {
	Title *string `json:"title" //validate:"omitempty,max=100"`
	Text  *string `json:"text" //validate:"omitempty,max=1000"`
}

// CreatePosts godoc
//
//	@Summary		Create a post
//	@Description	Creates a new post
//	@Tags			Posts
//	@Accept			json
//	@Produce		json
//	@Param			payload body		CreatePost true	"postPayload"
//	@Success		200		{object}	store.Post
//	@Failure		500		{object}	error	"Internal Server Error"
//	@Security		ApiKeyAuth
//	@Router			/posts [post]
func (app *application) CreatePostsHandler(w http.ResponseWriter, r *http.Request) {
	var postPayload CreatePost
	if err := readJSON(w, r, &postPayload); err != nil {
		app.jsonResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if postPayload.Title == "" || postPayload.Text == "" || len(postPayload.Tags) == 0 {
		app.badRequestResponse(w, r, fmt.Errorf("title, text, and tags are required"))
		return
	}

	user := getUserFromCtx(r)

	post := &store.Post{
		Title:  postPayload.Title,
		Text:   postPayload.Text,
		UserID: user.ID,
		Tags:   postPayload.Tags,
	}

	ctx := r.Context()

	if err := app.store.Posts.CreatePost(ctx, post); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, post); err != nil {
		app.badRequestResponse(w, r, err)
	}
}

// GetAllPosts godoc
//
//	@Summary		Get all posts by ID
//	@Description	Get all posts by ID
//	@Tags			Feed
//	@Accept			json
//	@Produce		json
//	@Success		200		{object}	store.Post
//	@Failure		500		{object}	error	"Internal Server Error"
//	@Router			/feed [get]
func (app *application) getAllPostsHandler(w http.ResponseWriter, r *http.Request) {
	posts, err := app.store.Posts.GetAllPosts(r.Context())
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, posts); err != nil {

		app.internalServerError(w, r, err)
	}
}

// GetPostByID godoc
//
//	@Summary		Get a post by ID
//	@Description	Get a post by ID
//	@Tags			Feed
//	@Accept			json
//	@Produce		json
//	@Param postID path int true "Post ID" regexp(^[0-9]+$)
//	@Success		200		{object}	store.Post
//	@Failure		404		{object}	error	"Not found"
//	@Failure		500		{object}	error	"Internal Server Error"
//	@Router			/feed/{postID} [get]
func (app *application) getPostByIDHandler(w http.ResponseWriter, r *http.Request) {
	strID := chi.URLParam(r, "postID")

	id, err := strconv.ParseInt(strID, 10, 64)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	ctx := r.Context()

	post, err := app.store.Posts.GetPostByID(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
	}

	comments, err := app.store.Comments.GetByPostID(ctx, id)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	post.Comments = comments

	if err := app.jsonResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// UpdatePostByID godoc
//
//	@Summary		Updates a post by ID
//	@Description	Updates a post by ID
//	@Tags			Posts
//	@Accept			json
//	@Produce		json
//	@Param postID path int true "Post ID" regexp(^[0-9]+$)
//	@Param			payload body		UpdatePostPayload true	"payload"
//	@Success		200		{object}	store.Post
//	@Failure		400		{object}	error	"Bad Request"
//	@Failure		500		{object}	error	"Internal Server Error"
//	@Security		ApiKeyAuth
//	@Router			/posts/{postID} [patch]
func (app *application) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)

	var payload UpdatePostPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if payload.Title != nil {
		post.Title = *payload.Title
	}
	if payload.Text != nil {
		post.Text = *payload.Text
	}

	if err := app.store.Posts.UpdatePost(r.Context(), post); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// DeletePost godoc
//
//	@Summary		Deletes a post
//	@Description	Deletes a post
//	@Tags			Posts
//	@Accept			json
//	@Produce		json
//	@Param postID path int true "Post ID" regexp(^[0-9]+$)
//	@Success		200		{object}	store.Post
//	@Failure		404		{object}	error	"Not found"
//	@Failure		500		{object}	error	"Internal Server Error"
//	@Security		ApiKeyAuth
//	@Router			/posts/{postID} [delete]
func (app *application) DeletePostHandler(w http.ResponseWriter, r *http.Request) {
	strID := chi.URLParam(r, "postID")
	postID, err := strconv.ParseInt(strID, 10, 64)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	ctx := r.Context()
	err = app.store.Posts.DeletePost(ctx, postID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (app *application) PostsContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		strID := chi.URLParam(r, "postID")
		postID, err := strconv.ParseInt(strID, 10, 64)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		ctx := r.Context()

		post, err := app.store.Posts.GetPostByID(ctx, postID)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):

				app.notFoundResponse(w, r, err)
			default:

				app.internalServerError(w, r, err)
			}
			return
		}
		ctx = context.WithValue(ctx, postCtx, post)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getPostFromCtx(r *http.Request) *store.Post {
	post, _ := r.Context().Value(postCtx).(*store.Post)
	return post
}
