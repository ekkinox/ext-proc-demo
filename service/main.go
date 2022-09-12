package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"log"
)

func main() {
	// Fiber
	app := fiber.New()

	// Logger
	app.Use(logger.New())

	// Main endpoint
	app.Post("/", func(c *fiber.Ctx) error {

		//time.Sleep(3 * time.Second)

		h := map[string]interface{}{}
		c.Request().Header.VisitAll(func(key, value []byte) {
			h[string(key)] = string(value)
		})

		return c.JSON(h)
	})

	app.Get("/unprotected", func(c *fiber.Ctx) error {

		h := map[string]interface{}{}
		c.Request().Header.VisitAll(func(key, value []byte) {
			h[string(key)] = string(value)
		})

		return c.JSON(h)
	})

	log.Fatal(app.Listen(":8080"))
}
