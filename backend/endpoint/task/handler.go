package taskEndpoint

import (
	"backend/service/task"
	"backend/type/common"
)

type Handler struct {
	database    common.Database
	taskService task.Server
}

func Handle(database common.Database, taskService task.Server) *Handler {
	return &Handler{
		database:    database,
		taskService: taskService,
	}
}
