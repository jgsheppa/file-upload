package models

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Services struct {
	File FileService
	db   *gorm.DB
}

func NewServices(connectionInfo string) *Services {
	devLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			Colorful:                  true,        // Disable color
		},
	)

	db, err := gorm.Open(sqlite.Open(connectionInfo), &gorm.Config{Logger: devLogger})
	if err != nil {
		panic(err)
	}
	err = db.AfterInitialize(db.Exec(`PRAGMA journal_mode = wal;`))
	if err != nil {
		panic(err)
	}
	err = db.AfterInitialize(db.Exec(`PRAGMA foreign_keys = ON;`))
	if err != nil {
		panic(err)
	}

	return &Services{
		File: NewFileService(db),
		db:   db,
	}
}

// Will attempt to automigrate all database tables
func (s *Services) AutoMigrate() error {
	return s.db.AutoMigrate(&File{})
}
