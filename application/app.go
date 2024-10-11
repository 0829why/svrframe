package application

import (
	"fmt"
	"net/http"
	"os"
	"syscall"
	"time"

	"oversea-git.hotdogeth.com/poker/slots/svrframe/config"
	"oversea-git.hotdogeth.com/poker/slots/svrframe/constants"
	"oversea-git.hotdogeth.com/poker/slots/svrframe/helper"
	"oversea-git.hotdogeth.com/poker/slots/svrframe/logx"
)

func InitProgram(service_type string) bool {
	constants.Service_Type = service_type
	setTitle(service_type)

	if err := config.ParseConfig(); err != nil {
		return false
	}

	logx.InitLogx()

	_logo := logo + "\n"
	_logo += topLine + "\n"
	_logo += fmt.Sprintf("%s [coder] martin                                    %s", borderLine, borderLine) + "\n"
	_logo += fmt.Sprintf("%s [time] 2022-01-29                                 %s", borderLine, borderLine) + "\n"
	_logo += bottomLine + "\n"

	logx.InfoF("%s", _logo)

	return true
}

func ProgramRunning() {
	defer fmt.Println("server close pid = ", os.Getpid())
	defer func() {
		constants.GetServiceStopListener().NotifyAllListeners()
		constants.GetServiceStopWaitGroup().Wait()
		constants.GetServiceStopListener().Clear()
	}()
	defer notifySubKill()

	logx.InfoF("服务启动成功")
	if !writePid() {
		return
	}

	for {
		select {
		case sig, ok := <-constants.GetSignals():
			if ok {
				logx.InfoF("signal receive: %v", sig)
				switch sig {
				case syscall.SIGINT:
					return
				case syscall.SIGTERM:
					return
				}
			}
		default:
			helper.GetGlobalTimer().Update()
		}

		time.Sleep(time.Millisecond * 10)
	}
}

func EnablePprof(port int) {
	go http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), nil)
}
