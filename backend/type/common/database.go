package common

import (
	"backend/generate/psql"
	"context"
	"database/sql"
)

type Database interface {
	P() psql.PQuerier
	Ptx(context context.Context, opts *sql.TxOptions) (Tx, psql.PQuerier)
}

type Tx interface {
	Commit() error
	Rollback() error
}
