package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var GDB *gorm.DB

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

func ConnectDB() {
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
	GDB, err = gorm.Open(postgres.Open(configData), &gorm.Config{})
	if err != nil {
		log.Fatalln("Producer: Error Connecting to Database") // let it fail
	}
	log.Println("Producer: Connection Opened to Database")
}
