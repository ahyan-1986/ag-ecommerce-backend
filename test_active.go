package main

import (
	"fmt"
	"log"

	"github.com/ahyan-1986/ag-ecommerce-backend/internal/config"
	"github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	db, err := gorm.Open(postgres.Open(cfg.Database.GetDSN()), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	var users []domain.User
	db.Where("role = ?", "customer").Order("created_at desc").Limit(5).Find(&users)
	for _, u := range users {
		fmt.Printf("Customer Email: %s, IsActive: %v\n", u.Email, u.IsActive)
	}
}
