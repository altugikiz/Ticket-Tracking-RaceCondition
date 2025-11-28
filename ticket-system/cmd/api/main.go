package main

import (
	"log"
	"net/http"
	"os"
	"ticket-system/internal/models"
	"ticket-system/internal/worker"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func main() {
	// 1. .env Dosyasını Yükle (proje kök dizininden)
	// Önce mevcut dizini dene, sonra üst dizinleri dene
	envPaths := []string{".env", "../../.env"}
	envLoaded := false

	for _, path := range envPaths {
		if err := godotenv.Load(path); err == nil {
			log.Printf("✅ .env dosyası yüklendi: %s\n", path)
			envLoaded = true
			break
		}
	}

	if !envLoaded {
		log.Println("Uyarı: .env dosyası bulunamadı")
	}

	// 2. Veritabanı Bağlantısı
	dsn := os.Getenv("DATABASE_URL")

	if dsn == "" {
		log.Fatal("HATA: DATABASE_URL ortam değişkeni ayarlanmamış!")
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("DB Hatası:", err)
	}

	// Auto Migrate
	DB.AutoMigrate(&models.Event{}, &models.Booking{})

	// Worker Başlat
	worker.StartWorker(DB)

	// Gin Başlat
	r := gin.Default()

	r.POST("/buy", func(c *gin.Context) {
		var body struct {
			EventID string `json:"event_id"`
			UserID  string `json:"user_id"`
		}

		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veri"})
			return
		}

		select {
		case worker.TicketQueue <- worker.TicketRequest{EventID: body.EventID, UserID: body.UserID}:
			c.JSON(http.StatusOK, gin.H{"message": "İstek kuyruğa alındı", "status": "pending"})
		default:
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Sistem çok yoğun"})
		}
	})

	log.Println("🚀 Sunucu 8080 portunda çalışıyor...")
	r.Run(":8080")
}
