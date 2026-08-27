package postgres

import (
	"context"
	"io"
	"log"
	"os"
	"sync"
	"time"

	pgdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var (
	_db  *gorm.DB
	once sync.Once
)

func init() {
	// GORM's Scan path temporarily replaces the configured logger with its
	// package-level recorder. Configure that recorder too, otherwise Scan can
	// expand bound values even when the database logger is parameterized.
	gormlogger.RecorderParamsFilter = redactSQLParameters
}

func redactSQLParameters(_ context.Context, sql string, _ ...interface{}) (string, []interface{}) {
	return sql, nil
}

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
	return newGormDB(cfg, os.Stdout)
}

func newGormDB(cfg *Config, logOutput io.Writer) (*gorm.DB, error) {
	if logOutput == nil {
		logOutput = io.Discard
	}
	db, err := gorm.Open(pgdriver.Open(cfg.DSN), &gorm.Config{
		CreateBatchSize: cfg.CreateBatchSize,
		Logger: gormlogger.New(log.New(logOutput, "\r\n", log.LstdFlags), gormlogger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  gormlogger.Warn,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			Colorful:                  false,
		}),
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

// Get 返回已初始化的 DB 实例（与 mysql / redis / mongo 包保持一致的命名）
func Get() *gorm.DB {
	return _db
}
