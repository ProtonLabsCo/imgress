package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Image DB Model
type Image struct {
	ID         uint `gorm:"primaryKey"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ImageName  string
	ImageLink  string
	BeforeSize uint
	AfterSize  uint
	IpAddress  string
	IsDeleted  bool `gorm:"default:false"`
}

type DBClient struct {
	GDB        *gorm.DB
	ImageSaver chan []Image
}

func NewDBCLient() *DBClient {
	return &DBClient{
		GDB:        nil,
		ImageSaver: make(chan []Image),
	}
}

func (dbClient *DBClient) ConnectDB() {
	pgPort := os.Getenv("DATABASE_PORT")
	pgHost := os.Getenv("DATABASE_HOST")
	pgUser := os.Getenv("POSTGRES_USER")
	pgPassword := os.Getenv("POSTGRES_PASSWORD")
	pgName := os.Getenv("POSTGRES_DB")

	configData := fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=disable",
		pgUser,
		pgPassword,
		pgHost,
		pgPort,
		pgName,
	)

	var err error
	dbClient.GDB, err = gorm.Open(postgres.Open(configData), &gorm.Config{})
	if err != nil {
		log.Fatalln("Producer: Error Connecting to Database") // let it fail
	}
	log.Println("Producer: Connection Opened to Database")
}

func (dbClient *DBClient) Savior() {
	for {
		select {
		case images := <-dbClient.ImageSaver:
			go dbClient.saveToDB(images)
		default:
			// do nothing
		}
	}
}

func (dbClient *DBClient) saveToDB(images []Image) {
	if err := dbClient.GDB.Create(&images).Error; err != nil {
		log.Println("Producer: error while saving into DB: ", err)
		return
	}
	log.Println("Producer: successfully saved images into DB")
}
