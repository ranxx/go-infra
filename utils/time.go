package utils

import (
	"time"
)

// format
const (
	NormalFormat = "2006-01-02 15:04:05"
	locName      = "Asia/Shanghai"
)

var loc, _ = time.LoadLocation(locName)

// UnixTimeNormalFormat unix 2006-01-02 15:04:05 format
func UnixTimeNormalFormat(sec int64) string {
	return TimeNormalFormat(time.Unix(sec, 0))
}

// TimeNormalFormat time normal format
func TimeNormalFormat(t time.Time) string {
	return t.Format(NormalFormat)
}

// NormalParseInLocal date parse in local
func NormalParseInLocal(date string) (time.Time, error) {
	return time.ParseInLocation(NormalFormat, date, time.Local)
}
