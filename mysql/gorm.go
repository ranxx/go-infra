package mysql

import (
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	_db  *gorm.DB
	once sync.Once
)

// Init 初始化 MySQL 客户端
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

// NewGormDB 创建 GORM DB 客户端，连接失败时返回 nil 但不报错
func NewGormDB(cfg *Config) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		CreateBatchSize: cfg.CreateBatchSize,
	})
	if err != nil {
		return nil, err
	}

	// 获取底层 sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 配置连接池
	sqlDB.SetMaxIdleConns(cfg.IdleConns)                                   // 空闲连接数
	sqlDB.SetMaxOpenConns(cfg.MaxConns)                                    // 最大连接数
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.MaxLifetime) * time.Second) // 连接最大生命周期
	return db, nil
}
