package main

import (
	"backend/common/config"
	"backend/common/database"
	"backend/common/qdrant"
	"backend/generate/psql"
	"backend/type/common"
	"context"
	"embed"
	"fmt"
	"github.com/bsthun/gut"
	"github.com/davecgh/go-spew/spew"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	qd "github.com/qdrant/go-client/qdrant"
	"github.com/tmc/langchaingo/textsplitter"
	"go.uber.org/fx"
	"strconv"
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

type EmbeddingResponse struct {
	Embeddings []float32 `json:"embeddings"`
}

type Worker struct {
	config       *config.Config
	database     common.Database
	qdrantClient *qd.Client
}

func main() {
	fx.New(
		fx.Supply(
			embedMigrations,
		),
		fx.Provide(
			config.Init,
			database.Init,
			qdrant.Init,
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
	qdrantClient *qd.Client,
) {
	// * create worker instance
	worker := &Worker{
		config:       config,
		database:     db,
		qdrantClient: qdrantClient,
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

	// * construct text
	title := new(string)
	content := task.Content

	if content == nil {
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

		content = &extractResp.Text
		title = &extractResp.Title
	} else {
		length := 100
		if len(*content) < length {
			length = len(*content)
		}
		title = gut.Ptr((*content)[:length])
	}

	// * call tokenization service
	var tokenResp *TokenResponse
	tokenPayload := map[string]string{
		"text": *content,
	}

	_, err = resty.New().R().
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

	// * split content to chunks
	splitter := textsplitter.NewMarkdownTextSplitter(
		textsplitter.WithChunkSize(500),
		textsplitter.WithChunkOverlap(100),
		textsplitter.WithSeparators([]string{
			"\n\n", // * paragraphs first
			"\n",   // * then newlines
			". ",   // * then sentences
			", ",   // * then commas
			" ",    // * then spaces
			"",     // * then chars
		}),
	)

	chunks, err := splitter.SplitText(*content)
	if err != nil {
		if err := r.database.P().TaskUpdateFailed(context.Background(), &psql.TaskUpdateFailedParams{
			Id:           task.Id,
			FailedReason: gut.Ptr(fmt.Sprintf("text splitting error: %v", err)),
		}); err != nil {
			gut.Fatal("failed to update task as failed", err)
		}
		return
	}

	// * process each chunk
	for i, chunk := range chunks {
		// * get embedding for chunk
		embeddingResp := new(EmbeddingResponse)
		embeddingPayload := map[string]string{
			"text": chunk,
		}

		_, err = resty.New().R().
			SetHeader("Content-Type", "application/x-www-form-urlencoded").
			SetFormData(embeddingPayload).
			SetResult(&embeddingResp).
			Post("http://10.2.1.179:8001/embed")
		if err != nil {
			if err := r.database.P().TaskUpdateFailed(context.Background(), &psql.TaskUpdateFailedParams{
				Id:           task.Id,
				FailedReason: gut.Ptr(fmt.Sprintf("embedding error: %v", err)),
			}); err != nil {
				gut.Fatal("failed to update task as failed", err)
			}
			return
		}

		// * search in qdrant for similar chunks with 95% threshold
		searchResp, err := r.qdrantClient.GetPointsClient().Search(context.Background(), &qd.SearchPoints{
			CollectionName: *r.config.QdrantCollection,
			Vector:         embeddingResp.Embeddings,
			Limit:          uint64(1),
			ScoreThreshold: gut.Ptr(float32(0.95)),
		})
		if err != nil {
			if err := r.database.P().TaskUpdateFailed(context.Background(), &psql.TaskUpdateFailedParams{
				Id:           task.Id,
				FailedReason: gut.Ptr(fmt.Sprintf("qdrant search error: %v", err)),
			}); err != nil {
				gut.Fatal("failed to update task as failed", err)
			}
			return
		}

		// * check if duplicate found
		if len(searchResp.Result) > 0 {
			spew.Dump(searchResp.Result)
			if err := r.database.P().TaskUpdateFailed(context.Background(), &psql.TaskUpdateFailedParams{
				Id:           task.Id,
				FailedReason: gut.Ptr("duplicate"),
			}); err != nil {
				gut.Fatal("failed to update task as failed", err)
			}
			return
		}

		// * generate uuid for point
		pointId := uuid.New().String()

		// * upsert point to qdrant with taskId and chunkNo payload
		point := &qd.PointStruct{
			Id: &qd.PointId{
				PointIdOptions: &qd.PointId_Uuid{
					Uuid: pointId,
				},
			},
			Vectors: &qd.Vectors{
				VectorsOptions: &qd.Vectors_Vector{
					Vector: &qd.Vector{
						Data: embeddingResp.Embeddings,
					},
				},
			},
			Payload: map[string]*qd.Value{
				"taskId": {
					Kind: &qd.Value_StringValue{
						StringValue: strconv.FormatUint(*task.Id, 10),
					},
				},
				"chunkNo": {
					Kind: &qd.Value_StringValue{
						StringValue: strconv.Itoa(i),
					},
				},
			},
		}

		_, err = r.qdrantClient.Upsert(context.Background(), &qd.UpsertPoints{
			CollectionName: *r.config.QdrantCollection,
			Points: []*qd.PointStruct{
				point,
			},
		})
		if err != nil {
			if err := r.database.P().TaskUpdateFailed(context.Background(), &psql.TaskUpdateFailedParams{
				Id:           task.Id,
				FailedReason: gut.Ptr(fmt.Sprintf("qdrant upsert error: %v", err)),
			}); err != nil {
				gut.Fatal("failed to update task as failed", err)
			}
			return
		}
	}

	// * update task as completed
	if err := r.database.P().TaskUpdateCompleted(context.Background(), &psql.TaskUpdateCompletedParams{
		Id:         task.Id,
		Title:      title,
		Content:    content,
		TokenCount: &tokenResp.TokenCount,
	}); err != nil {
		gut.Fatal("failed to update task as completed", err)
	}
}
