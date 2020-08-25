package interceptor

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/zoobc/zoobc-core/common/blocker"
	"github.com/zoobc/zoobc-core/common/crypto"
	"github.com/zoobc/zoobc-core/common/model"
	"github.com/zoobc/zoobc-core/common/monitoring"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type simpleRateLimiter struct {
	numberOfAllowedRequest uint32
	numberOfRequest        uint32
	sync.Mutex
}

func (rl *simpleRateLimiter) isAllowed() bool {
	if rl.numberOfAllowedRequest == 0 {
		return true
	}
	rl.Lock()
	defer rl.Unlock()
	if rl.numberOfRequest >= rl.numberOfAllowedRequest {
		return false
	}
	rl.numberOfRequest++
	return true
}

func (rl *simpleRateLimiter) requestFinished() {
	rl.Lock()
	defer rl.Unlock()
	if rl.numberOfRequest > 0 {
		rl.numberOfRequest--
	}
}

func (rl *simpleRateLimiter) start() {
	if rl.numberOfAllowedRequest == 0 {
		return
	}
	ticker := time.NewTicker(time.Second)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-ticker.C:
			func() {
				rl.Lock()
				defer rl.Unlock()
				rl.numberOfRequest = 0
			}()
		case <-sigs:
			ticker.Stop()
			return
		}
	}
}

/*
NewServerRateLimiterInterceptor function can used to add rate limit to the server call
*/

func NewServerRateLimiterInterceptor(requestLimitPerSecond uint32) grpc.UnaryServerInterceptor {
	rateLimiter := &simpleRateLimiter{
		numberOfAllowedRequest: requestLimitPerSecond,
	}
	go rateLimiter.start()
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if !rateLimiter.isAllowed() {
			return nil, status.Error(codes.ResourceExhausted, "requests are limited")
		}
		defer rateLimiter.requestFinished()
		return handler(ctx, req)
	}
}

/*
NewServerInterceptor function can use to inject middleware like:
	- `recover`
	- `log` triggered
	- validate `authentication` if needed
With `recover()` function can handle re-run the app while got panic or error
*/
func NewServerInterceptor(
	logger *logrus.Logger,
	ownerAddress string,
	ignoredErrCodes map[codes.Code]string,
) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		monitoring.IncrementRunningAPIHandling(info.FullMethod)
		var (
			authorizedErr, errHandler error
			resp                      interface{}
			start                     = time.Now()
			fields                    logrus.Fields
		)

		fields = logrus.Fields{
			"method": info.FullMethod,
			"time":   start.String(),
			"source": "serverHandler",
		}

		defer func() {
			var (
				err     = recover()
				latency = time.Since(start)
			)

			fields["latency"] = fmt.Sprintf("%d ns", latency.Nanoseconds())
			monitoring.SetAPIResponseTime(info.FullMethod, latency.Seconds())
			monitoring.DecrementRunningAPIHandling(info.FullMethod)
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
					if _, ok := ignoredErrCodes[status.Code(errHandler)]; !ok {
						logger.WithFields(fields).Error(fmt.Sprint(errHandler))
					}
				default:
					logger.WithFields(fields).Info("success")
				}
			}
		}()

		// authorize request
		authorizedErr = authRequest(ctx, info.FullMethod, ownerAddress)
		if authorizedErr != nil {
			return nil, authorizedErr
		}

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
func NewClientInterceptor(logger *logrus.Logger, ignoredErrors map[codes.Code]string) grpc.UnaryClientInterceptor {
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
			start      = time.Now()
			fields     = logrus.Fields{
				"method": method,
				"time":   start.String(),
				"source": "clientHandler",
			}
		)

		defer func() {
			var (
				err = recover()
			)

			fields["latency"] = fmt.Sprintf("%d ns", time.Since(start).Nanoseconds())
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
					if _, ok := ignoredErrors[status.Code(errInvoker)]; !ok {
						logger.WithFields(fields).Error(fmt.Sprint(errInvoker))
					}
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
NewStreamInterceptor validate request against the destination service and the signature with the node owner
*/
func NewStreamInterceptor(ownerAddress string) grpc.StreamServerInterceptor {
	return func(
		srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler,
	) error {
		var (
			errHandler error
		)
		err := authRequest(ss.Context(), info.FullMethod, ownerAddress)
		if err != nil {
			return err
		}

		errHandler = handler(srv, ss)
		return errHandler
	}
}

// authRequest shared logic to authorize an off-chain api requests
func authRequest(ctx context.Context, method, ownerAddress string) error {
	var (
		requestType model.RequestType
	)
	switch method {
	case "/service.NodeHardwareService/GetNodeHardware":
		requestType = model.RequestType_GetNodeHardware
	case "/service.NodeAdminService/GetProofOfOwnership":
		requestType = model.RequestType_GetProofOfOwnership
	case "/service.NodeAdminService/GenerateNodeKey":
		requestType = model.RequestType_GeneratetNodeKey
	case "/service.NodeAdminService/GetPendingNodeRegistrations":
		requestType = model.RequestType_GetPendingNodeRegistrationsStream
	default:
		// unprotected service, by pass the auth checking
		requestType = -1
	}

	if requestType > -1 {
		md, ok := metadata.FromIncomingContext(ctx)
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
		// validate request
		// todo: this is verifying against whatever owner address is in the config file, update this
		// todo: to follow how `claim` node work.
		err := crypto.VerifyAuthAPI(
			ownerAddress,
			authSlice[0],
			requestType,
		)
		if err != nil {
			return err
		}
	}
	return nil
}
