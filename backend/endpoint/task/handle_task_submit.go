package taskEndpoint

import (
	"backend/type/common"
	"backend/type/payload"
	"backend/type/response"
	"github.com/bsthun/gut"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

func (r *Handler) HandleTaskSubmit(c *fiber.Ctx) error {
	// * parse request
	var req payload.TaskSubmitRequest
	if err := c.BodyParser(&req); err != nil {
		return gut.Err(false, "invalid request body", err)
	}

	// * validate request
	if err := gut.Validate(&req); err != nil {
		return gut.Err(false, "validation failed", err)
	}

	// * get user claims
	l := c.Locals("l").(*jwt.Token).Claims.(*common.LoginClaims)

	// * create task
	task, er := r.taskService.TaskCreate(c.Context(), l.UserId, req.Category, req.Type, req.Url)
	if er != nil {
		return er
	}

	// * response
	return c.JSON(response.Success(c, &payload.TaskSubmitResponse{
		TaskId: task.Id,
	}))
}
