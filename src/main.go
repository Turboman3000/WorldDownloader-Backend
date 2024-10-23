package main

import (
	"archive/zip"
	"bytes"
	"math/rand"
	"os"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"

	"github.com/go-resty/resty/v2"
)

var worlds []IWorld
var iips []IIP

func main() {
	err := godotenv.Load()

	if err != nil {
		panic(err)
	}

	apiClient := resty.New()

	os.RemoveAll("./worlds")
	os.Mkdir("./worlds", os.ModePerm)

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
		c.Response().Header.Add("Content-Type", "application/json")

		var alreadyIP = false

		for i, ip := range iips {
			if ip.IP != c.IP() {
				continue
			}

			var newARRAY []IIP = iips

			if ip.Expires <= time.Now().Unix() {
				newARRAY = slices.Delete(newARRAY, slices.IndexFunc(newARRAY, func(item IIP) bool {
					return item.IP == ip.IP
				}), 1)
			}

			if ip.TimesUsed >= 5 {
				c.Status(429)
				return c.SendString(`{"status":"Rate Limited","code":429}`)
			} else {
				newARRAY[i].TimesUsed += 1
			}

			iips = newARRAY

			alreadyIP = true
		}

		if !alreadyIP {
			iips = append(iips, IIP{
				IP:        c.IP(),
				TimesUsed: 1,
				Expires:   time.Now().Add(time.Hour * 6).Unix(),
			})
		}

		name := c.FormValue("name")

		file, err := c.FormFile("file")

		if err != nil {
			panic(err)
		}

		if !strings.HasSuffix(file.Filename, ".zip") {
			c.Status(400)
			return c.SendString(`{"status":"Invalid File","code":400}`)
		}

		var id = generateID()
		fileContent, err := file.Open()

		if err != nil {
			panic(err)
		}

		fileData := make([]byte, file.Size)

		resp, err := apiClient.R().SetFileReader("file", "file.zip", bytes.NewReader(fileData)).Post("http://" + os.Getenv("CLAMAV_HOST") + "/scan")

		if err != nil {
			panic(err)
		}

		var jsonMap ClamAVResp
		json.Unmarshal(resp.Body(), &jsonMap)

		if jsonMap.Status == "FOUND" {
			c.Status(400)
			return c.SendString(`{"status":"Malicious File","code":400}`)
		}

		fileContent.Read(fileData)

		zipCont, err := zip.NewReader(fileContent, file.Size)

		if err != nil {
			panic(err)
		}

		if !testFiles(zipCont.File) {
			c.Status(400)
			return c.SendString(`{"status":"Invalid File","code":400}`)
		}

		err = os.WriteFile(path.Join("./worlds/"+id+".zip"), fileData, os.ModeAppend)

		if err != nil {
			panic(err)
		}

		var world = IWorld{
			ID:      id,
			Name:    name,
			Expires: time.Now().Add(time.Minute * 30).Unix(),
		}

		go removeWorld(world)

		return c.SendString(`{"status":"ok","code":200,"id":"` + id + `"}`)
	})

	app.Get("/api/v1/download", func(c *fiber.Ctx) error {
		var code = c.Query("c")

		var name = ""

		for _, world := range worlds {
			if world.ID == code {
				name = world.Name
				break
			}
		}

		_, err := os.OpenFile("./worlds/"+code+".zip", os.O_RDONLY, 0644)

		if err != nil {
			c.Response().Header.Add("Content-Type", "application/json")
			return c.SendString(`{"status":"Not Found","code":404}`)
		}

		c.Response().Header.Add("Content-Type", "application/zip")
		c.Response().Header.Add("Content-Disposition", `attachment; filename="`+name+`.zip"`)
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

func removeWorld(world IWorld) {
	worlds = append(worlds, world)
	time.Sleep(time.Duration(world.Expires-time.Now().Unix()) * time.Second)

	os.Remove("./worlds/" + world.ID + ".zip")
	worlds = removeSplice(worlds, world)
}

type IIP struct {
	IP        string
	Expires   int64
	TimesUsed int32
}

type IWorld struct {
	ID      string
	Name    string
	Expires int64
}

type ClamAVResp struct {
	Status      string `json:"Status"`
	Description string `json:"Description"`
}

func removeSplice(s []IWorld, i IWorld) []IWorld {
	s[slices.Index(worlds, i)] = s[len(s)-1]
	return s[:len(s)-1]
}

func testFiles(files []*zip.File) bool {
	var haveRegions = false
	var haveLevel = false

	var regionCheckDone = false
	var levelCheckDone = false

	for _, file := range files {
		path := strings.Split(file.Name, "/")
		endings := strings.Split(file.Name, ".")

		if haveLevel && haveRegions {
			break
		}

		if !regionCheckDone && path[1] == "region" && endings[len(endings)-1] == "mca" {
			haveRegions = true
			regionCheckDone = true
		}

		if !levelCheckDone && path[1] == "level.dat" {
			haveLevel = true
			levelCheckDone = true
		}
	}

	return haveRegions && haveLevel
}
