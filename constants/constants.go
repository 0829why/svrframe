package constants

import (
	"context"
	"net"
	"os"
	"runtime"
	"strings"
	"syscall"

	"google.golang.org/grpc/peer"
)

type (
	ContextKey string
)

const (
	RpcSendRecvMaxSize = 1024 * 1024 * 64
)

const (
	RequestParam ContextKey = "requestParam"
)

const (
	ServiceMode_TEST   = "test"   //测试环境
	ServiceMode_FORMAL = "formal" //正式环境
)

const (
	RuntimeMode_Debug   = "debug"
	RuntimeMode_Release = "release"
)

const (
	Env_ServiceHost = "service_host"
	Env_Gopath      = "GOPATH"
	Env_RuntimeMode = "runtime_mode"
	Env_ServiceMode = "service_mode"
	Env_Coredump    = "coredump"
	Env_LogLevel    = "log_level"
	Env_LogPath     = "log_path"
	Env_LogRuntime  = "log_runtime"
	Env_PidPath     = "pid_path"
)

const (
	TimeFormatString      = "2006-01-02 15:04:05"
	TimeFormatStringShort = "2006-01-02"
	TimeFormatYMD         = "20060102"
	TimeFormatYM          = "200601"
)

const (
	TenThousandthRatio = 0.0001
)

const (
	SystemStatus_Normal   = iota
	SystemStatus_Maintain //维护
)

var (
	//服务状态
	system_status int32
)

func SetSystemStatus(status int32) {
	system_status = status
}

func GetSystemStatus() int32 {
	return system_status
}

// /////////////////////////////////////////////////////////////////////////////
func GetSystem() string {
	return runtime.GOOS
}

func GetServiceMode() string {
	mode := strings.ToLower(os.Getenv(Env_ServiceMode))
	if len(mode) <= 0 || !IsValidServiceMode(mode) {
		mode = ServiceMode_TEST
	}
	return mode
}

func IsValidServiceMode(mode string) bool {
	_, ok := Service_mode_map[mode]
	return ok
}

func IsCoredump() bool {
	coredump_mode := os.Getenv(Env_Coredump)
	return len(coredump_mode) > 0
}

func IsDebug() bool {
	runtime_mode := os.Getenv(Env_RuntimeMode)
	return strings.ToLower(runtime_mode) == RuntimeMode_Debug
}

func GetServiceHost() string {
	return ServiceHost
}

func GetPeerAddr(ctx context.Context) string {
	var addr string
	if pr, ok := peer.FromContext(ctx); ok {
		if tcpAddr, ok := pr.Addr.(*net.TCPAddr); ok {
			addr = tcpAddr.IP.String()
		} else {
			addr = pr.Addr.String()
		}
	}
	return addr
}

func ProgramExit() {
	signals <- syscall.SIGTERM
}

func GetSignals() chan os.Signal {
	return signals
}
