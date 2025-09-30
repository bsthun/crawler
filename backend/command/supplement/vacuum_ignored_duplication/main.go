package main

import (
	"backend/common/config"
	"backend/common/database"
	"backend/generate/psql"
	"backend/type/common"
	"context"
	"embed"
	"regexp"
	"strconv"

	"github.com/bsthun/gut"
	"go.uber.org/fx"
)

var embedMigrations embed.FS

type Vacuumer struct {
	config   *config.Config
	database common.Database
}

func main() {
	fx.New(
		fx.Supply(
			embedMigrations,
		),
		fx.Provide(
			config.Init,
			database.Init,
		),
		fx.Invoke(
			invoke,
		),
	).Run()
}

func invoke(
	lifecycle fx.Lifecycle,
	config *config.Config,
	db common.Database,
) {
	// * create vacuumer instance
	vacuumer := &Vacuumer{
		config:   config,
		database: db,
	}

	vacuumer.vacuum()
}

func (r *Vacuumer) vacuum() {
	ctx := context.Background()

	// * get all failed raw duplicate tasks
	tasks, err := r.database.P().TaskListFailedRawDuplicates(ctx)
	if err != nil {
		gut.Fatal("failed to list failed raw duplicate tasks", err)
	}

	gut.Debug("found %d failed raw duplicate tasks", len(tasks))

	// * regex pattern to match "duplicate #(number)(optional colon)(whitespace)"
	duplicatePattern := regexp.MustCompile(`duplicate #(\d+)(:|\s)`)

	processedCount := 0

	for _, task := range tasks {
		if task.FailedReason == nil {
			continue
		}

		// * extract task id from failed reason using regex
		matches := duplicatePattern.FindStringSubmatch(*task.FailedReason)
		if len(matches) < 2 {
			gut.Debug("task %d: failed reason doesn't match pattern: %s", *task.Id, *task.FailedReason)
			continue
		}

		// * parse extracted task id
		duplicateTaskId, err := strconv.ParseUint(matches[1], 10, 64)
		if err != nil {
			gut.Debug("task %d: failed to parse duplicate task id: %s", *task.Id, matches[1])
			continue
		}

		// * check if the duplicate task exists and has status 'ignored'
		duplicateTask, err := r.database.P().TaskGetById(ctx, &duplicateTaskId)
		if err != nil {
			gut.Debug("task %d: duplicate task %d not found", *task.Id, duplicateTaskId)
			continue
		}

		if *duplicateTask.Task.Status != "ignored" {
			gut.Debug("task %d: duplicate task %d status is %v, not 'ignored'", *task.Id, duplicateTaskId, *duplicateTask.Task.Status)
			continue
		}

		// * begin transaction
		tx, querier := r.database.Ptx(ctx, nil)

		// * update task status to completed and set revised_task_id
		err = querier.TaskUpdateCompleted(ctx, &psql.TaskUpdateCompletedParams{
			Id:            task.Id,
			Title:         nil,
			Content:       nil,
			TokenCount:    nil,
			RevisedTaskId: duplicateTask.Task.Id,
		})
		if err != nil {
			_ = tx.Rollback()
			gut.Debug("task %d: failed to update status: %v", *task.Id, err)
			continue
		}

		// * commit transaction
		if err := tx.Commit(); err != nil {
			gut.Debug("task %d: failed to commit transaction: %v", *task.Id, err)
			continue
		}

		gut.Debug("task %d: updated to completed with revised_task_id %d", *task.Id, duplicateTaskId)
		processedCount++
	}

	gut.Debug("processed %d tasks", processedCount)
}
