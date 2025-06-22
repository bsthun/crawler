package main

import (
	"backend/common/config"
	"backend/common/database"
	"backend/common/fiber"
	"backend/common/fiber/middleware"
	"backend/common/ollama"
	"backend/common/qdrant"
	"backend/endpoint"
	publicEndpoint "backend/endpoint/public"
	stateEndpoint "backend/endpoint/state"
	taskEndpoint "backend/endpoint/task"
	taskProcedure "backend/service/task"
	"embed"
	"go.uber.org/fx"
)

//go:embed database/postgres/migration/*.sql
var embedMigrations embed.FS

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
			fiber.Init,
			middleware.Init,
			taskProcedure.Serve,
			publicEndpoint.Handle,
			stateEndpoint.Handle,
			taskEndpoint.Handle,
		),
		fx.Invoke(
			endpoint.Bind,
		),
	).Run()
}
