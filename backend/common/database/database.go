package database

import (
	"backend/common/config"
	"backend/generate/psql"
	"backend/type/common"
	"database/sql"
	"github.com/bsthun/gut"
	_ "github.com/lib/pq"
)

func Init(config *config.Config) common.Database {
	// * initialize postgres database
	postgres, err := sql.Open("postgres", *config.PostgresDsn)
	if err != nil {
		gut.Fatal("unable to connect to postgres database", err)
	}

	if err = postgres.Ping(); err != nil {
		gut.Fatal("unable to ping database", err)
	}

	// * wrap the database with query logging
	postgresQuerier := psql.New(&Wrapper{
		db: postgres,
	})

	// * construct struct
	database := &Database{
		PQuerier: postgresQuerier,
		PConn:    postgres,
	}

	return database
}
