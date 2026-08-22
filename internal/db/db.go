package db

import (
	"fmt"
	"github.com/weeb-vip/anime-sync/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"time"
)

type DB struct {
	DB *gorm.DB
}

func NewDB(cfg config.DBConfig) *DB {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DataBase, cfg.SSLMode)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	sqlDB, err := db.DB()
	if err != nil {
		panic("failed to get database connection")
	}

	// Set maximum number of open connections
	// This prevents too many connections to the database
	sqlDB.SetMaxOpenConns(25)

	// Set maximum number of idle connections
	// This maintains a pool of reusable connections
	sqlDB.SetMaxIdleConns(10)

	// Set maximum lifetime of a connection
	// Kept below any server-side idle timeout so the pool never hands out a
	// connection the server has already closed.
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// Set maximum idle time for a connection
	// This helps clean up idle connections
	sqlDB.SetConnMaxIdleTime(90 * time.Second)

	return &DB{DB: db}
}
