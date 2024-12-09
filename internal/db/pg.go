package db

import (
	"errors"
	"fmt"
	"log"
	"stargazer/video-recording/config"
	error_handler "stargazer/video-recording/internal/error"
	"stargazer/video-recording/internal/models"
	"stargazer/video-recording/internal/utils"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var Pg Postgres

type Postgres struct {
	client *gorm.DB
}

func (pg *Postgres) Connect() error {
	var err error
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", config.App.DB_HOST, config.App.DB_USER, config.App.DB_PASS, config.App.DB_NAME, config.App.DB_PORT)
	fmt.Println(dsn)
	pg.client, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	return err
}

func (pg *Postgres) MigrateTables() error {
	err := pg.client.AutoMigrate(models.Interview{}, models.ParticipantStatus{}, models.ParticipantStatus{})
	if err != nil {
		utils.CheckError(err, error_handler.MigrationsFailed)
	}
	return err
}

func (pg *Postgres) GetClient() *gorm.DB {
	return pg.client
}

func (pg *Postgres) GetTxClient() (*gorm.DB, error) {

	db := pg.GetClient()

	tx := db.Begin()
	if tx.Error != nil {
		log.Fatal("Failed to start transaction: ", tx.Error)
		return nil, fmt.Errorf("failed to start transaction: %s", tx.Error)
	}

	return tx, nil
}

func HandleRollbackFail(tx *gorm.DB) {
	if r := recover(); r != nil {
		if err := tx.Rollback().Error; err != nil && !errors.Is(err, gorm.ErrInvalidTransaction) {
			log.Printf("Failed to rollback transaction during panic recovery: %v", err)
		}
		panic(r)
	} else if err := tx.Rollback().Error; err != nil && !errors.Is(err, gorm.ErrInvalidTransaction) {
		log.Printf("Failed to rollback transaction: %v", err)
	}
}

func ExecuteTransaction(fn func(*gorm.DB) error) error {
	tx, err := Pg.GetTxClient()
	if err != nil {
		return fmt.Errorf("error getting transaction client: %w", err)
	}

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}
