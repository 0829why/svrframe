package helper

import (
	"time"

	"github.com/0829why/svrframe/constants"
)

const Second_one_day = uint32(24 * 3600)
const Millisecond_one_day = int64(Second_one_day) * 1000
const Second_one_hour = uint32(3600)
const Second_one_week = Second_one_day * 7

var (
	local_timestamp_milli       int64
	local_timestamp_milli_begin int64
)

func init() {
	local_timestamp_milli_begin = time.Now().UnixMilli()
	local_timestamp_milli = local_timestamp_milli_begin
}

func GetNowTime() time.Time {
	local_milli := local_timestamp_milli + (time.Now().UnixMilli() - local_timestamp_milli_begin)
	t := time.UnixMilli(local_milli)
	t = t.UTC()
	return t
}
func GetNowTimestamp() uint32 {
	return uint32(GetNowTime().Unix())
}
func GetNowTimestampMilli() int64 {
	return GetNowTime().UnixMilli()
}

func ModifyTime(tMilli int64) time.Time {
	local_timestamp_milli_begin = time.Now().UnixMilli()
	if tMilli == 0 {
		local_timestamp_milli = local_timestamp_milli_begin
	} else {
		local_timestamp_milli = tMilli
	}

	return GetNowTime()
}

// 时间字符串转时间戳(默认UTC时间)
func StrToUTCTimestamp(datetime string, local *time.Location) uint32 {
	loc := local
	if loc == nil {
		loc = time.UTC
	}
	theTime, err := time.ParseInLocation(constants.TimeFormatString, datetime, loc)
	if err != nil {
		theTime, err = time.ParseInLocation(constants.TimeFormatStringShort, datetime, loc)
		if err != nil {
			return 0
		}
	}
	unixTime := uint32(theTime.Unix())
	return unixTime
}

// 获取指定时间的时间戳
func GetAppointDate(year int, month time.Month, day, hour, min, sec int, local *time.Location) time.Time {
	loc := local
	if loc == nil {
		loc = time.UTC
	}
	t := time.Date(year, month, day, hour, min, sec, 0, loc)
	return t
}

// 获取当天0点时间戳
func GetToday0ClockTimestamp() uint32 {
	currentTime := GetNowTime()
	t := GetAppointDate(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, currentTime.Location()).Unix()
	return uint32(t)
}
