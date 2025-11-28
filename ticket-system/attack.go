package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Hedef API ve Event ID (Senin Event ID'ni buraya yapıştır)
const (
	url     = "http://localhost:8080/buy"
	eventID = "155ff34d-51ec-4053-841e-a6cc24253256" // <-- BURAYI KENDİ EVENT ID'N İLE KONTROL ET
)

func main() {
	var wg sync.WaitGroup
	totalRequests := 100 // Toplam gönderilecek istek sayısı (Kotadan az olsun ki sonucu görelim)

	fmt.Printf("🚀 Saldırı başlıyor! %d istek gönderilecek...\n", totalRequests)
	startTime := time.Now()

	for i := 0; i < totalRequests; i++ {
		wg.Add(1)
		// Her isteği ayrı bir goroutine içinde atıyoruz (Paralel saldırı)
		go func(i int) {
			defer wg.Done()

			// Her istek için farklı bir user_id (UUID formatında değil ama test için string gönderiyorsan,
			// DB tarafında user_id'yi UUID yerine VARCHAR yapman gerekebilir.
			// Ya da buraya rastgele UUID üreten bir kod eklemeliyiz.)
			// Şimdilik DB'de hata almamak için sabit geçerli bir UUID kullanalım veya
			// DB'deki user_id sütununu text'e çevirelim.
			// DEMO İÇİN EN KOLAYI: Her seferinde aynı kullanıcı alıyor gibi yapalım.
			payload := fmt.Sprintf(`{"event_id": "%s", "user_id": "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"}`, eventID)

			resp, err := http.Post(url, "application/json", strings.NewReader(payload))
			if err != nil {
				fmt.Printf("İstek hatası: %v\n", err)
				return
			}
			resp.Body.Close()
		}(i)
	}

	wg.Wait()
	fmt.Printf("🏁 Test bitti! Geçen süre: %v\n", time.Since(startTime))
}
