package main

import (
	"backend/common/config"
	"backend/common/database"
	"backend/common/ollama"
	"backend/common/qdrant"
	"backend/generate/psql"
	"backend/type/common"
	"context"
	"embed"
	"flag"
	"fmt"
	"github.com/bsthun/gut"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/ollama/ollama/api"
	qd "github.com/qdrant/go-client/qdrant"
	"github.com/tmc/langchaingo/textsplitter"
	"go.uber.org/fx"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var embedMigrations embed.FS

type ExtractResponse struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

type ErrorResponse struct {
	Detail any `json:"detail"`
}

type TokenResponse struct {
	TokenCount int32 `json:"tokenCount"`
}

type EmbeddingResponse struct {
	Embeddings []float32 `json:"embeddings"`
}

type Stat struct {
	StartedAt           *time.Time       `json:"startedAt"`
	FinishedAt          *time.Time       `json:"finishedAt"`
	ExtractDurations    []*time.Duration `json:"extractDurations"`
	TokenCountDurations []*time.Duration `json:"tokenCountDurations"`
	EmbeddingDurations  []*time.Duration `json:"embeddingDurations"`
	ChunkCount          int              `json:"chunkCount"`
}

type Worker struct {
	config       *config.Config
	database     common.Database
	qdrantClient *qd.Client
	ollamaClient *api.Client
	ExtractPool  *Pool[*string]
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
			ollama.Init,
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
	ollamaClient *api.Client,
) {
	// * create worker instance
	worker := &Worker{
		config:       config,
		database:     db,
		qdrantClient: qdrantClient,
		ollamaClient: ollamaClient,
		ExtractPool:  NewPool(config.EndpointExtracts),
	}

	// * Parse arguments
	thread := flag.Int("thread", 1, "Number of worker threads")
	flag.Parse()

	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			for i := 0; i < *thread; i++ {
				go func() {
					for {
						worker.process()
						time.Sleep(1 * time.Second)
					}
				}()
			}
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

	// * construct stat
	stat := &Stat{
		StartedAt:           gut.Ptr(time.Now()),
		FinishedAt:          nil,
		ExtractDurations:    nil,
		TokenCountDurations: nil,
		EmbeddingDurations:  nil,
		ChunkCount:          0,
	}
	_ = stat

	// * construct text
	title := new(string)
	content := task.Content

	if content == nil {
		base := *r.ExtractPool.Get()
		path := ""
		defer r.ExtractPool.Put(&base)

		switch *task.Type {
		case "web":
			path = *r.config.EndpointWebPath
		case "doc":
			path = *r.config.EndpointDocPath
		case "youtube":
			path = *r.config.EndpointYoutubePath
		default:
			gut.Fatal("unknown task type", fmt.Errorf("%s", *task.Type))
			return
		}

		// * get endpoint from pool
		endpoint, err := url.JoinPath(base, path)
		if err != nil {
			gut.Fatal("failed to join url path", err)
			return
		}

		// * prepare request payload
		payload := map[string]string{
			"url": *task.Source,
		}

		// * extraction service with retry
		extractAttempt := 0
		extractResp := new(ExtractResponse)
		errorResp := new(ErrorResponse)

	extractAttempt:
		extractAttempt++
		extractStart := time.Now()
		resp, err := resty.New().R().
			SetHeader("Content-Type", "application/json").
			SetBody(payload).
			SetResult(extractResp).
			SetError(errorResp).
			Post(endpoint)
		if err != nil {
			// * network error for extraction
			if err := r.database.P().TaskUpdateFailed(context.Background(), &psql.TaskUpdateFailedParams{
				Id:           task.Id,
				FailedReason: gut.Ptr(fmt.Sprintf("extraction error: %v", err)),
				Title:        nil,
				Content:      nil,
				TokenCount:   nil,
			}); err != nil {
				gut.Fatal("failed to update task as failed", err)
			}
			return
		}

		// * handle server error with retry
		if resp.StatusCode() >= 500 {
			if extractAttempt < 3 {
				stat.ExtractDurations = append(stat.ExtractDurations, gut.Ptr(time.Since(extractStart)))
				time.Sleep(3 * time.Second)
				goto extractAttempt
			}
			// * update task as failed with server error reason
			if err := r.database.P().TaskUpdateFailed(context.Background(), &psql.TaskUpdateFailedParams{
				Id:           task.Id,
				FailedReason: gut.Ptr(fmt.Sprintf("extraction %d (%s)", resp.StatusCode(), resp.Body())),
				Title:        nil,
				Content:      nil,
				TokenCount:   nil,
			}); err != nil {
				gut.Fatal("failed to update task as failed", err)
			}
			return
		}

		// * handle client error (4xx should not be retried)
		if resp.StatusCode() >= 400 {
			if extractAttempt < 3 {
				stat.ExtractDurations = append(stat.ExtractDurations, gut.Ptr(time.Since(extractStart)))
				time.Sleep(3 * time.Second)
				goto extractAttempt
			}

			var message string
			if detail, ok := errorResp.Detail.(string); ok {
				message = fmt.Sprintf("extraction: %s", detail)
			} else if detail, ok := errorResp.Detail.(map[string]any); ok {
				if msg, ok := detail["error"].(string); ok {
					message = fmt.Sprintf("extraction: %s", msg)
				} else {
					message = fmt.Sprintf("extraction: %v", detail)
				}
			} else {
				message = fmt.Sprintf("extraction: %s", resp.Body())
			}

			// * update task as failed with error reason
			if err := r.database.P().TaskUpdateFailed(context.Background(), &psql.TaskUpdateFailedParams{
				Id:           task.Id,
				FailedReason: &message,
				Title:        nil,
				Content:      nil,
				TokenCount:   nil,
			}); err != nil {
				gut.Fatal("failed to update task as failed", err)
			}
			return
		}

		if extractAttempt == 1 && len(extractResp.Text) < 512 {
			goto extractAttempt
		}

		content = gut.Ptr(strings.ToValidUTF8(extractResp.Text, ""))
	} else {
		content = gut.Ptr(strings.ToValidUTF8(*content, ""))
	}
	length := 100
	contentLen := len([]rune(*content))
	if contentLen > length {
		contentLen = length
	}
	runes := []rune(*content)
	title = gut.Ptr(string(runes[:contentLen]))

	// * call tokenization service
	tokenResp := new(TokenResponse)
	tokenPayload := map[string]string{
		"text": *content,
	}

	resp, err := resty.New().R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetFormData(tokenPayload).
		SetResult(&tokenResp).
		Post(*r.config.EndpointTokenCount)
	if err != nil {
		// * network error for tokenization
		if err := r.database.P().TaskUpdateFailed(context.Background(), &psql.TaskUpdateFailedParams{
			Id:           task.Id,
			FailedReason: gut.Ptr(fmt.Sprintf("token error: %v", err)),
			Title:        title,
			Content:      content,
			TokenCount:   nil,
		}); err != nil {
			gut.Fatal("failed to update task as failed", err)
		}
		return
	}

	// * handle server error
	if resp.StatusCode() >= 500 {
		if err := r.database.P().TaskUpdateFailed(context.Background(), &psql.TaskUpdateFailedParams{
			Id:           task.Id,
			FailedReason: gut.Ptr(fmt.Sprintf("tokenization %d (%s)", resp.StatusCode(), resp.Body())),
			Title:        title,
			Content:      content,
			TokenCount:   nil,
		}); err != nil {
			gut.Fatal("failed to update task as failed", err)
		}
		return
	}

	// * split content to chunks
	splitter := textsplitter.NewRecursiveCharacter(
		textsplitter.WithChunkSize(65535),
		textsplitter.WithChunkOverlap(64),
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
			Title:        title,
			Content:      content,
			TokenCount:   &tokenResp.TokenCount,
		}); err != nil {
			gut.Fatal("failed to update task as failed", err)
		}
		return
	}

	duplicateCount := 0
	duplicateTaskIds := make([]string, 0)
	for i, chunk := range chunks {
		// * get embedding
		embeddingAttempt := 0
		var embeddingResp *api.EmbedResponse
	embeddingAttempt:
		embeddingAttempt++
		embeddingResp, err := r.ollamaClient.Embed(context.Background(), &api.EmbedRequest{
			Model:     *r.config.OllamaEmbeddingModel,
			Input:     chunk,
			KeepAlive: nil,
			Truncate:  nil,
			Options:   nil,
		})
		if err != nil {
			if embeddingAttempt < 3 {
				time.Sleep(2 * time.Second)
				goto embeddingAttempt
			}
			if err := r.database.P().TaskUpdateFailed(context.Background(), &psql.TaskUpdateFailedParams{
				Id:           task.Id,
				FailedReason: gut.Ptr(fmt.Sprintf("embedding error: %v", err)),
				Title:        title,
				Content:      content,
				TokenCount:   &tokenResp.TokenCount,
			}); err != nil {
				gut.Fatal("failed to update task as failed", err)
			}
			return
		}

		// * search in qdrant for similarity
		searchResp, err := r.qdrantClient.GetPointsClient().Search(context.Background(), &qd.SearchPoints{
			CollectionName: *r.config.QdrantCollection,
			Vector:         embeddingResp.Embeddings[0],
			Limit:          uint64(1),
			ScoreThreshold: gut.Ptr(float32(0.90)),
			WithPayload: &qd.WithPayloadSelector{
				SelectorOptions: &qd.WithPayloadSelector_Enable{
					Enable: true,
				},
			},
			Filter: &qd.Filter{
				Must: []*qd.Condition{
					{
						ConditionOneOf: &qd.Condition_Field{
							Field: &qd.FieldCondition{
								Key: "type",
								Match: &qd.Match{
									MatchValue: &qd.Match_Keyword{
										Keyword: *task.Type,
									},
								},
							},
						},
					},
				},
				MustNot: []*qd.Condition{
					{
						ConditionOneOf: &qd.Condition_Field{
							Field: &qd.FieldCondition{
								Key: "taskId",
								Match: &qd.Match{
									MatchValue: &qd.Match_Keyword{
										Keyword: strconv.FormatUint(*task.Id, 10),
									},
								},
							},
						},
					},
				},
			},
		})
		if err != nil {
			if err := r.database.P().TaskUpdateFailed(context.Background(), &psql.TaskUpdateFailedParams{
				Id:           task.Id,
				FailedReason: gut.Ptr(fmt.Sprintf("qdrant search error: %v", err)),
				Title:        title,
				Content:      content,
				TokenCount:   &tokenResp.TokenCount,
			}); err != nil {
				gut.Fatal("failed to update task as failed", err)
			}
			return
		}

		// * check if duplicate found
		if len(searchResp.Result) > 0 {
			// * extract duplicate taskId
			duplicateTaskId, err := strconv.ParseUint(searchResp.Result[0].Payload["taskId"].GetStringValue(), 10, 64)
			if err != nil {
				gut.Fatal("failed to parse duplicate taskId", err)
			}

			duplicateCount++
			duplicateTaskIds = append(duplicateTaskIds, fmt.Sprintf("#%d %.4f%%", duplicateTaskId, searchResp.Result[0].Score*100))
		}

		// * generate uuid for point
		pointId := uuid.New().String()

		// * upsert point to qdrant
		point := &qd.PointStruct{
			Id: &qd.PointId{
				PointIdOptions: &qd.PointId_Uuid{
					Uuid: pointId,
				},
			},
			Vectors: &qd.Vectors{
				VectorsOptions: &qd.Vectors_Vector{
					Vector: &qd.Vector{
						Data: embeddingResp.Embeddings[0],
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
					Kind: &qd.Value_IntegerValue{
						IntegerValue: int64(i),
					},
				},
				"type": {
					Kind: &qd.Value_StringValue{
						StringValue: *task.Type,
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
				Title:        title,
				Content:      content,
				TokenCount:   &tokenResp.TokenCount,
			}); err != nil {
				gut.Fatal("failed to update task as failed", err)
			}
			return
		}
	}

	duplicate := duplicateCount > len(chunks)*2/3
	if duplicate {
		// * rollback qdrant upsert
		if _, err := r.qdrantClient.Delete(context.Background(), &qd.DeletePoints{
			CollectionName: *r.config.QdrantCollection,
			Points: &qd.PointsSelector{
				PointsSelectorOneOf: &qd.PointsSelector_Filter{
					Filter: &qd.Filter{
						Must: []*qd.Condition{
							{
								ConditionOneOf: &qd.Condition_Field{
									Field: &qd.FieldCondition{
										Key: "taskId",
										Match: &qd.Match{
											MatchValue: &qd.Match_Keyword{
												Keyword: strconv.FormatUint(*task.Id, 10),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}); err != nil {
			gut.Fatal("failed to rollback qdrant upsert", err)
		}

		if err := r.database.P().TaskUpdateFailed(context.Background(), &psql.TaskUpdateFailedParams{
			Id:           task.Id,
			FailedReason: gut.Ptr(fmt.Sprintf("duplicate %s", strings.Join(duplicateTaskIds, ", "))),
			Title:        title,
			Content:      content,
			TokenCount:   &tokenResp.TokenCount,
		}); err != nil {
			gut.Fatal("failed to update task as failed", err)
		}
		return
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
