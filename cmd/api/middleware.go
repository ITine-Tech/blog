package main

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	"github.com/ITine-Tech/blog/internal/store"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func (app *application) basicAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			//Read the auth header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				app.unauthorizedBasicErrorResponse(w, r, errors.New("missing Authorization header"))
				return
			}

			//parse it --> get base 64
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Basic" {
				app.unauthorizedBasicErrorResponse(w, r, errors.New("invalid Authorization header"))
				return
			}

			//decode it
			decoded, err := base64.StdEncoding.DecodeString(parts[1])
			if err != nil {
				app.unauthorizedBasicErrorResponse(w, r, errors.New("invalid Authorization header"))
				return
			}

			username := app.config.auth.basic.username
			pass := app.config.auth.basic.pass

			//check credentials
			credentials := strings.SplitN(string(decoded), ":", 2)
			if len(credentials) != 2 || credentials[0] != username || credentials[1] != pass {
				app.unauthorizedResponse(w, r, errors.New("invalid Authorization header"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (app *application) AuthTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			app.unauthorizedResponse(w, r, errors.New("missing Authorization header"))
			return
		}
		parts := strings.SplitN(authHeader, " ", 2) //authorization: Bearer <token>
		if len(parts) != 2 || parts[0] != "Bearer" {
			app.unauthorizedResponse(w, r, errors.New("invalid Authorization header"))
			return
		}

		token := parts[1]

		jwtToken, err := app.authenticator.ValidateToken(token)
		if err != nil {
			app.unauthorizedResponse(w, r, err)
			return
		}

		claims, _ := jwtToken.Claims.(jwt.MapClaims)

		userID, err := uuid.Parse(claims["sub"].(string))
		if err != nil {
			app.unauthorizedResponse(w, r, err)
			return
		}

		ctx := r.Context()
		user, err := app.store.Users.GetUserByID(ctx, userID)
		if err != nil {
			app.unauthorizedResponse(w, r, err)
		}

		ctx = context.WithValue(ctx, userCTx, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) checkPostOwnership(requiredRole string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := getUserFromCtx(r)
		post := getPostFromCtx(r)

		//if the post is the user's post
		if post.UserID == user.ID {
			next.ServeHTTP(w, r)
			return
		}

		allowedRole, err := app.checkRolePrecedence(r.Context(), user, requiredRole)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		if !allowedRole {
			app.forbiddenResponse(w, r, err)
			return
		}

		next.ServeHTTP(w, r)
	})

}

// Doesn't work yet: How can I get the user from the authentication?
/*func (app *application) checkUserOwnership(requiredRole string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := getUserFromCtx(r) //gets the user from the Auth

		userIDfromURL := chi.URLParam(r, "userID")
		userID, err := uuid.Parse(userIDfromURL)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		if user.ID == userID {
			next.ServeHTTP(w, r)
			return
		}

		allowedRole, err := app.checkRolePrecedence(r.Context(), user, requiredRole)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		if !allowedRole {
			app.forbiddenResponse(w, r, err)
			return
		}
		next.ServeHTTP(w, r)

	})
}
*/
func (app *application) checkRolePrecedence(ctx context.Context, user *store.User, roleName string) (bool, error) {
	role, err := app.store.Roles.GetByName(ctx, roleName)
	if err != nil {
		return false, err
	}
	// role level is higher
	return user.Role.Level >= role.Level, nil
}
