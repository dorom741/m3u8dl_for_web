package middleware

import (
	"fmt"
	"m3u8dl_for_web/infra"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		c.Next() // 调用该请求的剩余处理程序
		stopTime := time.Since(startTime)
		spendTime := fmt.Sprintf("%d ms", stopTime.Milliseconds())
		//hostName, err := os.Hostname()
		//if err != nil {
		//    hostName = "Unknown"
		//}
		statusCode := c.Writer.Status()
		//userAgent := c.Request.UserAgent()
		dataSize := c.Writer.Size()
		if dataSize < 0 {
			dataSize = 0
		}
		method := c.Request.Method
		url := c.Request.RequestURI
		Log := infra.Logger.WithFields(logrus.Fields{
			//"HostName": hostName,
			"elapsed": spendTime,
			"path":      url,
			"Method":    method,
			"status":    statusCode,
			"ip": c.ClientIP(),
			//"DataSize": dataSize,
			//"UserAgent": userAgent,
		})
		if len(c.Errors) > 0 {
			Log.Error(c.Errors[0].Error())
		}
		if statusCode >= 500 {
			Log.Error()
		} else if statusCode >= 400 {
			Log.Warn()
		} else {
			Log.Info()
		}
	}
}
