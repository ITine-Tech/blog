package main

import (
	"context"
	"errors"
	"net/http"

	_ "github.com/ITine-Tech/blog/docs"
	"github.com/ITine-Tech/blog/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type userKey string

const userCTx userKey = "userID"

type UpdateUserPayload struct {
	Username *string `json:"username" //validate:"omitempty,max=100"`
	Email    *string `json:"email" //validate:"omitempty,max=100"`
	//Password *string `json:"password" //validate:"omitempty"`
}

// ActivateUser godoc
//
// @Summary	Activates/registers a user
// @Description Activates/registers a user by invitation token
// @Tags Users
// @Produce json
// @Param token path string true "Invitation token"
// @Success 204 {string}	string "User activated"
// @Failure 404 {object} error
// @Failure 500 {object} error "Internal Server Error"
// @Security ApiKeyAuth
// @Router 	/users/activate/{token} [put]
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	err := app.store.Users.Activate(r.Context(), token)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}
	if err := app.jsonResponse(w, http.StatusNoContent, "User activated"); err != nil {
		app.internalServerError(w, r, err)
	}
}

// GetAllUsers godoc
//
//	@Summary		Fetches all user profiles
//	@Description	Fetches all user profiles
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Success		200		{object}	store.User
//	@Failure		500		{object}	error	"Internal Server Error"
//	@Security		ApiKeyAuth
//	@Router			/users [get]
func (app *application) getUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := app.store.Users.GetAllUsers(r.Context())
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.notFoundResponse(w, r, err)
			return
		default:
			app.jsonResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	if err := app.jsonResponse(w, http.StatusOK, users); err != nil {
		app.internalServerError(w, r, err)
	}

}

// GetUserByID godoc
//
//	@Summary		Fetches a user profile by ID
//	@Description	Fetches a user profile by ID
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		string	true	"User ID"
//	@Success		200		{object}	store.User
//	@Failure		400		{object}	error	"Bad Request"
//	@Failure		404		{object}	error	"Not Found"
//	@Failure		500		{object}	error	"Internal Server Error"
//	@Security		ApiKeyAuth
//	@Router			/users/{userID} [get]
//
// getUserByIDHandler fetches a user profile by ID from the request context.
// It responds with a JSON representation of the user profile if found,
// or an appropriate error response if not found or an internal server error occurs.
//
// Parameters:
// - w: http.ResponseWriter to write the response.
// - r: *http.Request containing the request data.
//
// Returns:
// - No explicit return value. Writes the response directly to the http.ResponseWriter.
func (app *application) getUserByIDHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromCtx(r)

	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// UpdateUser godoc
//
//	@Summary		Updates a user profile by ID
//	@Description	Updates a user profile by ID
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		string	true	"User ID"
//	@Param			payload body		UpdateUserPayload true	"payload"
//	@Success		200		{object}	store.User
//	@Failure		400		{object}	error	"Bad Request"
//	@Failure		500		{object}	error	"Internal Server Error"
//	@Security		ApiKeyAuth
//	@Router			/users/{userID} [patch]
func (app *application) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromCtx(r)

	var payload UpdateUserPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if payload.Username != nil {
		user.Username = *payload.Username
	}
	if payload.Email != nil {
		user.Email = *payload.Email
	}

	if err := app.store.Users.UpdateUser(r.Context(), user); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
		return
	}

}

// DeleteUser godoc
//
//	@Summary		Deletes a user profile
//	@Description	Deletes a user profile
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		string	true	"User ID"
//	@Failure		404		{object}	error	"Not found"
//	@Failure		500		{object}	error	"Internal Server Error"
//	@Security		ApiKeyAuth
//	@Router			/users/{userID} [delete]
func (app *application) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromCtx(r)

	if err := app.store.Users.DeleteUser(r.Context(), user.ID); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}
	w.WriteHeader(http.StatusNotFound)
}

func (app *application) userContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		strID := chi.URLParam(r, "userID")
		userID, err := uuid.Parse(strID)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		ctx := r.Context()

		user, err := app.store.Users.GetUserByID(ctx, userID)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):

				app.notFoundResponse(w, r, err)
			default:

				app.internalServerError(w, r, err)
			}
			return
		}
		ctx = context.WithValue(ctx, userCTx, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// This function makes it easier to get the context
func getUserFromCtx(r *http.Request) *store.User {
	user, _ := r.Context().Value(userCTx).(*store.User)
	return user
}
