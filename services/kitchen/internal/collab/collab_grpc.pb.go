// Code generated from proto/collab/collab.proto. DO NOT EDIT.
// Regenerate: protoc --go_out=. --go_opt=paths=source_relative \
//               --go-grpc_out=. --go-grpc_opt=paths=source_relative \
//               proto/collab/collab.proto

package collab

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CollabServiceClient is the client API for CollabService service.
type CollabServiceClient interface {
	EditSession(ctx context.Context, opts ...grpc.CallOption) (CollabService_EditSessionClient, error)
}

// CollabServiceServer is the server API for CollabService service.
type CollabServiceServer interface {
	EditSession(CollabService_EditSessionServer) error
	mustEmbedUnimplementedCollabServiceServer()
}

// UnimplementedCollabServiceServer must be embedded to have forward-compatible implementations.
type UnimplementedCollabServiceServer struct{}

func (UnimplementedCollabServiceServer) EditSession(CollabService_EditSessionServer) error {
	return status.Errorf(codes.Unimplemented, "method EditSession not implemented")
}
func (UnimplementedCollabServiceServer) mustEmbedUnimplementedCollabServiceServer() {}

// UnsafeCollabServiceServer may be embedded to opt out of forward compatibility for this service.
type UnsafeCollabServiceServer interface {
	mustEmbedUnimplementedCollabServiceServer()
}

// CollabService_EditSessionClient is the streaming client interface.
type CollabService_EditSessionClient interface {
	Send(*EditRequest) error
	Recv() (*EditResponse, error)
	grpc.ClientStream
}

// CollabService_EditSessionServer is the streaming server interface.
type CollabService_EditSessionServer interface {
	Send(*EditResponse) error
	Recv() (*EditRequest, error)
	grpc.ServerStream
}

// CollabServiceDesc is the grpc.ServiceDesc for CollabService service.
var CollabServiceDesc = grpc.ServiceDesc{
	ServiceName: "collab.CollabService",
	HandlerType: (*CollabServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "EditSession",
			Handler:       editSessionHandler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "proto/collab/collab.proto",
}

// RegisterCollabServiceServer registers the CollabService implementation with the gRPC server.
func RegisterCollabServiceServer(s grpc.ServiceRegistrar, srv CollabServiceServer) {
	s.RegisterService(&CollabServiceDesc, srv)
}

func editSessionHandler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(CollabServiceServer).EditSession(&collabServiceEditSessionServer{stream})
}

type collabServiceEditSessionServer struct {
	grpc.ServerStream
}

func (x *collabServiceEditSessionServer) Send(m *EditResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *collabServiceEditSessionServer) Recv() (*EditRequest, error) {
	m := new(EditRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}
