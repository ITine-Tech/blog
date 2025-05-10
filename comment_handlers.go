package main

import (
	"berta2/store"
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

type CreateComment struct {
	Content string `json:"content"`
}

// CreateCommentsHandler godoc
//
//	@Summary		Create a comment
//	@Description	Creates a new comment
//	@Tags			Comments
//	@Accept			json
//	@Produce		json
//	@Param postID path int true "Post ID" regexp(^[0-9]+$)
//	@Param			payload body		CreateComment true	"commentsPayload"#
//	@Success		200		{object}	store.Comment
//	@Failure		500		{object}	error	"Internal Server Error"
//	@Security		ApiKeyAuth
//	@Router			/posts/comments/{postID} [post]
func (app *application) CreateCommentsHandler(w http.ResponseWriter, r *http.Request) {
	strID := chi.URLParam(r, "postID")

	postID, err := strconv.ParseInt(strID, 10, 64)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	var commentsPayload CreateComment
	if err := readJSON(w, r, &commentsPayload); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	//Validating input
	if commentsPayload.Content == "" {
		app.badRequestResponse(w, r, fmt.Errorf("some text is required"))
		return
	}

	user := getUserFromCtx(r)

	comment := &store.Comment{
		Content: commentsPayload.Content,
		UserID:  user.ID,
		PostID:  int(postID),
	}

	ctx := r.Context()

	if err := app.store.Comments.CreateComment(ctx, comment); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := writeJSON(w, http.StatusCreated, comment); err != nil {
		app.badRequestResponse(w, r, err)
	}
}
