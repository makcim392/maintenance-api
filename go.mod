module github.com/makcim392/maintenance-api

go 1.22

toolchain go1.23.3

// replace github.com/makcim392/swordhealth-interviewer => ./

require (
	github.com/DATA-DOG/go-sqlmock v1.5.2
	github.com/go-sql-driver/mysql v1.8.1
	github.com/golang-jwt/jwt/v4 v4.5.1
	github.com/google/uuid v1.6.0
	github.com/gorilla/mux v1.8.1
	github.com/joho/godotenv v1.5.1
	github.com/prometheus/client_golang v1.19.0
	github.com/stretchr/testify v1.10.0
	golang.org/x/crypto v0.31.0
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.8.1 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
