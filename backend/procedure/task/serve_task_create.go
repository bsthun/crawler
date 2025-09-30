package taskProcedure

import (
	"backend/generate/psql"
	"context"

	"github.com/bsthun/gut"
)

func (r *Service) TaskCreate(ctx context.Context, querier psql.PQuerier, userId *uint64, uploadId *uint64, categoryName *string, taskType *string, source *string) (*psql.Task, *gut.ErrorInstance) {
	// * get category by name
	category, err := querier.CategoryGetByName(ctx, categoryName)
	if err != nil {
		return nil, gut.Err(false, "category not found", err)
	}

	// * create task
	task, err := querier.TaskCreateForUserId(ctx, &psql.TaskCreateForUserIdParams{
		UserId:     userId,
		UploadId:   uploadId,
		CategoryId: category.Id,
		Type:       taskType,
		Source:     source,
		IsRaw:      gut.Ptr(false),
		Title:      nil,
		Content:    nil,
	})
	if err != nil {
		return nil, gut.Err(false, "failed to create task", err)
	}

	return &task, nil
}
