package main

import (
	"backend/common/config"
	"backend/common/database"
	"backend/generate/psql"
	"backend/type/common"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"github.com/bsthun/gut"
	"github.com/go-resty/resty/v2"
	"go.uber.org/fx"
	"time"
)

var embedMigrations embed.FS

type ExtractResponse struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type Worker struct {
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
	// * create worker instance
	worker := &Worker{
		config:   config,
		database: db,
	}

	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go worker.run()
			gut.Debug("worker started")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			gut.Debug("worker stopped")
			return nil
		},
	})
}

func (r *Worker) run() {
	for {
		// * process one task
		r.process()

		// * sleep 1 second before next iteration
		time.Sleep(1 * time.Second)
	}
}

func (r *Worker) process() {
	// * claim pending task
	task, err := r.database.P().TaskClaimPending(context.Background())
	if err != nil {
		// * no pending tasks or database error, continue
		return
	}

	// * determine endpoint
	var endpoint string
	switch *task.Type {
	case "web":
		endpoint = *r.config.EndpointWebExtract
	case "doc":
		endpoint = *r.config.EndpointDocExtract
	case "youtube":
		endpoint = *r.config.EndpointYoutubeExtract
	default:
		if err := r.database.P().TaskUpdateFailed(context.Background(), &psql.TaskUpdateFailedParams{
			Id:           task.Id,
			FailedReason: gut.Ptr("invalid task type"),
		}); err != nil {
			gut.Fatal("failed to update task as failed", err)
		}

		return
	}

	// * prepare request payload
	payload := map[string]string{
		"url": *task.Source,
	}

	// * extraction service
	resp, err := resty.New().R().
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		Post(endpoint)

	if err != nil {
		// * network error
		if err := r.database.P().TaskUpdateFailed(context.Background(), &psql.TaskUpdateFailedParams{
			Id:           task.Id,
			FailedReason: gut.Ptr(fmt.Sprintf("network error: %v", err)),
		}); err != nil {
			gut.Fatal("failed to update task as failed", err)
		}
		return
	}

	// * handle response based on status code
	if resp.StatusCode() == 200 {
		// * success response
		var extractResp *ExtractResponse
		if err := json.Unmarshal(resp.Body(), &extractResp); err != nil {
			if err := r.database.P().TaskUpdateFailed(context.Background(), &psql.TaskUpdateFailedParams{
				Id:           task.Id,
				FailedReason: gut.Ptr(fmt.Sprintf("parsing: %v", err)),
			}); err != nil {
				gut.Fatal("failed to update task as failed", err)
			}
			return
		}

		// * update task as completed
		if err := r.database.P().TaskUpdateCompleted(context.Background(), &psql.TaskUpdateCompletedParams{
			Id:      task.Id,
			Title:   &extractResp.Title,
			Content: &extractResp.Text,
		}); err != nil {
			gut.Fatal("failed to update task as completed", err)
		}

	} else if resp.StatusCode() >= 400 && resp.StatusCode() < 500 {
		// * parse error response
		var errorResp *ErrorResponse
		if err := json.Unmarshal(resp.Body(), &errorResp); err != nil {
			// * fallback to raw response body
			if err := r.database.P().TaskUpdateFailed(context.Background(), &psql.TaskUpdateFailedParams{
				Id:           task.Id,
				FailedReason: gut.Ptr("error parsing"),
			}); err != nil {
				gut.Fatal("failed to update task as failed", err)
			}
			return
		}

		// * update task as failed with error reason
		if err := r.database.P().TaskUpdateFailed(context.Background(), &psql.TaskUpdateFailedParams{
			Id:           task.Id,
			FailedReason: &errorResp.Error,
		}); err != nil {
			gut.Fatal("failed to update task as failed", err)
		}

	} else {
		// * other status codes
		if err := r.database.P().TaskUpdateFailed(context.Background(), &psql.TaskUpdateFailedParams{
			Id:           task.Id,
			FailedReason: gut.Ptr(fmt.Sprintf("HTTP %d", resp.StatusCode())),
		}); err != nil {
			gut.Fatal("failed to update task as failed", err)
		}
	}
}
