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
	"go.uber.org/fx"
)

func main() {
	fx.New(
		fx.Provide(
			config.Init,
			database.Init,
			qdrant.Init,
			ollama.Init,
			fiber.Init,
			middleware.Init,
			publicEndpoint.Handle,
			stateEndpoint.Handle,
		),
		fx.Invoke(
			endpoint.Bind,
		),
	).Run()
}
