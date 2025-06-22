package taskEndpoint

import (
	"backend/procedure/task"
	"backend/type/common"
)

type Handler struct {
	database      common.Database
	taskProcedure taskProcedure.Server
}

func Handle(database common.Database, taskService taskProcedure.Server) *Handler {
	return &Handler{
		database:      database,
		taskProcedure: taskService,
	}
}
