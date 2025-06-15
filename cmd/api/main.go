package main

import (
	"log"
	"os"
	"time"

	"berta2/internal/auth"
	"berta2/internal/db"
	_ "berta2/docs"
	"berta2/internal/store"

	"github.com/joho/godotenv"
)

const version = "0.0.1"

//	@title			Berta's Beautiful Blog
//	@description	API for Berta's blog
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	Christine Gundel
//	@contact.email	christine.gundel@mail.schwarz

// @license.name				Apache 2.0
// @license.url				http://www.apache.org/licenses/LICENSE-2.0.html
// @BasePath					/
//
// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
// @description
func main() {
	godotenv.Load()
	cfg := config{
		addr:   os.Getenv("LOCALHOST_ADDR"),
		apiURL: os.Getenv(" API_URL"),
		db: dbConfig{
			addr:         os.Getenv("DB_CONN_STRING"),
			maxOpenConns: 30, //This can all be done in the ENV (Chapter 19, 3:35)
			maxIdleConns: 30,
			maxIdleTime:  "15m",
		},
		mail: mailConfig{
			exp: time.Hour * 24 * 3, // 3 days to accept invitation
		},
		auth: authConfig{
			basic: basicConfig{
				username: os.Getenv("ADMIN_NAME"),
				pass:     os.Getenv("ADMIN_PASSWORD"),
			},
			token: tokenConfig{
				secret: os.Getenv("TOKEN_SECRET"),
				expiry: time.Hour * 24 * 3,
				issuer: os.Getenv("TOKEN_ISSUER"),
			},
		},
	}

	db, err := db.NewDB(
		cfg.db.addr,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxIdleTime,
	)
	if err != nil {
		log.Panic(err)
	}

	defer db.Close()
	log.Println("Database connection established")

	myStore := store.NewPostgresStorage(db)

	tokenHost := os.Getenv("TOKEN_AUDIENCE")
	JWTAuthenticator := auth.NewJWTAuthenticator(cfg.auth.token.secret, tokenHost, cfg.auth.token.issuer)

	app := &application{
		config:        cfg,
		store:         myStore,
		authenticator: JWTAuthenticator,
	}

	mux := app.mount()
	log.Fatal(app.run(mux))
}
