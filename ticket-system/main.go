package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Veritabanı modellerimizi Go struct olarak da tanımlayalım
type Event struct {
	ID             string `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name           string
	TotalQuota     int
	AvailableQuota int
	Version        int
}

type Booking struct {
	ID      string `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	EventID string
	UserID  string
}

var DB *gorm.DB

func main() {
	// .env dosyasından şifreleri okumak için 
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found")
	}

	dsn := os.Getenv("DATABASE_URL")
    // dsn := "postgres://postgres:sifreniz@db.x.supabase.co:5432/postgres"

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Veritabanına bağlanılamadı:", err)
	}

	fmt.Println("🚀 Supabase bağlantısı başarılı!")

	// Tablodaki veriyi kontrol edelim
	var event Event
	result := DB.First(&event)
	if result.Error != nil {
		log.Println("Etkinlik bulunamadı, önce SQL ile veri eklediğinden emin ol.")
	} else {
		fmt.Printf("Hedef Etkinlik: %s | Kalan Kota: %d\n", event.Name, event.AvailableQuota)
	}
}