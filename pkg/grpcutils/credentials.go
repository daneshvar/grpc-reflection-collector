package grpcutils

import (
	"crypto/tls"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func withTransportCredentials(secure bool) grpc.DialOption {
	return grpc.WithTransportCredentials(transportCredentials(secure))
}

func transportCredentials(secure bool) credentials.TransportCredentials {
	if secure {
		return credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})
	} else {
		return insecure.NewCredentials()
	}
}
