package main

import (
	"log"
	"os"
	"time"

	_ "github.com/ITine-Tech/blog/docs"
	"github.com/ITine-Tech/blog/internal/auth"
	"github.com/ITine-Tech/blog/internal/db"
	"github.com/ITine-Tech/blog/internal/store"

	"github.com/joho/godotenv"
)

const version = "0.0.1"

//	@title			Beautiful Blog
//	@description	API for a blog
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	Christine Gundel
//	@contact.email	frau.gundi@outlook.com

// @license.name				Apache 2.0
// @license.url				http://www.apache.org/licenses/LICENSE-2.0.html
// @BasePath					/
//
// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
// @description
func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Warning: .env file not found or could not be loaded")
	}
	cfg := config{
		addr:   os.Getenv("LOCALHOST_ADDR"),
		apiURL: os.Getenv("API_URL"),
		db: dbConfig{
			addr:         os.Getenv("DB_CONN_STRING"),
			maxOpenConns: 30,
			maxIdleConns: 30,
			maxIdleTime:  "15m",
		},
		mail: mailConfig{
			exp: time.Hour * 24 * 3,
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
