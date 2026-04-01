package main

import (
	"log"

	gameengine "api-server/game-engine"

	"github.com/gofiber/fiber/v3"
)

func submitHandler(engine *gameengine.GameEngine) fiber.Handler {
	return func(c fiber.Ctx) error {
		var s gameengine.Submission
		if err := c.Bind().Body(&s); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
		}
		engine.Submit(s)
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "received"})
	}
}

func main() {
	engine := gameengine.New()

	app := fiber.New()
	app.Post("/submit", submitHandler(engine))
	app.Get("/start", func(c fiber.Ctx) error {
		engine.Start()
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "started"})
	})
	app.Get("/reset", func(c fiber.Ctx) error {
		engine.Reset()
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "reset"})
	})
	log.Fatal(app.Listen(":3000"))
}
