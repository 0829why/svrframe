package helper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/0829why/svrframe/constants"
	"github.com/0829why/svrframe/helper"
)

func strToUTCTimestamp(datetime string) uint32 {
	loc, _ := time.LoadLocation("Local")
	theTime, err := time.ParseInLocation(constants.TimeFormatString, datetime, loc)
	if err != nil {
		return 0
	}
	unixTime := uint32(theTime.Unix())
	return unixTime
}

// func GetGMTNow() time.Time {
// 	loc, _ := time.LoadLocation("")
// 	loca_t := helper.GetNowTime()
// 	fmt.Println("loca_t = ", loca_t)
// }

func Test_StrToUTCTimestamp(t *testing.T) {
	tm := helper.StrToUTCTimestamp("2023-08-17 17:56:32", time.Local)
	fmt.Println(tm)

	// loc, _ := time.LoadLocation("")
	loca_t := helper.GetNowTime()
	fmt.Println("loca_t = ", loca_t)

	loca_gmt := loca_t.In(time.UTC)
	fmt.Println("loca_gmt = ", loca_gmt)
}
