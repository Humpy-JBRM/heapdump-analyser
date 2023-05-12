package api

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewRouter() (*gin.Engine, error) {
	router := gin.Default()
	router.MaxMultipartMemory = 8 << 30 // 8 GB
	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		log.Printf("endpoint %v %v %v %v\n", httpMethod, absolutePath, handlerName, nuHandlers)
	}

	// Make sure we propagate the headers so they can be logged
	router.Use(
		func(ctx *gin.Context) {
			ctx.Set("X-Real-Ip", ctx.Request.Header.Get("X-Real-Ip"))
		},
	)

	router.Use(CorsMiddleware())
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// your custom format
		return fmt.Sprintf("%s | %s | %s | %s %s | %d | %d | %s\n",
			param.TimeStamp.Format(time.RFC1123),
			param.ClientIP,
			param.Request.Header.Get("X-Real-Ip"),
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency.Microseconds(),
			param.ErrorMessage,
		)
	}))
	router.GET("/metrics", prometheusHandler())

	return router, nil
}

func prometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
