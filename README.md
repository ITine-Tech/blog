# Blog
## A backend API for a blog

This is a blog built with Go and PostgreSQL. It aims to provide a platform to share thoughts and experiences.

## Getting Started

1. Install Go: Download and install the latest version of Go from https://golang.org/dl/.

2. Start PostgreSQL using Docker Compose:

Run the following command in the project directory:

```sh
docker compose up -d
```

This will start a PostgreSQL instance as defined in the `docker-compose.yml` file.


## Database Migrations

To manage database schema changes, this project uses the [migrate](https://github.com/golang-migrate/migrate) tool.

### Installing migrate

You can install the migrate CLI tool using Homebrew (macOS/Linux):

```sh
brew install golang-migrate
```

Or download a pre-built binary from the [releases page](https://github.com/golang-migrate/migrate/releases).

### Running Migrations

To apply migrations to your PostgreSQL database, run:

```sh
migrate -path ./migrations -database "postgres://postgres:<your_password>@localhost/<your_database_name>" up
```

Replace password and database name.

## API Documentation (Swagger)

This project uses Swagger for interactive API documentation.

Once the application is running, you can access the Swagger UI at:

[http://localhost:3000/swagger/index.html](http://localhost:3000/swagger/index.html)

Use this interface to explore available endpoints, view request/response formats, and test API operations.

For more information about Swagger, see the [Swagger Documentation](https://swagger.io/docs/).


## Contributing
Contributions are welcome! If you find any bugs or have suggestions for improvements, please open an issue or submit a pull request.