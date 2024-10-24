package util_http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/0829why/svrframe/helper"
)

var (
	httpServer *http.Server
)

func init() {
	httpServer = nil
}

func getHeadersString(c *gin.Context) string {
	headers := map[string]string{}
	device_id := c.GetHeader("device_id")
	channel := c.GetHeader("channel")
	sysplatform := c.GetHeader("sysplatform")
	headers["device_id"] = device_id
	headers["channel"] = channel
	headers["sysplatform"] = sysplatform

	return helper.ToJson(headers)
}
