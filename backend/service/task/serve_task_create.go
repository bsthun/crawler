package taskProcedure

import (
	"backend/generate/psql"
	"context"
	"github.com/bsthun/gut"
)

func (r *Service) TaskCreate(ctx context.Context, userId *uint64, categoryName *string, taskType *string, url *string) (*psql.Task, *gut.ErrorInstance) {
	// * get category by name
	category, err := r.database.P().CategoryGetByName(ctx, categoryName)
	if err != nil {
		return nil, gut.Err(false, "category not found", err)
	}

	// * create task
	task, err := r.database.P().TaskCreateForUserId(ctx, &psql.TaskCreateForUserIdParams{
		UserId:     userId,
		CategoryId: category.Id,
		Type:       taskType,
		Url:        url,
	})
	if err != nil {
		return nil, gut.Err(false, "failed to create task", err)
	}

	return &task, nil
}
