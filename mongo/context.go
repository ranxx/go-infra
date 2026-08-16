package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

type txMongo struct{}

// WithContext 将数据库连接注入上下文
func WithContext(ctx context.Context, tx *mongo.Database) context.Context {
	return context.WithValue(ctx, txMongo{}, tx)
}

// GetByContext 从上下文中获取数据库连接，如果没有则返回默认数据库连接
func GetByContext(ctx context.Context) *mongo.Database {
	tx := ctx.Value(txMongo{})
	if tx != nil {
		if db, ok := tx.(*mongo.Database); ok {
			return db
		}
	}
	return Get()
}

// IsTransaction 检查上下文是否在事务中
func IsTransaction(ctx context.Context) bool {
	_db := ctx.Value(txMongo{})
	if _db != nil {
		return true
	}
	return _db != nil
}
