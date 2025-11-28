package worker

import (
	"log"
	"ticket-system/internal/models"
	"time"

	"gorm.io/gorm"
)

// Bilet isteği için basit bir veri yapısı
type TicketRequest struct {
	EventID string
	UserID  string
}

// 1. Buffered Channel: Aynı anda 1000 istek kuyrukta bekleyebilir
// Bu kanal dolarsa, API yeni istekleri reddetmeye başlar (Backpressure)
var TicketQueue = make(chan TicketRequest, 1000)

// Worker Fonksiyonu: Kuyruğu dinleyip sırayla işleyen eleman
func StartWorker(db *gorm.DB) {
	go func() {
		log.Println("👷 Bilet Worker'ı iş başı yaptı, kuyruk dinleniyor...")
		
		for req := range TicketQueue {
			processTicket(db, req)
		}
	}()
}

// Veritabanı işlemini yapan fonksiyon
func processTicket(db *gorm.DB, req TicketRequest) {
	// Transaction başlatıyoruz (Ya hepsi olur ya hiçbiri)
	tx := db.Begin()

	var event models.Event
	
	// 1. Etkinliği bul
	if err := tx.First(&event, "id = ?", req.EventID).Error; err != nil {
		log.Printf("❌ Etkinlik bulunamadı: %v\n", req.EventID)
		tx.Rollback()
		return
	}

	// 2. Kota kontrolü (Memory'de değil, güncel DB verisiyle)
	if event.AvailableQuota <= 0 {
		log.Printf("⚠️ KOTA DOLDU! Kullanıcı: %s işlem yapamadı.\n", req.UserID)
		tx.Rollback()
		return
	}

	// 3. Kotayı düş
	event.AvailableQuota -= 1
	if err := tx.Save(&event).Error; err != nil {
		log.Println("❌ Kota güncellenemedi")
		tx.Rollback()
		return
	}

	// 4. Booking kaydı oluştur
	booking := models.Booking{
		EventID: req.EventID,
		UserID:  req.UserID,
		Status:  "SUCCESS",
	}
	
	if err := tx.Create(&booking).Error; err != nil {
		log.Println("❌ Booking oluşturulamadı")
		tx.Rollback()
		return
	}

	// Her şey yolundaysa commitle
	tx.Commit()
	log.Printf("✅ Bilet Satıldı! Kalan: %d | Alan: %s\n", event.AvailableQuota, req.UserID)
	
	// Veritabanını yormamak için çok ufak bir yapay gecikme (Simülasyon için)
	time.Sleep(50 * time.Millisecond)
}