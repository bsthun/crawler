package taskEndpoint

import (
	"backend/generate/psql"
	"backend/type/common"
	"backend/type/payload"
	"backend/type/response"
	"github.com/bsthun/gut"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

func (r *Handler) HandleTaskList(c *fiber.Ctx) error {
	// * login claims
	l := c.Locals("l").(*jwt.Token).Claims.(*common.LoginClaims)

	// * parse body
	body := new(payload.TaskListRequest)
	if err := c.BodyParser(body); err != nil {
		return gut.Err(false, "invalid body", err)
	}

	// * set default values
	if body.Limit == nil {
		body.Limit = gut.Ptr(int32(10))
	}
	if body.Offset == nil {
		body.Offset = gut.Ptr(int32(0))
	}

	// * validate uploadId ownership if provided
	if body.UploadId != nil {
		_, err := r.database.P().UploadGetByIdAndUserId(c.Context(), &psql.UploadGetByIdAndUserIdParams{
			Id:     body.UploadId,
			UserId: l.UserId,
		})
		if err != nil {
			return gut.Err(false, "upload not found or not owned by user", err)
		}
	}

	// * count tasks
	count, err := r.database.P().TaskCountByUserId(c.Context(), &psql.TaskCountByUserIdParams{
		UserId:   l.UserId,
		UploadId: body.UploadId,
	})
	if err != nil {
		return gut.Err(false, "failed to count tasks", err)
	}

	// * list tasks
	tasks, err := r.database.P().TaskListByUserId(c.Context(), &psql.TaskListByUserIdParams{
		UserId:   l.UserId,
		Limit:    body.Limit,
		Offset:   body.Offset,
		UploadId: body.UploadId,
	})
	if err != nil {
		return gut.Err(false, "failed to list tasks", err)
	}

	// * map to response
	taskItems, _ := gut.Iterate(tasks, func(task psql.TaskListByUserIdRow) (*payload.TaskListItem, *gut.ErrorInstance) {
		return &payload.TaskListItem{
			Id:           task.Id,
			UserId:       task.UserId,
			UploadId:     task.UploadId,
			CategoryId:   task.CategoryId,
			Type:         task.Type,
			Url:          task.Url,
			Status:       task.Status,
			FailedReason: task.FailedReason,
			TokenCount:   task.TokenCount,
			CreatedAt:    task.CreatedAt,
			UpdatedAt:    task.UpdatedAt,
		}, nil
	})

	// * response
	return c.JSON(response.Success(c, &payload.TaskListResponse{
		Count: count,
		Tasks: taskItems,
	}))
}
