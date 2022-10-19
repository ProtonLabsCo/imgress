package main

import (
	"fmt"
	"log"
	"mime/multipart"
	"runtime"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/template/html"
)

func main() {
	// 4 threads/childs max
	runtime.GOMAXPROCS(4)

	// start a cleanup cron-job
	go CleanUp()

	engine := html.New("./static", ".html")

	// create new fiber instance  and use across whole app
	app := fiber.New(fiber.Config{
		Views:     engine,
		BodyLimit: 100 * 1024 * 1024, // this is the default limit of 100MB
		Prefork:   true,
	})

	// middleware to allow all clients to communicate using http and allow cors
	app.Use(cors.New())

	app.Static("/images", "./images")

	// homepage
	app.Get("/", func(c *fiber.Ctx) error {
		// rendering the "index" template with content passing
		return c.Render("index", fiber.Map{})
	})

	// handle image uploading using post request
	app.Post("/", handleFileupload)

	// start dev server on port 4000
	log.Fatal(app.Listen(":4000"))
}

func handleFileupload(c *fiber.Ctx) error {
	form, err := c.MultipartForm()
	if err != nil {
		return err
	}

	hcaptchaPass := HandleCaptcha(form.Value["h-captcha-response"][0])
	if !hcaptchaPass {
		panic("Please, solve hCaptcha puzzle!")
	}

	levels := form.Value["compr-level"][0]
	compressionLevel, err := strconv.Atoi(levels) // check if it is in (20, 50, 80)
	if err != nil {
		return err
	}

	imageLocs := make(map[string]int)
	var files []*multipart.FileHeader
	for i := 1; i <= 5; i++ {
		frmField := fmt.Sprintf("image%d", i)
		file := form.File[frmField]
		if len(file) > 0 {
			imageLocs[file[0].Filename] = i
			files = append(files, file[0])
		}
	}

	var beforeSize int64 = 0
	var afterSize int64 = 0
	returnChan := make(chan ReturnVal, len(files))
	for _, file := range files {
		beforeSize += file.Size
		go ValidateAndProcess(file, compressionLevel, returnChan)
	}

	dlLinks := []string{"", "", "", "", ""}
	for i := 0; i < len(files); i++ {
		result := <-returnChan
		// stop if there is any error
		if result.statusCode != 201 {
			return c.JSON(fiber.Map{"status": result.statusCode, "message": result.statusMsg})
		}
		afterSize += result.afterSize
		if loc, ok := imageLocs[result.filename[9:]]; ok {
			dlLinks[loc-1] = "http://localhost:4000/images/" + result.filename
		}
	}

	messageBody := fmt.Sprintf(
		"Image compressed successfully. You saved around %.3f MB",
		float64((beforeSize-afterSize))/(1024*1024),
	)

	// Render index
	return c.Render("index", fiber.Map{
		"message":       messageBody,
		"hasLink1":      len(dlLinks[0]) > 0,
		"hasLink2":      len(dlLinks[1]) > 0,
		"hasLink3":      len(dlLinks[2]) > 0,
		"hasLink4":      len(dlLinks[3]) > 0,
		"hasLink5":      len(dlLinks[4]) > 0,
		"DownloadLink1": dlLinks[0],
		"DownloadLink2": dlLinks[1],
		"DownloadLink3": dlLinks[2],
		"DownloadLink4": dlLinks[3],
		"DownloadLink5": dlLinks[4],
	})
}
