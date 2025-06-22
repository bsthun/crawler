package publicEndpoint

import (
	"backend/type/payload"
	"backend/type/response"
	"github.com/gofiber/fiber/v2"
)

func (r *Handler) HandleLoginRedirect(c *fiber.Ctx) error {
	// * construct redirect url
	redirectUrl := r.Oauth2Config.AuthCodeURL("1")

	return c.JSON(response.Success(c, &payload.OauthRedirect{
		RedirectUrl: &redirectUrl,
	}))
}
