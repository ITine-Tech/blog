package main

import (
	"berta2/auth"
	"berta2/docs"
	store2 "berta2/store"
	"fmt"
	httpSwagger "github.com/swaggo/http-swagger"
	"net/http"
	"time"

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

// this is in a struct so it is not hardcoded, and can be changed easily
type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

type mailConfig struct {
	exp time.Duration
}

// This function sets up the ServeMux outside the run() method. This is better to use with testing purposes.
// mount sets up the HTTP request multiplexer (ServeMux) and returns it as an http.Handler.
// This function is responsible for defining the routes and middleware for the application.
func (app *application) mount() http.Handler {

	r := chi.NewRouter()

	// A good base middleware stack:
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	// Recover from panics
	r.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	// This makes the nice logging in the terminal
	r.Use(middleware.Logger)

	// Apply basic authentication middleware to the /healthcheck route
	r.With(app.basicAuthMiddleware()).Get("/healthcheck", app.healthCheck)

	// Serve Swagger documentation at /swagger/*
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:3000/swagger/doc.json")))

	// Define routes for retrieving and listing posts
	r.Get("/feed", app.getAllPostsHandler)
	r.Get("/feed/{postID}", app.getPostByIDHandler)

	// Define routes for managing posts
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

	// Define routes for user authentication
	r.Route("/authentication", func(r chi.Router) {
		r.Post("/user", app.registerUserHandler)
		r.Post("/token", app.createTokenHandler)
	})

	// Define route for activating user accounts
	r.Put("/users/activate/{token}", app.activateUserHandler)

	// Define routes for managing users
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
	// Docs: defining the Swagger doc variables dynamically
	docs.SwaggerInfo.Version = version
	docs.SwaggerInfo.Host = app.config.apiURL
	docs.SwaggerInfo.BasePath = "/"

	// mux := http.NewServeMux() --> moved to the mount() method

	server := &http.Server{
		Addr:    app.config.addr,
		Handler: mux,

		// This should always be added for server shutdown:
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		// Idle connections will be closed after this duration.
		IdleTimeout: time.Minute,
	}

	fmt.Println("Starting the server on", app.config.addr)
	return server.ListenAndServe()
}
