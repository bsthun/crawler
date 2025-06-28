package stateEndpoint

import (
	"backend/generate/psql"
	"backend/type/common"
	"backend/type/payload"
	"backend/type/response"
	"github.com/bsthun/gut"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

func (r *Handler) HandleStateOverview(c *fiber.Ctx) error {
	// * login claims
	l := c.Locals("l").(*jwt.Token).Claims.(*common.LoginClaims)

	// * get overview statistics
	statsRows, err := r.database.P().TaskOverviewByUserId(c.Context(), l.UserId)
	if err != nil {
		return gut.Err(false, "failed to get overview statistics", err)
	}

	// * get pool token
	categoryRows, err := r.database.P().PoolTokenOverviewByCategory(c.Context())
	if err != nil {
		return gut.Err(false, "failed to get pool token overview by category", err)
	}

	// * map daily stats to histories array
	histories, _ := gut.Iterate(statsRows, func(row psql.TaskOverviewByUserIdRow) (*payload.OverviewHistoryItem, *gut.ErrorInstance) {
		return &payload.OverviewHistoryItem{
			Submitted: row.Submitted,
			Pending:   row.Pending,
			Completed: row.Completed,
			Failed:    row.Failed,
		}, nil
	})

	// * map category stats to pool token array
	poolTokens, _ := gut.Iterate(categoryRows, func(row psql.PoolTokenOverviewByCategoryRow) (*payload.PoolTokenCategoryItem, *gut.ErrorInstance) {
		return &payload.PoolTokenCategoryItem{
			CategoryId:   row.CategoryId,
			CategoryName: row.CategoryName,
			TokenCount:   row.TokenCount,
		}, nil
	})

	// * response
	return c.JSON(response.Success(c, &payload.Overview{
		TokenCount: statsRows[0].TokenCount,
		Histories:  histories,
		PoolTokens: poolTokens,
	}))
}
