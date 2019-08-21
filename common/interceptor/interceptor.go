package interceptor

import (
	"context"
	"fmt"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"google.golang.org/grpc/metadata"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"

	"google.golang.org/grpc"
)

/*
NewServerInterceptor function can use to inject middlewares like:
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
		var (
			errHandler error
			resp       interface{}
		)
		start := time.Now()
		fields := logrus.Fields{
			"method": info.FullMethod,
			"time":   start.String(),
		}

		defer func() {

			fields["latency"] = fmt.Sprintf("%d ns", time.Since(start).Nanoseconds())
			err := recover()
			if err != nil {
				// get stack after panic called and perhaps its first error
				_, file, line, _ := runtime.Caller(4)
				fields["panic"] = fmt.Sprintf("%s %d", file, line)
			} else if errHandler != nil {
				fields["error"] = errHandler
			}

			if logger != nil {
				switch {
				case err != nil:
					logger.WithFields(fields).Error(fmt.Sprint(err))
				case errHandler != nil:
					logger.WithFields(fields).Warning(errHandler)
				default:
					logger.WithFields(fields).Info("success")
				}
			}
		}()

		resp, errHandler = handler(ctx, req)
		return resp, errHandler
	}
}

/*
NewClientInterceptor function can use to inject using grpc client like:
	- `recover`
	- `log` triggered
	- add access token if needed
With `recover()` function can handle re-run the app while got panic
*/
func NewClientInterceptor(logger *logrus.Logger) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req interface{},
		reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		var (
			errInvoker error
		)
		start := time.Now()
		fields := logrus.Fields{
			"method": method,
			"time":   start.String(),
		}

		defer func() {
			fields["latency"] = fmt.Sprintf("%d ns", time.Since(start).Nanoseconds())
			err := recover()
			if err != nil {
				// get stack after panic called and perhaps its first error
				_, file, line, _ := runtime.Caller(4)
				fields["panic"] = fmt.Sprintf("%s %d", file, line)
			}
			if logger != nil {
				switch {
				case err != nil:
					logger.WithFields(fields).Error(fmt.Sprint(err))
				case errInvoker != nil:
					logger.WithFields(fields).Warning(fmt.Sprint(errInvoker))
				default:
					logger.WithFields(fields).Info("success")
				}
			}
		}()
		errInvoker = invoker(ctx, method, req, reply, cc, opts...)
		return errInvoker
	}

}

/*
NewStreamInterceptor
validate request against the destination service and the signature with the node owner
*/
func NewStreamInterceptor(ownerAddress string) grpc.StreamServerInterceptor {
	return func(
		srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler,
	) error {
		var (
			errHandler  error
			requestType model.RequestType
		)
		switch info.FullMethod {
		case "/service.NodeHardwareService/GetNodeHardware":
			requestType = model.RequestType_GetNodeHardware
		default:
			// unprotected service, by pass the auth checking
			requestType = -1
		}
		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			return blocker.NewBlocker(
				blocker.AuthErr,
				"metadata not provided",
			)
		}
		authSlice := md.Get("authorization")
		if authSlice == nil {
			return blocker.NewBlocker(
				blocker.AuthErr,
				"authorization metadata not provided",
			)
		}
		if requestType > -1 {
			// validate request
			// todo: this is verifying against whatever owner address is in the config file, update this
			// todo: to follow how `claim` node work.
			err := crypto.VerifyAuthAPI(
				ownerAddress,
				authSlice[0],
				requestType)
			if err != nil {
				return err
			}
		}

		errHandler = handler(srv, ss)
		return errHandler
	}
}
