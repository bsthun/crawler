package taskProcedure

import (
	"backend/generate/psql"
	"context"
	"github.com/bsthun/gut"
)

func (r *Service) TaskRawCreate(ctx context.Context, userId *uint64, uploadId *uint64, categoryName *string, taskType *string, source *string, title *string, content *string) (*psql.Task, *gut.ErrorInstance) {
	// * get category by name
	category, err := r.database.P().CategoryGetByName(ctx, categoryName)
	if err != nil {
		return nil, gut.Err(false, "category not found", err)
	}

	// * create raw task with title and content
	task, err := r.database.P().TaskCreateForUserId(ctx, &psql.TaskCreateForUserIdParams{
		UserId:     userId,
		UploadId:   uploadId,
		CategoryId: category.Id,
		Type:       taskType,
		Source:     source,
		IsRaw:      gut.Ptr(true),
		Title:      title,
		Content:    content,
	})
	if err != nil {
		return nil, gut.Err(false, "failed to create raw task", err)
	}

	return &task, nil
}
