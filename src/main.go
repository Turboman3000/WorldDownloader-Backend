package main

import (
	"context"
	"log"
	"math/rand"
	"os"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	ctx := context.Background()
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalln(err)
	}

	s3Client, err := minio.New(os.Getenv("EU_S3_URL"), &minio.Options{
		Creds:  credentials.NewStaticV4(os.Getenv("S3_ACCESS_TOKEN"), os.Getenv("S3_SECRET_TOKEN"), ""),
		Region: "auto",
		Secure: true,
	})

	if err != nil {
		log.Fatalln(err)
	}

	app := fiber.New(fiber.Config{
		JSONEncoder:   json.Marshal,
		JSONDecoder:   json.Unmarshal,
		ServerHeader:  "WDL-Backend",
		CaseSensitive: true,
		AppName:       "WorldDownloader Backend",
	})
	app.Use(cors.New())

	app.Post("/api/v1/upload", func(c *fiber.Ctx) error {
		c.Response().Header.Add("Content-Type", "application/json")

		var id = generateID()

		return c.SendString(`{"status":"ok","code":200,"id":"` + id + `"}`)
	})

	app.Get("/api/v1/download", func(c *fiber.Ctx) error {
		c.Response().Header.Add("Content-Type", "application/json")
		return c.SendString("{}")
	})

	app.Listen(":8080")
}

func generateID() string {
	var id = ""
	var chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789"

	for i := 0; i < 12; i++ {
		id += string(chars[rand.Intn(len(chars))])
	}

	return id
}
