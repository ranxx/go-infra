package context

import (
	stlcontext "context"

	"github.com/gin-gonic/gin"
	"github.com/ranxx/go-infra/postgres"
	"gorm.io/gorm"
)

type pgTxdb struct{}

// WithPostgres 将 PostgreSQL 连接注入上下文。
// 用于在请求 / 事务范围内覆盖全局单例 postgres.Get()。
func WithPostgres(ctx Context, db *gorm.DB) Context {
	return stlcontext.WithValue(ctx, pgTxdb{}, db)
}

// GetPostgres 从上下文中获取 PostgreSQL 连接；没有则回退到全局单例 postgres.Get()。
// 若 ctx 来自 gin.Context，则尝试从 gin keys 取，便于 HTTP handler 使用同一接口。
func GetPostgres(ctx Context) *gorm.DB {
	if gc, ok := ctx.(*gin.Context); ok {
		return GinCtxGetPostgres(gc)
	}
	v := ctx.Value(pgTxdb{})
	if v != nil {
		if db, ok := v.(*gorm.DB); ok {
			return db
		}
	}
	return postgres.Get()
}

// GinCtxWithPostgres 将 PostgreSQL 连接注入 gin.Context
func GinCtxWithPostgres(ctx *gin.Context, db *gorm.DB) *gin.Context {
	ctx.Set("-gin-postgres-", db)
	return ctx
}

// GinCtxGetPostgres 从 gin.Context 中获取 PostgreSQL 连接，没有则回退到全局单例
func GinCtxGetPostgres(ctx *gin.Context) *gorm.DB {
	v, ok := ctx.Get("-gin-postgres-")
	if ok {
		if db, ok := v.(*gorm.DB); ok {
			return db
		}
	}
	return postgres.Get()
}

// IsPostgresTransaction 检查上下文是否在 PG 事务中（即是否注入了事务级 *gorm.DB）
func IsPostgresTransaction(ctx Context) bool {
	v := ctx.Value(pgTxdb{})
	if v != nil {
		return true
	}
	if gc, ok := ctx.(*gin.Context); ok {
		_, exists := gc.Get("-gin-postgres-")
		return exists
	}
	return false
}