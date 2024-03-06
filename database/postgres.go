package database

import (
	"fmt"
	"log"
	"os"

	"PROJ/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB represents a Database instance
var DB *gorm.DB

// ConnectToDB connects the server with database
func ConnectToDB() {
	// โหลดข้อมูลการกำหนดค่าจากไฟล์ .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading env file \n", err)
	}

	// สร้าง DSN (Data Source Name) สำหรับเชื่อมต่อกับฐานข้อมูล PostgreSQL จากข้อมูลที่ได้จากไฟล์ .env
	dsn := fmt.Sprintf("host=db user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Bangkok", os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"), os.Getenv("DB_PORT"))

	// เชื่อมต่อกับฐานข้อมูล PostgreSQL โดยใช้ DSN ที่ได้จากขั้นตอนก่อนหน้า
	log.Print("Connecting to PostgreSQL DB...")
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	// ตรวจสอบความถูกต้องของการเชื่อมต่อ
	if err != nil {
		log.Fatal("Failed to connect to database. \n", err)
	}
	log.Print("Connected to Database")

	// turned on the loger on info mode
	DB.Logger = logger.Default.LogMode(logger.Info)

	log.Print("Running the migration ...")

	//  Auto migration
	DB.AutoMigrate(
		&models.User{},
		&models.Claims{},
		&models.Post{},
	)
}
