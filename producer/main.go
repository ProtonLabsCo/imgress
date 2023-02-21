package main

import (
	"fmt"
	"log"
	"mime/multipart"
	"strconv"
	"strings"

	"imgress-producer/database"
	"imgress-producer/messageq"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/template/html"
	"github.com/google/uuid"
)

type custom_handler struct {
	DBCl      *database.DBClient
	RMQPubCl  *messageq.RMQPubClient
	RMQConsCl *messageq.RMQConsClient
}

func newHandler(dbc *database.DBClient, pubc *messageq.RMQPubClient, consc *messageq.RMQConsClient) custom_handler {
	return custom_handler{
		DBCl:      dbc,
		RMQPubCl:  pubc,
		RMQConsCl: consc,
	}
}

func main() {
	dbClient := database.NewDBCLient()
	dbClient.ConnectDB()
	dbClient.GDB.AutoMigrate(&database.Image{})
	go dbClient.Savior()

	pubClient := messageq.NewPublisher()
	consClient := messageq.NewConsumer()

	pubClient.Connect()
	go pubClient.Publisher()
	defer pubClient.Chan.Close()
	defer pubClient.Conn.Close()

	consClient.Connect()
	go consClient.Consumer()
	defer consClient.Chan.Close()
	defer consClient.Conn.Close()

	hndlr := newHandler(dbClient, pubClient, consClient)

	engine := html.New("./static", ".html")

	app := fiber.New(fiber.Config{
		Views:     engine,
		BodyLimit: 100 * 1024 * 1024, // this is the default limit of 100MB
	})

	app.Use(cors.New())

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{})
	})

	app.Post("/", hndlr.handleFileupload)

	log.Fatal(app.Listen(":8080"))
}

func (hndlr custom_handler) handleFileupload(c *fiber.Ctx) error {
	form, err := c.MultipartForm()
	if err != nil {
		return c.Render("index", fiber.Map{"message": err})
	}

	hcaptchaPass := HandleCaptcha(form.Value["h-captcha-response"][0])
	if !hcaptchaPass {
		return c.Render("index", fiber.Map{"message": "Please, solve hCaptcha puzzle!"})
	}

	levels := form.Value["compr-level"][0]
	compressionLevel, err := strconv.Atoi(levels) // check if it is in (20, 50, 80)
	if err != nil {
		return c.Render("index", fiber.Map{"message": err})
	}

	imageLocs := make(map[string]int)
	var files []*multipart.FileHeader
	for i := 1; i <= 5; i++ {
		frmField := fmt.Sprintf("image%d", i)
		file := form.File[frmField]
		if len(file) > 0 {
			orgFilename := file[0].Filename
			uuidStr := strings.Replace(uuid.New().String(), "-", "", -1)
			uniqueFilename := uuidStr[len(uuidStr)-8:] + "_" + orgFilename
			file[0].Filename = uniqueFilename

			imageLocs[file[0].Filename] = i
			files = append(files, file[0])
		}
	}

	// Validate and send each image separately
	uuid_str := strings.Replace(uuid.New().String(), "-", "", -1)
	respQueueName := uuid_str[len(uuid_str)-8:]

	sttsCode, sttsMsg, beforeSize, beforeSizeSum := ValidateAndPublish(files, compressionLevel, respQueueName, hndlr.RMQPubCl)
	if sttsCode != 201 {
		return c.Render("index", fiber.Map{"message": sttsMsg})
	}

	// PART-2: Start Consumer..
	hndlr.RMQConsCl.Fanus[respQueueName] = make(chan []messageq.ConfirmMsgBody)
	hndlr.RMQConsCl.ConfData <- messageq.ConfirmExpected{len(files), respQueueName}
	var confirmations []messageq.ConfirmMsgBody
	completed := false
	for !completed {
		select {
		case confirmations = <-hndlr.RMQConsCl.Fanus[respQueueName]:
			log.Println("Producer: confirmation completed!")
			completed = true
		default:
			// Do nothing
		}
	}

	var afterSizeSum uint = 0
	dlLinks := []string{"", "", "", "", ""}
	var images []database.Image
	for _, resultConf := range confirmations {
		if resultConf.Filename == "error" {
			return c.Render("index", fiber.Map{"message": "Internal Error."})
		}
		afterSizeSum += resultConf.AfterSize
		loc, ok := imageLocs[resultConf.Filename]
		if ok {
			dlLinks[loc-1] = resultConf.FileLink
			image := database.Image{
				ImageName:  resultConf.Filename,
				ImageLink:  resultConf.FileLink,
				BeforeSize: beforeSize[loc-1],
				AfterSize:  resultConf.AfterSize,
				IpAddress:  c.IP(),
			}
			images = append(images, image)
		}
	}
	close(hndlr.RMQConsCl.Fanus[respQueueName])

	// save into DB async (user should not wait for db save)
	hndlr.DBCl.ImageSaver <- images

	messageBody := fmt.Sprintf(
		"Image compressed successfully. You saved around %.3f MB",
		float64((beforeSizeSum-afterSizeSum))/(1024*1024),
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
