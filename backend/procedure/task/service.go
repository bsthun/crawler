package taskProcedure

import (
	"backend/generate/psql"
	"backend/type/common"
	"context"
	"github.com/bsthun/gut"
)

type Server interface {
	TaskCreate(ctx context.Context, querier psql.PQuerier, userId *uint64, uploadId *uint64, categoryName *string, taskType *string, source *string) (*psql.Task, *gut.ErrorInstance)
	TaskRawCreate(ctx context.Context, querier psql.PQuerier, userId *uint64, uploadId *uint64, categoryName *string, taskType *string, source *string, title *string, content *string) (*psql.Task, *gut.ErrorInstance)
}

type Service struct {
	database common.Database
}

func Serve(database common.Database) Server {
	return &Service{
		database: database,
	}
}
