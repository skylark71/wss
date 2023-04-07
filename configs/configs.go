package configs

import (
	"os"

	"github.com/go-pg/pg/v10"
)

var Port = os.Getenv("POSTGRES_DSN_HOST")
var User = os.Getenv("POSTGRES_DSN_USER")
var Password = os.Getenv("POSTGRES_DSN_PASSWORD")
var DBname = os.Getenv("POSTGRES_DSN_DB")

var (
	Db *pg.DB
)
