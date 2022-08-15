package grpcutils

import (
	"context"

	"google.golang.org/grpc"
)

type DirectClientConn struct {
	conn    *grpc.ClientConn
	apiPath string
}

// impl ClientConnInterface
func (cc *DirectClientConn) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	err := cc.conn.Invoke(ctx, cc.apiPath+method, args, reply, opts...)
	// ToDo: maybe https://github.com/kubernetes/ingress-nginx/issues/2963
	if err != nil && err.Error() != errCloseWithoutTrailers {
		return err
	}
	return nil
}

// impl ClientConnInterface
func (cc *DirectClientConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	cs, err := cc.conn.NewStream(ctx, desc, cc.apiPath+method, opts...)
	// ToDo: maybe https://github.com/kubernetes/ingress-nginx/issues/2963
	if err != nil && err.Error() != errCloseWithoutTrailers {
		return nil, err
	}
	return cs, nil
}

func (cc *DirectClientConn) Close() error {
	return cc.conn.Close()
}

// func (cc *DirectClientConn) unaryInterceptor(
// 	ctx context.Context,
// 	method string,
// 	req interface{},
// 	reply interface{},
// 	conn *grpc.ClientConn,
// 	invoker grpc.UnaryInvoker,
// 	opts ...grpc.CallOption,
// ) error {
// 	return invoker(ctx, cc.apiPath+method, req, reply, conn, opts...)
// }

// func (cc *DirectClientConn) streamInterceptor(ctx context.Context, desc *grpc.StreamDesc, conn *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
// 	return streamer(ctx, desc, conn, cc.apiPath+method, opts...)
// }
