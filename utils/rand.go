package utils

import (
	"math/rand"
	"time"
)

var (
	source  = rand.NewSource(time.Now().UnixNano())
	letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// Intn int
func Intn() int {
	return int(source.Int63())
}

// Perm 生成随机排列
func Perm(n int) []int {
	return rand.Perm(n)
}

// RandString 生成指定长度的随机字符串
func RandString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
