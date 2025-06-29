package adminEndpoint

import (
	"backend/generate/psql"
	"backend/type/common"
	"backend/type/payload"
	"backend/type/response"
	"github.com/bsthun/gut"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

func (r *Handler) HandleUserList(c *fiber.Ctx) error {
	// * get user claims
	_ = c.Locals("l").(*jwt.Token).Claims.(*common.LoginClaims)

	// * list all users
	users, err := r.database.P().UserList(c.Context())
	if err != nil {
		return gut.Err(false, "failed to list users", err)
	}

	// * map to response
	userItems, _ := gut.Iterate(users, func(user psql.User) (*payload.UserListItem, *gut.ErrorInstance) {
		return &payload.UserListItem{
			Id:        user.Id,
			Oid:       user.Oid,
			Firstname: user.Firstname,
			Lastname:  user.Lastname,
			Email:     user.Email,
			PhotoUrl:  user.PhotoUrl,
			IsAdmin:   user.IsAdmin,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}, nil
	})

	// * response
	return c.JSON(response.Success(c, &payload.UserListResponse{
		Users: userItems,
	}))
}
