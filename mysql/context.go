package mysql

import (
	"context"

	"gorm.io/gorm"
)

type txMysql struct{}

// WithContext 将数据库连接注入上下文
func WithContext(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txMysql{}, tx)
}

// GetByContext 从上下文中获取数据库连接，如果没有则返回默认数据库连接
func GetByContext(ctx context.Context) *gorm.DB {
	tx := ctx.Value(txMysql{})
	if tx != nil {
		if db, ok := tx.(*gorm.DB); ok {
			return db
		}
	}
	return Get()
}

// IsTransaction 检查上下文是否在事务中
func IsTransaction(ctx context.Context) bool {
	_db := ctx.Value(txMysql{})
	if _db != nil {
		return true
	}
	return _db != nil
}
