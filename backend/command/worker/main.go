package main

import (
	"backend/common/config"
	"backend/common/database"
	"backend/generate/psql"
	"backend/type/common"
	"context"
	"embed"
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

type TokenResponse struct {
	TokenCount int32 `json:"tokenCount"`
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
			go func() {
				for {
					worker.process()
					time.Sleep(1 * time.Second)
				}
			}()
			gut.Debug("worker started")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			gut.Debug("worker stopped")
			return nil
		},
	})
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
	extractResp := new(ExtractResponse)
	errorResp := new(ErrorResponse)
	resp, err := resty.New().R().
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		SetResult(extractResp).
		SetError(errorResp).
		Post(endpoint)
	if err != nil {
		if err := r.database.P().TaskUpdateFailed(context.Background(), &psql.TaskUpdateFailedParams{
			Id:           task.Id,
			FailedReason: gut.Ptr(fmt.Sprintf("extraction error: %v", err)),
		}); err != nil {
			gut.Fatal("failed to update task as failed", err)
		}
		return
	}

	// * handle server error
	if resp.StatusCode() >= 500 {
		// * update task as failed with server error reason
		if err := r.database.P().TaskUpdateFailed(context.Background(), &psql.TaskUpdateFailedParams{
			Id:           task.Id,
			FailedReason: gut.Ptr(fmt.Sprintf("extraction %d (%s)", resp.StatusCode(), resp.Body())),
		}); err != nil {
			gut.Fatal("failed to update task as failed", err)
		}
		return
	}

	// * handle client error
	if resp.StatusCode() >= 400 {
		// * update task as failed with error reason
		if err := r.database.P().TaskUpdateFailed(context.Background(), &psql.TaskUpdateFailedParams{
			Id:           task.Id,
			FailedReason: &errorResp.Error,
		}); err != nil {
			gut.Fatal("failed to update task as failed", err)
		}
		return
	}

	// * handle response based on status code
	if resp.StatusCode() == 200 {
		// * call tokenization service
		var tokenResp *TokenResponse
		tokenPayload := map[string]string{
			"text": extractResp.Text,
		}

		_, err := resty.New().R().
			SetHeader("Content-Type", "application/x-www-form-urlencoded").
			SetFormData(tokenPayload).
			SetResult(&tokenResp).
			Post("http://10.2.1.179:8003/tokenize")
		if err != nil {
			// * network error for tokenization
			if err := r.database.P().TaskUpdateFailed(context.Background(), &psql.TaskUpdateFailedParams{
				Id:           task.Id,
				FailedReason: gut.Ptr(fmt.Sprintf("token error: %v", err)),
			}); err != nil {
				gut.Fatal("failed to update task as failed", err)
			}
			return
		}

		// * update task as completed
		if err := r.database.P().TaskUpdateCompleted(context.Background(), &psql.TaskUpdateCompletedParams{
			Id:         task.Id,
			Title:      &extractResp.Title,
			Content:    &extractResp.Text,
			TokenCount: &tokenResp.TokenCount,
		}); err != nil {
			gut.Fatal("failed to update task as completed", err)
		}
	}
}
