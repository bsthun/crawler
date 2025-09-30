package taskEndpoint

import (
	"backend/type/common"
	"backend/type/payload"
	"backend/type/response"
	"regexp"
	"strconv"

	"github.com/bsthun/gut"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
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
	if task.Task.FailedReason != nil {
		re := regexp.MustCompile(`(\s)#(\d+)(\s)`)
		*task.Task.FailedReason = re.ReplaceAllStringFunc(*task.Task.FailedReason, func(match string) string {
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
		Id:           task.Task.Id,
		UserId:       task.Task.UserId,
		UploadId:     task.Task.UploadId,
		CategoryId:   task.Task.CategoryId,
		Type:         task.Task.Type,
		Source:       task.Task.Source,
		IsRaw:        task.Task.IsRaw,
		Status:       task.Task.Status,
		FailedReason: task.Task.FailedReason,
		Title:        task.Task.Title,
		Content:      task.Task.Content,
		TokenCount:   task.Task.TokenCount,
		CreatedAt:    task.Task.CreatedAt,
		UpdatedAt:    task.Task.UpdatedAt,
		User: &payload.UserListItem{
			Id:        task.User.Id,
			Oid:       task.User.Oid,
			Firstname: task.User.Firstname,
			Lastname:  task.User.Lastname,
			Email:     task.User.Email,
			PhotoUrl:  task.User.PhotoUrl,
			IsAdmin:   task.User.IsAdmin,
			CreatedAt: task.User.CreatedAt,
			UpdatedAt: task.User.UpdatedAt,
		},
		Category: &payload.TaskCategoryItem{
			Id:        task.Category.Id,
			Name:      task.Category.Name,
			CreatedAt: task.Category.CreatedAt,
			UpdatedAt: task.Category.UpdatedAt,
		},
	}))
}
