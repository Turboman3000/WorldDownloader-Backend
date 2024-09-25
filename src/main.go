package main

import (
	"context"
	"math/rand"
	"os"
	"path"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

var ctx = context.Background()

func main() {
	app := fiber.New(fiber.Config{
		JSONEncoder:   json.Marshal,
		JSONDecoder:   json.Unmarshal,
		ServerHeader:  "WDL-Backend",
		BodyLimit:     5 * 1024 * 1024 * 1024 * 1024,
		CaseSensitive: true,
		AppName:       "WorldDownloader Backend",
	})
	app.Use(cors.New())

	app.Get("/api/v1/status", func(c *fiber.Ctx) error {
		c.Response().Header.Add("Content-Type", "application/json")
		return c.SendString(`{"status":"ok","code":200}`)
	})

	app.Post("/api/v1/upload", func(c *fiber.Ctx) error {
		var id = generateID()

		//name := c.FormValue("name")

		file, err := c.FormFile("file")

		if err != nil {
			panic(err)
		}

		fileContent, err := file.Open()

		if err != nil {
			panic(err)
		}

		fileData := make([]byte, file.Size)
		fileContent.Read(fileData)

		os.WriteFile(path.Join("./worlds/"+id+".zip"), fileData, os.ModeAppend)

		c.Response().Header.Add("Content-Type", "application/json")
		return c.SendString(`{"status":"ok","code":200,"id":"` + id + `"}`)
	})

	app.Get("/api/v1/download", func(c *fiber.Ctx) error {
		var code = c.Query("c")

		_, err := os.OpenFile("./worlds/"+code+".zip", os.O_RDONLY, 0644)

		if err != nil {
			c.Response().Header.Add("Content-Type", "application/json")
			return c.SendString(`{"status":"Not Found","code":404}`)
		}

		c.Response().Header.Add("Content-Type", "application/zip")
		c.Response().Header.Add("Content-Disposition", `attachment; filename="`+code+`.zip"`)
		return c.SendFile("./worlds/"+code+".zip", true)
	})

	app.Listen("0.0.0.0:8080")
}

func generateID() string {
	var id = ""
	var chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789"

	for i := 0; i < 15; i++ {
		id += string(chars[rand.Intn(len(chars))])
	}

	return id
}
