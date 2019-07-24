// Code generated by protoc-gen-grpc-gateway. DO NOT EDIT.
// source: service/nodeRegistration.proto

/*
Package service is a reverse proxy.

It translates gRPC into RESTful JSON APIs.
*/
package service

import (
	"context"
	"io"
	"net/http"

	"github.com/golang/protobuf/proto"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/grpc-ecosystem/grpc-gateway/utilities"
	"github.com/zoobc/zoobc-core/common/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/status"
)

var _ codes.Code
var _ io.Reader
var _ status.Status
var _ = runtime.String
var _ = utilities.NewDoubleArray

var (
	filter_NodeRegistrationService_GetNodeRegistrations_0 = &utilities.DoubleArray{Encoding: map[string]int{}, Base: []int(nil), Check: []int(nil)}
)

func request_NodeRegistrationService_GetNodeRegistrations_0(ctx context.Context, marshaler runtime.Marshaler, client NodeRegistrationServiceClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq model.GetNodeRegistrationsRequest
	var metadata runtime.ServerMetadata

	if err := runtime.PopulateQueryParameters(&protoReq, req.URL.Query(), filter_NodeRegistrationService_GetNodeRegistrations_0); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := client.GetNodeRegistrations(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

var (
	filter_NodeRegistrationService_GetNodeRegistration_0 = &utilities.DoubleArray{Encoding: map[string]int{}, Base: []int(nil), Check: []int(nil)}
)

func request_NodeRegistrationService_GetNodeRegistration_0(ctx context.Context, marshaler runtime.Marshaler, client NodeRegistrationServiceClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq model.GetNodeRegistrationRequest
	var metadata runtime.ServerMetadata

	if err := runtime.PopulateQueryParameters(&protoReq, req.URL.Query(), filter_NodeRegistrationService_GetNodeRegistration_0); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := client.GetNodeRegistration(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

// RegisterNodeRegistrationServiceHandlerFromEndpoint is same as RegisterNodeRegistrationServiceHandler but
// automatically dials to "endpoint" and closes the connection when "ctx" gets done.
func RegisterNodeRegistrationServiceHandlerFromEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error) {
	conn, err := grpc.Dial(endpoint, opts...)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if cerr := conn.Close(); cerr != nil {
				grpclog.Infof("Failed to close conn to %s: %v", endpoint, cerr)
			}
			return
		}
		go func() {
			<-ctx.Done()
			if cerr := conn.Close(); cerr != nil {
				grpclog.Infof("Failed to close conn to %s: %v", endpoint, cerr)
			}
		}()
	}()

	return RegisterNodeRegistrationServiceHandler(ctx, mux, conn)
}

// RegisterNodeRegistrationServiceHandler registers the http handlers for service NodeRegistrationService to "mux".
// The handlers forward requests to the grpc endpoint over "conn".
func RegisterNodeRegistrationServiceHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return RegisterNodeRegistrationServiceHandlerClient(ctx, mux, NewNodeRegistrationServiceClient(conn))
}

// RegisterNodeRegistrationServiceHandlerClient registers the http handlers for service NodeRegistrationService
// to "mux". The handlers forward requests to the grpc endpoint over the given implementation of "NodeRegistrationServiceClient".
// Note: the gRPC framework executes interceptors within the gRPC handler. If the passed in "NodeRegistrationServiceClient"
// doesn't go through the normal gRPC flow (creating a gRPC client etc.) then it will be up to the passed in
// "NodeRegistrationServiceClient" to call the correct interceptors.
func RegisterNodeRegistrationServiceHandlerClient(ctx context.Context, mux *runtime.ServeMux, client NodeRegistrationServiceClient) error {

	mux.Handle("GET", pattern_NodeRegistrationService_GetNodeRegistrations_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_NodeRegistrationService_GetNodeRegistrations_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_NodeRegistrationService_GetNodeRegistrations_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("GET", pattern_NodeRegistrationService_GetNodeRegistration_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_NodeRegistrationService_GetNodeRegistration_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_NodeRegistrationService_GetNodeRegistration_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	return nil
}

var (
	pattern_NodeRegistrationService_GetNodeRegistrations_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2}, []string{"v1", "nodeRegistration", "GetNodeRegistrations"}, ""))

	pattern_NodeRegistrationService_GetNodeRegistration_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2}, []string{"v1", "nodeRegistration", "GetNodeRegistration"}, ""))
)

var (
	forward_NodeRegistrationService_GetNodeRegistrations_0 = runtime.ForwardResponseMessage

	forward_NodeRegistrationService_GetNodeRegistration_0 = runtime.ForwardResponseMessage
)
