package endpoint

import (
	"backend/common/fiber/middleware"
	"backend/endpoint/public"
	"backend/endpoint/state"
	"github.com/gofiber/fiber/v2"
)

func Bind(
	app *fiber.App,
	publicEndpoint *publicEndpoint.Handler,
	stateEndpoint *stateEndpoint.Handler,
	middleware *middleware.Middleware,
) {
	api := app.Group("/api")

	// * public endpoints
	public := api.Group("/public")
	public.Get("/login/redirect", publicEndpoint.HandleLoginRedirect)
	public.Post("/login/callback", publicEndpoint.HandleLoginCallback)

	// * state endpoints
	state := api.Group("/state", middleware.Jwt())
	state.Post("/state", stateEndpoint.HandleState)

	// * static files
	app.Static("/attachment", ".local/attachment")
}
