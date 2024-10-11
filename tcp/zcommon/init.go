package zcommon

import (
	"encoding/binary"
	"time"

	"oversea-git.hotdogeth.com/poker/slots/svrframe/logx"
	"oversea-git.hotdogeth.com/poker/slots/svrframe/tcp/utils"

	"golang.org/x/time/rate"
)

var (
	ByteOrder binary.ByteOrder
)

const (
	Limiter_limit          rate.Limit    = 20
	Limiter_bucket         int           = 20
	Limiter_Timeout        time.Duration = time.Second * 1
	Limiter_FailedMaxCount               = 20 //超过多少次限流成功,判定为非法连接
)

func init() {
	ByteOrder = binary.LittleEndian
}

func PrintLogo() {
	logx.DebugF("[Zinx] Version: %s, MaxConn: %d, MaxPacketSize: %d\n",
		utils.GlobalObject.Version,
		utils.GlobalObject.MaxConn,
		utils.GlobalObject.MaxPacketSize)
}
