package main

import (
	"berta2/store"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"net/http"
	"time"
)

type RegisterUserPayload struct {
	Username string `json:"username" validate:"required,max=100"`
	Email    string `json:"email" validate:"required,max=255"`
	Password string `json:"password" validate:"required,min=5,max=50"`
}

type UserWithToken struct {
	*store.User
	Token string `json:"token"`
}

type CreateUserTokenPayload struct {
	Username string `json:"username" validate:"required,max=100"`
	Password string `json:"password" validate:"required,min=5,max=50"`
}

// registerUserHandler godoc
//
// @Summary	Register a user
// @Description Register a user
// @Tags Authentication
// @Accept	json
// @Produce	json
// @Param	payload body	RegisterUserPayload true "userPayload"
// @Success 201		{object} UserWithToken	"User registered"
// @Failure	400		{object} error	"Bad Request"
// @Failure 500		{object} error	"Internal Server Error"
// @Router	/authentication/user [post]
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var userPayload RegisterUserPayload
	if err := readJSON(w, r, &userPayload); err != nil {
		app.badRequestResponse(w, r, err)
	}

	//Validating input
	if userPayload.Username == "" || userPayload.Email == "" || userPayload.Password == "" {
		app.badRequestResponse(w, r, fmt.Errorf("username, email, and password are required"))
		return
	}

	user := &store.User{
		Username: userPayload.Username,
		Email:    userPayload.Email,
		Role: store.Role{
			Name: "user",
		},
	}

	//Hash the password
	if err := user.Password.Set(userPayload.Password); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	ctx := r.Context()

	newToken := uuid.New().String()

	//store the token in db
	hash := sha256.Sum256([]byte(newToken))
	hashToken := hex.EncodeToString(hash[:])

	err := app.store.Users.CreateAndInvite(ctx, user, hashToken, app.config.mail.exp)
	if err != nil {
		switch err {
		case store.ErrDuplicateUsername:
			app.badRequestResponse(w, r, err)
		case store.ErrDuplicateEmail:
			app.badRequestResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	//send e-mail to user here
	//The struct here is used to send payload so token can be used in Bruno
	userWithToken := UserWithToken{
		User:  user,
		Token: newToken,
	}

	if err := app.jsonResponse(w, http.StatusCreated, userWithToken); err != nil {
		app.internalServerError(w, r, err)
	}
}

// createTokenHandler godoc
//
// @Summary	creates a token
// @Description creates a token for a user
// @Tags Authentication
// @Accept json
// @Produce json
// @Param payload body	CreateUserTokenPayload true "User credentials"
// @Success 201 {object} string "Token"
// @Failure 400 {object}	error
// @Failure 401 {object}	error
// @Failure 500 {object}	error "Internal Server Error"
// @Router /authentication/token [post]
func (app *application) createTokenHandler(w http.ResponseWriter, r *http.Request) {
	//parse payload credentials
	var userPayload CreateUserTokenPayload
	if err := readJSON(w, r, &userPayload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	//Validating input
	if userPayload.Username == "" || userPayload.Password == "" {
		app.badRequestResponse(w, r, fmt.Errorf("username and password are required"))
		return
	}

	//fetch the user (check if user exists) from the payload
	user, err := app.store.Users.GetUserByUsername(r.Context(), userPayload.Username)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.unauthorizedResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := user.Password.Compare(userPayload.Password); err != nil {
		app.unauthorizedResponse(w, r, err)
		return
	}

	//generate the token --> add claims
	claims := jwt.MapClaims{
		"sub": user.ID, //subject
		"exp": time.Now().Add(app.config.auth.token.expiry).Unix(),
		"iat": time.Now().Unix(), //Issued at
		"nbf": time.Now().Unix(), //Not before
		"iss": app.config.auth.token.issuer,
		"aud": app.config.auth.token.audience,
	}
	token, err := app.authenticator.GenerateToken(claims)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	//send it to the client

	if err := app.jsonResponse(w, http.StatusCreated, token); err != nil {
		app.internalServerError(w, r, err)
	}
}
