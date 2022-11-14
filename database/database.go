package database

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var GDB *gorm.DB

// Image DB Model
type Image struct {
	ID         uint `gorm:"primarykey"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ImageName  string
	ImageLink  string
	BeforeSize uint
	AfterSize  uint
	IpAddress  string
	IsDeleted  bool `gorm:"default:false"`
}

func ConnectDB() error {
	err := godotenv.Load()
	if err != nil {
		return err
	}

	pgPort := os.Getenv("DB_PORT")
	pgHost := os.Getenv("DB_HOST")
	pgUser := os.Getenv("DB_USER")
	pgPassword := os.Getenv("DB_PASSWORD")
	pgName := os.Getenv("DB_NAME")

	configData := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		pgHost,
		pgUser,
		pgPassword,
		pgName,
		pgPort,
	)

	GDB, err = gorm.Open(postgres.Open(configData), &gorm.Config{})
	if err != nil {
		return err
	}

	fmt.Println("Connection Opened to Database")
	return nil
}
