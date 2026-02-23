package main

import (
	"github.com/your-username/gopher-social-ecom/pkg/config"
	"github.com/your-username/gopher-social-ecom/pkg/utils"
	"log"
)

func main() {
	// 1. Load configuration from .env
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Could not load config: %v", err)
	}

	// 2. Connect to Database (Postgres)
	db, err := utils.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Could not connect to DB: %v", err)
	}

	log.Println("✅ Server initialized successfully!")
	log.Println("✅ Database connection established!")
}
