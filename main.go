package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/resably/go-rest-api/config"
	"github.com/resably/go-rest-api/routes"
	"github.com/resably/go-rest-api/utils"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// 1. Veritabanı bağlantısını kur
	config.ConnectDatabase()

	// 2. JWT konfigürasyonunu yükle
	utils.InitJWTConfig()

	// 3. Router'ı ayarla
	r := routes.SetupRoutes(config.DB)

	// 4. Sunucuyu başlat
	port := ":8080"
	fmt.Printf("Server running on port %s\n", port)
	fmt.Println("API Versions:")
	fmt.Println("- /api/v1")
	r.Run(port)
}
