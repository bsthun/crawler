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

func (r *Handler) HandleTaskUploadList(c *fiber.Ctx) error {
	// * login claims
	l := c.Locals("l").(*jwt.Token).Claims.(*common.LoginClaims)

	// * list uploads
	uploads, err := r.database.P().UploadListByUserId(c.Context(), l.UserId)
	if err != nil {
		return gut.Err(false, "failed to list uploads", err)
	}

	// * map to response
	uploadItems, _ := gut.Iterate(uploads, func(upload psql.Upload) (*payload.TaskUploadItem, *gut.ErrorInstance) {
		return &payload.TaskUploadItem{
			Id:        upload.Id,
			UserId:    upload.UserId,
			CreatedAt: upload.CreatedAt,
			UpdatedAt: upload.UpdatedAt,
		}, nil
	})

	// * response
	return c.JSON(response.Success(c, &payload.TaskUploadListResponse{
		Uploads: uploadItems,
	}))
}
