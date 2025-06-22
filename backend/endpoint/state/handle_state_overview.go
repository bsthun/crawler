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

	// * get overview statistics from database
	statsRows, err := r.database.P().TaskOverviewByUserId(c.Context(), l.UserId)
	if err != nil {
		return gut.Err(false, "failed to get overview statistics", err)
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

	// * response
	return c.JSON(response.Success(c, &payload.Overview{
		Histories:      histories,
		TokenHistories: statsRows[0].TokenHistories,
		TokenCount:     statsRows[0].TokenCount,
		PoolTokenCount: statsRows[0].PoolTokenCount,
	}))
}
