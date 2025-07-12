package taskEndpoint

import (
	"backend/type/common"
	"backend/type/payload"
	"backend/type/response"
	"github.com/bsthun/gut"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"regexp"
	"strconv"
)

func (r *Handler) HandleTaskDetail(c *fiber.Ctx) error {
	// * login claims
	l := c.Locals("l").(*jwt.Token).Claims.(*common.LoginClaims)
	_ = l

	// * parse body
	body := new(payload.TaskDetailRequest)
	if err := c.BodyParser(body); err != nil {
		return gut.Err(false, "invalid body", err)
	}

	// * get task by id and validate ownership
	task, err := r.database.P().TaskGetById(c.Context(), body.TaskId)
	if err != nil {
		return gut.Err(false, "task not found or not owned by user", err)
	}

	// * replace encode ids
	if task.FailedReason != nil {
		re := regexp.MustCompile(`(\s)#(\d+)(\s)`)
		*task.FailedReason = re.ReplaceAllStringFunc(*task.FailedReason, func(match string) string {
			submatch := re.FindStringSubmatch(match)
			if len(submatch) == 4 {
				num, err := strconv.ParseUint(submatch[2], 10, 64)
				if err == nil {
					return submatch[1] + "#" + gut.EncodeId(num) + submatch[3]
				}
			}
			return match
		})
	}

	// * response
	return c.JSON(response.Success(c, &payload.TaskDetailResponse{
		Id:           task.Id,
		UserId:       task.UserId,
		UploadId:     task.UploadId,
		CategoryId:   task.CategoryId,
		Type:         task.Type,
		Source:       task.Source,
		IsRaw:        task.IsRaw,
		Status:       task.Status,
		FailedReason: task.FailedReason,
		Title:        task.Title,
		Content:      task.Content,
		TokenCount:   task.TokenCount,
		CreatedAt:    task.CreatedAt,
		UpdatedAt:    task.UpdatedAt,
	}))
}
