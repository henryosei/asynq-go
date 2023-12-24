package handlers

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/henryosei/async-go/tasks"
	"io"
)

func FileUpload(c *fiber.Ctx) error {

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "upload failed"})
	}
	fileData, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to open file"})
	}
	defer fileData.Close()
	data, err := io.ReadAll(fileData)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to open file"})
	}
	resize, err := tasks.NewImageResizerTask(data, file.Filename)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create image resize tasks"})
	}
	client := tasks.GetClient()
	for _, task := range resize {
		if _, err := client.Enqueue(task); err != nil {
			fmt.Printf("Error enqueuing task: %v\n")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not enqueue image resize task"})
		}
	}
	return c.JSON(fiber.Map{"message": "image uploaded and resizing task started"})
}
