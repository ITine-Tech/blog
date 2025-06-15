package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ITine-Tech/blog/docs"
	"github.com/ITine-Tech/blog/internal/auth"
	store2 "github.com/ITine-Tech/blog/internal/store"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

type application struct {
	config        config
	store         store2.Storage
	authenticator auth.Authenticator
}

type config struct {
	addr   string
	db     dbConfig
	apiURL string
	mail   mailConfig
	auth   authConfig
}

type authConfig struct {
	basic basicConfig
	token tokenConfig
}

type basicConfig struct {
	username string
	pass     string
}

type tokenConfig struct {
	secret   string
	expiry   time.Duration
	issuer   string
	audience string
}

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

type mailConfig struct {
	exp time.Duration
}

// mount sets up the HTTP router and middleware for the application.
func (app *application) mount() http.Handler {

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(60 * time.Second))

	r.Use(middleware.Logger)

	// Apply basic authentication middleware to the /healthcheck route
	r.With(app.basicAuthMiddleware()).Get("/healthcheck", app.healthCheck)

	// Serve Swagger documentation at /swagger/*
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:3000/swagger/doc.json")))

	r.Get("/feed", app.getAllPostsHandler)
	r.Get("/feed/{postID}", app.getPostByIDHandler)

	r.Route("/posts", func(r chi.Router) {
		r.Use(app.AuthTokenMiddleware)
		r.Post("/", app.CreatePostsHandler)
		r.Post("/comments/{postID}", app.CreateCommentsHandler)
		r.Route("/{postID}", func(r chi.Router) {
			r.Use(app.PostsContextMiddleware)
			r.Patch("/", app.checkPostOwnership("admin", app.updatePostHandler))
			r.Delete("/", app.checkPostOwnership("admin", app.DeletePostHandler))
		})
	})

	r.Route("/authentication", func(r chi.Router) {
		r.Post("/user", app.registerUserHandler)
		r.Post("/token", app.createTokenHandler)
	})

	r.Put("/users/activate/{token}", app.activateUserHandler)

	r.Route("/users", func(r chi.Router) {
		r.Use(app.AuthTokenMiddleware)
		r.Get("/", app.getUsersHandler)
		r.Route("/{userID}", func(r chi.Router) {
			r.Use(app.userContextMiddleware)
			r.Get("/", app.getUserByIDHandler)
			r.Patch("/", app.updateUserHandler)
			r.Delete("/", app.deleteUserHandler)
		})
	})

	return r
}

// run starts the HTTP server and listens for incoming connections.
// It initializes the Swagger documentation variables dynamically and sets up the server configuration.
//
// Parameters:
// - mux: An http.Handler that handles incoming requests.
//
// Returns:
// - An error if the server fails to start listening for connections.
func (app *application) run(mux http.Handler) error {
	docs.SwaggerInfo.Version = version
	docs.SwaggerInfo.Host = app.config.apiURL
	docs.SwaggerInfo.BasePath = "/"

	server := &http.Server{
		Addr:    app.config.addr,
		Handler: mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout: time.Minute,
	}

	fmt.Println("Starting the server on", app.config.addr)
	return server.ListenAndServe()
}
