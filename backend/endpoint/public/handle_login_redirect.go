package publicEndpoint

import (
	"backend/type/payload"
	"backend/type/response"
	"github.com/bsthun/gut"
	"github.com/gofiber/fiber/v2"
)

func (r *Handler) HandleLoginRedirect(c *fiber.Ctx) error {
	return c.JSON(response.Success(c, &payload.OauthRedirect{
		RedirectUrl: gut.Ptr(r.Oauth2Config.AuthCodeURL("1")),
	}))
}
