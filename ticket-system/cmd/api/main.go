package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"ticket-system/internal/models"
	"ticket-system/internal/worker"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func main() {
	// 1. Config Yükle (önce mevcut dizin, sonra üst dizinler)
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

	// 2. DB Bağlantısı
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("HATA: DATABASE_URL boş olamaz!")
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("DB Hatası:", err)
	}

	DB.AutoMigrate(&models.Event{}, &models.Booking{})

	// 3. Worker Başlat
	worker.StartWorker(DB)

	// 4. Gin Setup
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

	// 5. HTTP Sunucusu Ayarları (Graceful Shutdown için gerekli)
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// Sunucuyu ayrı bir Goroutine'de başlatıyoruz
	go func() {
		log.Println("🚀 Sunucu 8080 portunda çalışıyor...")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Sunucu hatası: %s\n", err)
		}
	}()

	// 6. Kapanma Sinyalini Bekle (Ctrl+C)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("⚠️ Kapanma sinyali alındı, sunucu kapatılıyor...")

	// 7. Graceful Shutdown Süreci
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Sunucu zorla kapatıldı:", err)
	}

	log.Println("👋 Sunucu başarıyla kapatıldı. Görüşürüz!")
}
