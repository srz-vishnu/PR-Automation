package gormdb

import (
	"log"
	"pr-mail/app/domain"

	"gorm.io/gorm"
)

func Automigration(db *gorm.DB) error {
	if err := db.AutoMigrate(&domain.Credential{}); err != nil {
		log.Fatalf("Migration error for user:%v", err)
	}
	if err := db.AutoMigrate(&domain.Employee{}); err != nil {
		log.Fatalf("Migration error for employee:%v", err)
	}
	if err := db.AutoMigrate(&domain.PullRequest{}); err != nil {
		log.Fatalf("Migration error for PullRequest:%v", err)
	}
	if err := db.AutoMigrate(&domain.PRReport{}); err != nil {
		log.Fatalf("Migration error for Pr report:%v", err)
	}
	if err := db.AutoMigrate(&domain.PRSnapshot{}); err != nil {
		log.Fatalf("Migration error for Pr snapshot:%v", err)
	}
	return nil
}
