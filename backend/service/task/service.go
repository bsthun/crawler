package taskProcedure

import (
	"backend/generate/psql"
	"backend/type/common"
	"context"
	"github.com/bsthun/gut"
)

type Server interface {
	TaskCreate(ctx context.Context, userId *uint64, categoryName *string, taskType *string, url *string) (*psql.Task, *gut.ErrorInstance)
}

type Service struct {
	database common.Database
}

func Serve(database common.Database) Server {
	return &Service{
		database: database,
	}
}
