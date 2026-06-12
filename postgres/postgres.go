package postgres

import (
	"sync"
	"time"

	pgdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	_db  *gorm.DB
	once sync.Once
)

// Init 初始化 PostgreSQL 客户端（单例）
func Init(cfg *Config) (*gorm.DB, error) {
	var err error
	once.Do(func() {
		db, e := NewGormDB(cfg)
		if e != nil {
			err = e
			return
		}
		_db = db
	})
	return _db, err
}

// NewGormDB 创建 GORM DB 客户端
func NewGormDB(cfg *Config) (*gorm.DB, error) {
	db, err := gorm.Open(pgdriver.Open(cfg.DSN), &gorm.Config{
		CreateBatchSize: cfg.CreateBatchSize,
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(cfg.IdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.MaxLifetime) * time.Second)
	return db, nil
}

// DB 返回已初始化的 DB 实例
func DB() *gorm.DB {
	return _db
}
