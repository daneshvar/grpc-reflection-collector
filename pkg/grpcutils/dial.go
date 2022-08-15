package grpcutils

import (
	"context"

	"google.golang.org/grpc"
)

func DirectDial(targetUrl string, opts ...grpc.DialOption) (conn *DirectClientConn, err error) {
	return DirectDialContext(context.Background(), targetUrl, opts...)
}

func DirectDialContext(ctx context.Context, targetUrl string, opts ...grpc.DialOption) (conn *DirectClientConn, err error) {
	host, path, grpcs, err := parseURL(targetUrl)
	if err != nil {
		return nil, err
	}

	conn = &DirectClientConn{
		apiPath: path,
	}

	opts = append(opts, withTransportCredentials(grpcs))
	conn.conn, err = grpc.DialContext(ctx, host, opts...) // grpc.WithUnaryInterceptor(conn.unaryInterceptor), grpc.WithStreamInterceptor(conn.streamInterceptor)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
