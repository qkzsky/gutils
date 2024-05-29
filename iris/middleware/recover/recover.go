// Package recover provides recovery for specific routes or for the whole app via middleware. See _examples/miscellaneous/recover
package recover

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"net/http/httputil"
	"os"
	"runtime"
	"runtime/debug"
	//"github.com/kataras/iris/v12/middleware/requestid"
	"github.com/kataras/pio"
	"github.com/qkzsky/gutils/logger"
	"go.uber.org/zap"
)

func init() {
	context.SetHandlerName("go-utils/iris/middleware/recover.*", "go-utils.recover")
}

func getRequestLogs(ctx *context.Context) string {
	rawReq, _ := httputil.DumpRequest(ctx.Request(), false)
	return string(rawReq)
}

// New returns a new recover middleware,
// it recovers from panics and logs
// the panic message to the application's logger "Warn" level.
func New() context.Handler {
	skip := 1
	recoverLogger := logger.GetDefaultLogger().WithOptions(zap.AddCaller(), zap.AddCallerSkip(skip))

	return func(ctx iris.Context) {
		defer func() {
			if err := recover(); err != nil {
				if ctx.IsStopped() {
					return
				}

				var clientIp, method, uri, userAgent, stacktrace string
				httpCode := ctx.GetStatusCode()
				uri = ctx.Request().URL.RequestURI()
				method = ctx.Method()
				clientIp = ctx.RemoteAddr()
				userAgent = ctx.Request().UserAgent()

				var callers []string
				for i := 1 + skip; ; i++ {
					_, file, line, got := runtime.Caller(i)
					if !got {
						break
					}
					stacktrace += fmt.Sprintf("%s:%d\n", file, line)
					callers = append(callers, fmt.Sprintf("%s:%d", file, line))
				}

				// when stack finishes
				logMessage := fmt.Sprintf("Recovered from a route's Handler('%s')\n", ctx.HandlerName())
				logMessage += fmt.Sprint(getRequestLogs(ctx))
				logMessage += fmt.Sprintf("%s\n", err)
				logMessage += fmt.Sprintf("%s", stacktrace)

				fmt.Fprintf(os.Stderr, pio.Rich(logMessage, pio.Red)+"\n")
				recoverLogger.Error(fmt.Sprintf("Recovered from a route's Handler('%s')", ctx.HandlerName()),
					//zap.String("request_id", requestid.Get(ctx)),
					zap.String("uri", uri),
					zap.String("user_agent", userAgent),
					zap.String("client_ip", clientIp),
					zap.Int("http_code", httpCode),
					zap.String("method", method),
					zap.String("error", fmt.Sprintf("%v", err)),
					zap.String("stacktrace", stacktrace),
				)

				// get the list of registered handlers and the
				// handler which panic derived from.
				handlers := ctx.Handlers()
				handlersFileLines := make([]string, 0, len(handlers))
				currentHandlerIndex := ctx.HandlerIndex(-1)
				currentHandlerFileLine := "???"
				for i, h := range ctx.Handlers() {
					file, line := context.HandlerFileLine(h)
					fileLine := fmt.Sprintf("%s:%d", file, line)
					handlersFileLines = append(handlersFileLines, fileLine)
					if i == currentHandlerIndex {
						currentHandlerFileLine = fileLine
					}
				}

				ctx.StopWithPlainError(iris.StatusInternalServerError, &context.ErrPanicRecovery{
					Cause:                  err,
					Callers:                callers,
					Stack:                  debug.Stack(),
					RegisteredHandlers:     handlersFileLines,
					CurrentHandlerFileLine: currentHandlerFileLine,
				})
			}
		}()

		ctx.Next()
	}
}
