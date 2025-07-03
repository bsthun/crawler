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
	// * login claims
	l := c.Locals("l").(*jwt.Token).Claims.(*common.LoginClaims)

	// * parse body
	body := new(payload.TaskSubmitRequest)
	if err := c.BodyParser(body); err != nil {
		return gut.Err(false, "invalid body", err)
	}

	// * validate body
	if err := gut.Validate(body); err != nil {
		return err
	}

	// * create task
	task, er := r.taskProcedure.TaskCreate(c.Context(), r.database.P(), l.UserId, nil, body.Category, body.Type, body.Source)
	if er != nil {
		return er
	}

	// * response
	return c.JSON(response.Success(c, &payload.TaskSubmitResponse{
		TaskId: task.Id,
	}))
}
