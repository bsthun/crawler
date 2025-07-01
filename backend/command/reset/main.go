package main

import (
	"backend/common/config"
	"backend/common/database"
	"backend/common/qdrant"
	"backend/type/common"
	"context"
	"embed"
	"github.com/bsthun/gut"
	qd "github.com/qdrant/go-client/qdrant"
	"go.uber.org/fx"
	"strconv"
)

var embedMigrations embed.FS

type Resetter struct {
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
	// * create resetter instance
	resetter := &Resetter{
		config:       config,
		database:     db,
		qdrantClient: qdrantClient,
	}

	resetter.reset()
}

func (r *Resetter) reset() {
	ctx := context.Background()

	// * begin transaction
	tx, querier := r.database.Ptx(ctx, nil)
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
		}
	}()

	// * reset failed tasks to queuing
	tasks, err := querier.TaskResetFailed(ctx)
	if err != nil {
		gut.Fatal("failed to reset failed tasks", err)
	}

	gut.Debug("reset %d failed tasks", len(tasks))

	// * delete points from qdrant for each task
	for _, task := range tasks {
		_, err := r.qdrantClient.Delete(context.Background(), &qd.DeletePoints{
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
		})
		if err != nil {
			gut.Fatal("failed to delete qdrant points", err)
		}
		gut.Debug("deleted qdrant points for task", *task.Id)
	}

	// * commit transaction
	if err := tx.Commit(); err != nil {
		gut.Fatal("failed to commit transaction", err)
	}
}
