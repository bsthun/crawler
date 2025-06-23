package main

import (
	"github.com/gofiber/fiber/v2"
	"log"
	"time"
)

func main() {
	app := fiber.New()

	app.Post("/extract", extractHandler)

	log.Fatal(app.Listen(":3001"))
}

func extractHandler(c *fiber.Ctx) error {
	time.Sleep(10 * time.Second)
	result := map[string]any{
		"title": "Document Title",
		"text":  "Extracted text from the document",
	}

	return c.JSON(result)
}
