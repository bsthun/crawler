package taskEndpoint

import (
	"backend/generate/psql"
	"backend/type/payload"
	"backend/type/response"
	"github.com/bsthun/gut"
	"github.com/gofiber/fiber/v2"
)

func (r *Handler) HandleTaskCategoryList(c *fiber.Ctx) error {
	// * list categories
	categories, err := r.database.P().CategoryList(c.Context())
	if err != nil {
		return gut.Err(false, "failed to list categories", err)
	}

	// * map to response
	categoryItems, _ := gut.Iterate(categories, func(category psql.Category) (*payload.TaskCategoryItem, *gut.ErrorInstance) {
		return &payload.TaskCategoryItem{
			Id:        category.Id,
			Name:      category.Name,
			CreatedAt: category.CreatedAt,
			UpdatedAt: category.UpdatedAt,
		}, nil
	})

	// * response
	return c.JSON(response.Success(c, &payload.TaskCategoryListResponse{
		Categories: categoryItems,
	}))
}
