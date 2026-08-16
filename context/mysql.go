package context

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/ranxx/go-infra/mysql"
	"gorm.io/gorm"
)

type txdb struct{}

// WithMysql 将数据库连接注入上下文
func WithMysql(ctx Context, _db *gorm.DB) Context {
	return context.WithValue(ctx, txdb{}, _db)
}

// GetMysql 从上下文中获取数据库连接，如果没有则返回默认数据库连接
func GetMysql(ctx Context) *gorm.DB {
	if _, ok := ctx.(*gin.Context); ok {
		return GinCtxGetMysql(ctx.(*gin.Context))
	}
	_db := ctx.Value(txdb{})
	if _db != nil {
		if db, ok := _db.(*gorm.DB); ok {
			return db
		}
	}
	return mysql.Get()
}

// GinCtxWithMysql 将数据库连接注入 gin.Context
func GinCtxWithMysql(ctx *gin.Context, _db *gorm.DB) *gin.Context {
	ctx.Set("-gin-mysql-", _db)
	return ctx
}

// GinCtxGetMysql 从 gin.Context 中获取数据库连接，如果没有则返回默认数据库连接
func GinCtxGetMysql(ctx *gin.Context) *gorm.DB {
	_db, exists := ctx.Get("-gin-mysql-")
	if exists {
		if db, ok := _db.(*gorm.DB); ok {
			return db
		}
	}
	return mysql.Get()
}

// IsTransaction 检查上下文是否在事务中
func IsTransaction(ctx Context) bool {
	_db := ctx.Value(txdb{})
	if _db != nil {
		return true
	}
	if ctx, ok := ctx.(*gin.Context); ok {
		_, exists := ctx.Get("-gin-mysql-")
		return exists
	}
	return _db != nil
}
