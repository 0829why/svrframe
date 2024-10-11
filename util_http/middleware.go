package util_http

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"oversea-git.hotdogeth.com/poker/slots/svrframe/helper"
	"oversea-git.hotdogeth.com/poker/slots/svrframe/logx"

	"github.com/gin-gonic/gin"
)

type customResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func newResponseWriter(c *gin.Context) *customResponseWriter {
	return &customResponseWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
}

func (w customResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}
func (w customResponseWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func ResponseFailed(c *gin.Context, code int, msg ...string) {
	c.Set("error_code", code)
	//resp := gin.H{"error_code": code, "msg": cfgtable.GetErrorMsg(code), "time_unix": helper.GetNowTimestamp()}
	// _msg := comm.GetErrorMultitLangMsg(code)
	_msg := ""
	if len(msg) > 0 {
		_msg = msg[0]
	}
	resp := gin.H{"error_code": code, "msg": _msg, "time_unix": helper.GetNowTimestamp()}
	c.JSON(http.StatusOK, resp)
}

func ResponseSuccess(c *gin.Context, data interface{}) {
	resp := map[string]interface{}{
		"error_code": 0,
		"msg":        "success",
		"time_unix":  helper.GetNowTimestamp(),
		"data":       data,
	}
	c.JSON(http.StatusOK, resp)
}

func getRequestBodyBytes(c *gin.Context) []byte {
	body, _ := c.Request.GetBody()
	b, _ := io.ReadAll(body)
	return b
}

func ReBuildGetBody() gin.HandlerFunc {
	return func(c *gin.Context) {
		all, err := c.GetRawData()
		if err != nil {
			return
		}
		// 重写 GetBody 方法
		c.Request.GetBody = func() (io.ReadCloser, error) {
			buffer := bytes.NewBuffer(all)
			closer := io.NopCloser(buffer)
			return closer, nil
		}
		c.Request.Body, _ = c.Request.GetBody()

		c.Next()
	}
}

func SupportOptionsMethod() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "*")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")
		if c.Request.Method != "OPTIONS" {
			ctx := c.Request.Context()
			c.Request = c.Request.WithContext(ctx)

			c.Next()
		} else {
			c.AbortWithStatus(http.StatusNoContent)
		}
	}
}

func Process() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := helper.GetNowTime()

		blw := newResponseWriter(c)
		c.Writer = blw

		c.Next()

		body_req := getRequestBodyBytes(c)

		latencyTime := time.Since(startTime)
		reqMethod := c.Request.Method

		reqUri := c.Request.RequestURI

		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()

		logx.InfoF("%s %s from %s status[%s], [%v], headers = %s, request = %s, response = %s", reqMethod, reqUri, clientIP, http.StatusText(statusCode), latencyTime, getHeadersString(c), body_req, blw.body.String())
	}
}
