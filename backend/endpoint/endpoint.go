package endpoint

import (
	"backend/common/config"
	"backend/common/fiber/middleware"
	"backend/endpoint/public"
	"backend/endpoint/state"
	"backend/endpoint/task"
	"github.com/gofiber/fiber/v2"
	"path/filepath"
)

func Bind(
	app *fiber.App,
	publicEndpoint *publicEndpoint.Handler,
	stateEndpoint *stateEndpoint.Handler,
	taskEndpoint *taskEndpoint.Handler,
	middleware *middleware.Middleware,
	config *config.Config,
) {
	api := app.Group("/api")
	api.Use(middleware.Id())

	// * public endpoints
	public := api.Group("/public")
	public.Get("/login/redirect", publicEndpoint.HandleLoginRedirect)
	public.Post("/login/callback", publicEndpoint.HandleLoginCallback)

	// * state endpoints
	state := api.Group("/state", middleware.Jwt(true))
	state.Post("/state", stateEndpoint.HandleState)
	state.Post("/overview", stateEndpoint.HandleStateOverview)

	// * task endpoints
	task := api.Group("/task", middleware.Jwt(true))
	task.Post("/submit", taskEndpoint.HandleTaskSubmit)
	task.Post("/submit/batch", taskEndpoint.HandleTaskSubmitBatch)
	task.Post("/list", taskEndpoint.HandleTaskList)
	task.Post("/detail", taskEndpoint.HandleTaskDetail)
	task.Post("/category/list", taskEndpoint.HandleTaskCategoryList)
	task.Post("/upload/list", taskEndpoint.HandleTaskUploadList)

	// * static files
	app.Static("/file", ".local/file")

	// * static
	app.Static("/", *config.WebRoot)
	app.Get("/*", func(c *fiber.Ctx) error {
		return c.SendFile(filepath.Join(*config.WebRoot, "index.html"))
	})
}
