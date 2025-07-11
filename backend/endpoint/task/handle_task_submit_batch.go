package taskEndpoint

import (
	"backend/generate/psql"
	"backend/type/common"
	"backend/type/payload"
	"backend/type/response"
	"encoding/csv"
	"strings"

	"github.com/bsthun/gut"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

func (r *Handler) HandleTaskSubmitBatch(c *fiber.Ctx) error {
	// * login claims
	l := c.Locals("l").(*jwt.Token).Claims.(*common.LoginClaims)

	// * get csv file from multipart form
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return gut.Err(false, "failed to get csv file", err)
	}

	// * open file
	file, err := fileHeader.Open()
	if err != nil {
		return gut.Err(false, "failed to open file", err)
	}
	defer file.Close()

	// * create csv reader
	csvReader := csv.NewReader(file)
	records, err := csvReader.ReadAll()
	if err != nil {
		return gut.Err(false, "failed to read csv file", err)
	}

	if len(records) == 0 {
		return gut.Err(false, "csv file is empty", nil)
	}

	// * begin transaction
	tx, querier := r.database.Ptx(c.Context(), nil)
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
		}
	}()

	// * create upload record
	upload, err := querier.UploadCreate(c.Context(), l.UserId)
	if err != nil {
		return gut.Err(false, "failed to create upload record", err)
	}

	var createdTasks []*psql.Task

	// * iterate through csv records
	for i, record := range records {
		// * skip header row if exists
		if i == 0 && (strings.Contains(record[0], "category") || strings.Contains(record[0], "topic")) {
			continue
		}

		// * validate record has 4 columns
		if len(record) < 4 {
			return gut.Err(false, "csv record must have at least 4 columns (category, type, source, content)", nil)
		}

		category := strings.TrimSpace(record[0])
		source := strings.TrimSpace(record[1])
		taskType := strings.TrimSpace(record[2])
		content := strings.TrimSpace(record[3])

		if taskType == "pdf" {
			taskType = "doc"
		}

		// * validate required fields
		if category == "" || source == "" || taskType == "" {
			continue
		}

		var task *psql.Task
		var er *gut.ErrorInstance

		// * check content
		if content != "" {
			task, er = r.taskProcedure.TaskRawCreate(c.Context(), querier, l.UserId, upload.Id, &category, &taskType, &source, gut.Ptr(""), &content)
		} else {
			task, er = r.taskProcedure.TaskCreate(c.Context(), querier, l.UserId, upload.Id, &category, &taskType, &source)
		}

		if er != nil {
			return er
		}

		createdTasks = append(createdTasks, task)
	}

	// * commit transaction
	if err := tx.Commit(); err != nil {
		return gut.Err(false, "failed to commit transaction", err)
	}

	// * response
	return c.JSON(response.Success(c, &payload.TaskSubmitBatchResponse{
		TasksCreated: gut.Ptr(len(createdTasks)),
		Tasks:        createdTasks,
	}))
}
