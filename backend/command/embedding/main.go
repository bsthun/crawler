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
	"fmt"
	"github.com/bsthun/gut"
	"github.com/google/uuid"
	"github.com/ollama/ollama/api"
	qd "github.com/qdrant/go-client/qdrant"
	"github.com/tmc/langchaingo/textsplitter"
	"go.uber.org/fx"
	"strconv"
	"time"
)

var embedMigrations embed.FS

type Embedder struct {
	config       *config.Config
	database     common.Database
	qdrantClient *qd.Client
	ollamaClient *api.Client
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
	// * create embedder instance
	embedder := &Embedder{
		config:       config,
		database:     db,
		qdrantClient: qdrantClient,
		ollamaClient: ollamaClient,
	}

	embedder.processCompletedTasks()
}

func (r *Embedder) processCompletedTasks() {
	// * fetch all completed tasks
	tasks, err := r.database.P().TaskListCompleted(context.Background())
	if err != nil {
		gut.Fatal("failed to fetch completed tasks", err)
	}

	gut.Debug("found %d completed tasks to embed", len(tasks))

	// * process each task
	for _, task := range tasks {
		r.embedTask(&task)
	}

	gut.Debug("embedding completed for all tasks")
}

func (r *Embedder) embedTask(task *psql.Task) {
	// * validate task has content
	if task.Content == nil || *task.Content == "" {
		return
	}

	// * check if task already has embeddings in qdrant
	searchResp, err := r.qdrantClient.GetPointsClient().Scroll(context.Background(), &qd.ScrollPoints{
		CollectionName: *r.config.QdrantCollection,
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
		Limit: gut.Ptr(uint32(1)),
	})
	if err != nil {
		gut.Fatal("failed to search task in qdrant", fmt.Errorf("task %d: %v", *task.Id, err))
		return
	}

	if len(searchResp.Result) > 0 {
		gut.Debug("task %d already has embeddings, skipping", *task.Id)
		return
	}

	gut.Debug("embedding task %d", *task.Id)

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

	chunks, err := splitter.SplitText(*task.Content)
	if err != nil {
		gut.Fatal("failed to split task content into chunks", fmt.Errorf("task %d: %v", *task.Id, err))
		return
	}

	gut.Debug("task %d split into %d chunks", *task.Id, len(chunks))

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
			gut.Fatal("failed to get embedding", fmt.Errorf("task %d chunk %d: %v", *task.Id, i, err))
			return
		}

		// * search in qdrant for similarity
		searchResp, err := r.qdrantClient.GetPointsClient().Search(context.Background(), &qd.SearchPoints{
			CollectionName: *r.config.QdrantCollection,
			Vector:         embeddingResp.Embeddings[0],
			Limit:          uint64(1),
			ScoreThreshold: gut.Ptr(float32(1)),
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
			gut.Fatal("failed to search in qdrant", fmt.Errorf("task %d chunk %d: %v", *task.Id, i, err))
			return
		}

		// * check if duplicate found
		if len(searchResp.Result) > 0 {
			// * extract duplicate taskId
			duplicateTaskId, err := strconv.ParseUint(searchResp.Result[0].Payload["taskId"].GetStringValue(), 10, 64)
			if err != nil {
				gut.Fatal("failed to parse duplicate taskId", fmt.Errorf("task %d chunk %d: %v", *task.Id, i, err))
				return
			}

			duplicateCount++
			duplicateTaskIds = append(duplicateTaskIds, fmt.Sprintf("#%s", gut.EncodeId(duplicateTaskId)))
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
			gut.Fatal("failed to upsert point to qdrant", fmt.Errorf("task %d chunk %d: %v", *task.Id, i, err))
			return
		}
	}

	gut.Debug("task %d embedded successfully with %d chunks", *task.Id, len(chunks))
}
