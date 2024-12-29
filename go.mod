module github.com/makcim392/swordhealth-interviewer

go 1.21

replace github.com/makcim392/swordhealth-interviewer => ./

require (
	github.com/go-sql-driver/mysql v1.8.1
	github.com/google/uuid v1.6.0
	github.com/gorilla/mux v1.8.1
)

require filippo.io/edwards25519 v1.1.0 // indirect
