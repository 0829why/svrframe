package constants

import (
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	serviceStopListener  *IListenerManager
	serviceStopWaitGroup sync.WaitGroup
	Service_Type         string
	Service_mode_map     map[string]string

	ProjectName string //项目名称
	ServiceHost string //服务地址
)

var signals chan os.Signal

func init() {
	serviceStopListener = NewListenerManager()
	serviceStopWaitGroup = sync.WaitGroup{}
	Service_Type = "unknow"

	Service_mode_map = map[string]string{
		ServiceMode_TEST:   ServiceMode_TEST,
		ServiceMode_FORMAL: ServiceMode_FORMAL,
	}

	ServiceHost = getServiceHost()

	signals = make(chan os.Signal, 2)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
}

func GetServiceStopListener() *IListenerManager {
	return serviceStopListener
}
func GetServiceStopWaitGroup() *sync.WaitGroup {
	return &serviceStopWaitGroup
}

func getServiceHost() string {
	addr := os.Getenv(Env_ServiceHost)
	if len(addr) > 0 && net.ParseIP(addr) != nil {
		return addr
	}

	addrs, err := net.InterfaceAddrs()
	if err == nil {
		for _, address := range addrs {
			// 检查ip地址判断是否回环地址
			if ipnet, ok := address.(*net.IPNet); ok && ipnet.IP.IsGlobalUnicast() {
				if ipnet.IP.To4() != nil {
					return ipnet.IP.String()
				}
			}
		}
	}

	return "127.0.0.1"
}
