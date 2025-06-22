package stateEndpoint

import (
	"backend/type/common"
	"backend/type/payload"
	"backend/type/response"
	"github.com/bsthun/gut"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

func (r *Handler) HandleState(c *fiber.Ctx) error {
	// * get user claims from jwt token
	u := c.Locals("l").(*jwt.Token).Claims.(*common.UserClaims)

	// * get user from database
	user, err := r.database.P().UserGetById(c.Context(), u.UserId)
	if err != nil {
		return gut.Err(false, "failed to get user", err)
	}

	// * response
	return c.JSON(response.Success(c, &payload.StateResponse{
		UserId:      user.Id,
		DisplayName: user.Firstname,
		Email:       user.Email,
		PhotoUrl:    user.PhotoUrl,
	}))
}
