package utils

import (
	"math/rand"
	"strconv"

	"github.com/shopspring/decimal"
)

// Int64RandMinMax rand
func Int64RandMinMax(min, max int64) int64 {
	if max-min <= 0 {
		return min
	}
	return rand.Int63()%(max-min+1) + min
}

// Compare comp
type Compare interface {
	int64 | int32 | int | int16 | int8 | uint64 | uint32 | uint | uint16 | Compare2 | float64 | float32
}

// Compare2 comp2
type Compare2 interface {
	int64 | int32 | int | int16 | int8
}

// Min min
func Min[T Compare](a, b T) T {
	if a > b {
		return b
	}
	return a
}

// Max max
func Max[T Compare](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// Abs abs
func Abs[T Compare2](a T) T {
	if a > 0 {
		return a
	}
	return a * (T)(-1)
}

// StrCompare 字符串的比较
func StrCompare(a, b string) bool {
	aa, _ := strconv.ParseInt(a, 10, 64)
	bb, _ := strconv.ParseInt(b, 10, 64)
	return aa < bb
}

// MinDecimal 返回两个 decimal.Decimal 中的较小值
func MinDecimal(a, b decimal.Decimal) decimal.Decimal {
	if a.LessThan(b) {
		return a
	}
	return b
}

// MaxDecimal 返回两个 decimal.Decimal 中的较大值
func MaxDecimal(a, b decimal.Decimal) decimal.Decimal {
	if a.GreaterThan(b) {
		return a
	}
	return b
}
