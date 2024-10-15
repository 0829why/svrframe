package util_http

import (
	"fmt"
	"net/http"

	"github.com/0829why/svrframe/constants"
	"github.com/0829why/svrframe/logx"

	"github.com/gin-gonic/gin"
)

func update() {
	go func() {
		exit_ch := constants.GetServiceStopListener().AddListener()
		constants.GetServiceStopWaitGroup().Add(1)
		<-exit_ch.Done()
		httpServer.Close()
		httpServer = nil
		constants.GetServiceStopWaitGroup().Done()
	}()
}

func StartHttpService(g *gin.Engine, listen_port uint16) {
	port := listen_port
	if port == 0 {
		port = 80
	}
	addr := fmt.Sprintf(":%d", port)
	httpServer = &http.Server{
		Addr:    addr,
		Handler: g,
	}

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logx.ErrorF("%v", err)
		}
	}()

	logx.InfoF("HTTP SERVER [ %s ] RUNNING", constants.Service_Type)

	update()
}
