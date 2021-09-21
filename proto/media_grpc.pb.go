// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.1.0
// - protoc             v3.18.0
// source: proto/media.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// MediaServiceClient is the client API for MediaService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MediaServiceClient interface {
	DownloadImage(ctx context.Context, in *ImageInfo, opts ...grpc.CallOption) (*FileUploadInfo, error)
}

type mediaServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewMediaServiceClient(cc grpc.ClientConnInterface) MediaServiceClient {
	return &mediaServiceClient{cc}
}

func (c *mediaServiceClient) DownloadImage(ctx context.Context, in *ImageInfo, opts ...grpc.CallOption) (*FileUploadInfo, error) {
	out := new(FileUploadInfo)
	err := c.cc.Invoke(ctx, "/proto.MediaService/DownloadImage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MediaServiceServer is the server API for MediaService service.
// All implementations must embed UnimplementedMediaServiceServer
// for forward compatibility
type MediaServiceServer interface {
	DownloadImage(context.Context, *ImageInfo) (*FileUploadInfo, error)
	mustEmbedUnimplementedMediaServiceServer()
}

// UnimplementedMediaServiceServer must be embedded to have forward compatible implementations.
type UnimplementedMediaServiceServer struct {
}

func (UnimplementedMediaServiceServer) DownloadImage(context.Context, *ImageInfo) (*FileUploadInfo, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DownloadImage not implemented")
}
func (UnimplementedMediaServiceServer) mustEmbedUnimplementedMediaServiceServer() {}

// UnsafeMediaServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MediaServiceServer will
// result in compilation errors.
type UnsafeMediaServiceServer interface {
	mustEmbedUnimplementedMediaServiceServer()
}

func RegisterMediaServiceServer(s grpc.ServiceRegistrar, srv MediaServiceServer) {
	s.RegisterService(&MediaService_ServiceDesc, srv)
}

func _MediaService_DownloadImage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ImageInfo)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MediaServiceServer).DownloadImage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.MediaService/DownloadImage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MediaServiceServer).DownloadImage(ctx, req.(*ImageInfo))
	}
	return interceptor(ctx, in, info, handler)
}

// MediaService_ServiceDesc is the grpc.ServiceDesc for MediaService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var MediaService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.MediaService",
	HandlerType: (*MediaServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "DownloadImage",
			Handler:    _MediaService_DownloadImage_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/media.proto",
}
