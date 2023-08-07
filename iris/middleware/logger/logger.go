package logger

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"github.com/kataras/pio"
	"net/http"
	"os"
	"strconv"
	"time"
)

func init() {
	context.SetHandlerName("go-utils/iris/middleware/logger.*", "go-utils.logger")
}

func StatusColor(status string) string {
	code, _ := strconv.Atoi(status)

	switch {
	case code >= http.StatusOK && code < http.StatusMultipleChoices:
		return pio.Rich(status, pio.Green)
	case code >= http.StatusMultipleChoices && code < http.StatusBadRequest:
		return pio.Rich(status, pio.White)
	case code >= http.StatusBadRequest && code < http.StatusInternalServerError:
		return pio.Rich(status, pio.Yellow)
	default:
		return pio.Rich(status, pio.Red)
	}
}

func MethodColor(method string) string {
	switch method {
	case "GET":
		return pio.Rich(method, pio.Blue)
	case "POST":
		return pio.Rich(method, pio.Magenta)
	case "PUT":
		return pio.Rich(method, pio.Yellow)
	case "DELETE":
		return pio.Rich(method, pio.Red)
	case "PATCH":
		return pio.Rich(method, pio.Green)
	case "HEAD":
		return pio.Rich(method, pio.Blue)
	case "OPTIONS":
		return pio.Rich(method, pio.White)
	default:
		return method
	}
}

func NewConsoleLogger() context.Handler {
	return func(ctx iris.Context) {
		var httpCode, clientIp, method, path string
		var latency time.Duration
		var startTime, endTime time.Time
		startTime = time.Now()

		ctx.Next()

		endTime = time.Now()
		latency = endTime.Sub(startTime)
		httpCode = strconv.Itoa(ctx.GetStatusCode())
		clientIp = ctx.RemoteAddr()
		method = ctx.Method()
		path = ctx.Request().URL.RequestURI()

		fmt.Fprintf(os.Stdout, "%v | %s | %13v | %15s | %s | %s\n",
			endTime.Format("2006/01/02 15:04:05"),
			StatusColor(httpCode),
			latency,
			clientIp,
			MethodColor(method),
			path,
		)
	}
}
