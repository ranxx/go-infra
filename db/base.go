package db

import (
	"time"
)

// Base base
type Base struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement;comment:主键"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;comment:创建时间"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime;comment:更新时间"`
}

// BaseV2 base
type BaseV2 struct {
	ID        string    `gorm:"column:id;primaryKey;comment:主键"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;comment:创建时间"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime;comment:更新时间"`
}
