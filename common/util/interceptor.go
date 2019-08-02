package util

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"

	"google.golang.org/grpc"
)

/*
NewInterceptor function can use to inject middlewares like:
	- `recover`
	- `log` triggered
	- validate `authentication` if needed
With `recover()` function can handle re-run the app while got panic or error
*/
func NewServerInterceptor(logger *logrus.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		start := time.Now()
		fields := logrus.Fields{
			"method": info.FullMethod,
			"time":   start.String(),
		}
		defer func() {

			fields["latency"] = fmt.Sprintf("%d ns", time.Since(start).Nanoseconds())
			if err := recover(); err != nil {
				// get stack after panic called and perhaps its first error
				_, file, line, _ := runtime.Caller(4)
				fields["error"] = fmt.Sprintf("%s %d", file, line)
				if logger != nil {
					logger.WithFields(fields).Error(fmt.Sprint(err))
				}
			} else if logger != nil {
				logger.WithFields(fields).Info("success")
			}
		}()

		resp, err := handler(ctx, req)
		if err != nil {
			fields["exception"] = err
		}
		return resp, err
	}
}
