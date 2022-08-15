package reflection

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"

	"github.com/daneshvar/go-logger"
	"github.com/daneshvar/grpc-reflection-collector/pkg/grpcutils"
	rpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

type ServerReflectionServer struct {
	log           *logger.Logger
	services      map[string]rpb.ServerReflectionClient
	serviceIgnore map[string]struct{}
	cache         sync.Map
}

const (
	enableCache = false
	verbose     = false
)

func NewServerReflectionServer(ctx context.Context, log *logger.Logger, services map[string]string, ignores []string) (*ServerReflectionServer, error) {
	serviceIgnore := make(map[string]struct{})

	serviceIgnore["grpc.reflection.v1alpha.ServerReflection"] = struct{}{}
	for _, s := range ignores {
		serviceIgnore[s] = struct{}{}
	}

	s := &ServerReflectionServer{
		log:           log,
		services:      make(map[string]rpb.ServerReflectionClient),
		serviceIgnore: serviceIgnore,
	}

	for name, addr := range services {
		client, err := newServerReflectionClient(ctx, addr)
		if err != nil {
			return nil, err
		}
		s.services[name] = client
	}

	return s, nil
}

func newServerReflectionClient(ctx context.Context, addr string) (rpb.ServerReflectionClient, error) {
	conn, err := grpcutils.DirectDialContext(ctx, addr)
	if err != nil {
		return nil, err
	}

	client := rpb.NewServerReflectionClient(conn)
	return client, nil
}

func (sr *ServerReflectionServer) ServerReflectionInfo(stream rpb.ServerReflection_ServerReflectionInfoServer) error {
	ctx := stream.Context()

	// log.Println("Start")

	for {
		// exit if context is done
		// or continue
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// receive data from stream
		req, err := stream.Recv()
		if err == io.EOF {
			// return will close stream from server side
			// log.Println("EOF")
			break
		}
		if err != nil {
			sr.log.Errorf("receive error %v", err)
			continue
		}

		var r *rpb.ServerReflectionResponse = nil
		servicesAdded := make(map[string]struct{})

		if enableCache {
			if resp, ok := sr.cache.Load(fmt.Sprint(req)); ok {
				r = resp.(*rpb.ServerReflectionResponse)
			}
		}

		if r != nil {
			if verbose {
				sr.log.Infof("Request: req: %v cache", req)
			}
		} else {
			if verbose {
				sr.log.Infof("Request: req: %v", req)
			}

			matchSymbol := false
			for service := range sr.services {
				symbole := req.GetFileContainingSymbol()
				if symbole != "" && strings.HasPrefix(symbole, service+".") {
					matchSymbol = true
					break
				}
			}

			for service, client := range sr.services {
				if matchSymbol {
					symbole := req.GetFileContainingSymbol()
					if symbole != "" && !strings.HasPrefix(symbole, service+".") {
						continue
					}
				}

				// if req.GetFileByFilename() == "auth/common/common_register.proto" {
				// 	sr.log.Infof("getReflectionInfo: %s: req: %v", service, req)
				// }

				resp, err := sr.getReflectionInfo(ctx, service, client, req)
				if err != nil {
					sr.log.Errorf("Error on %s: req: %v error: %v", service, req, err)
					continue
				}

				// if strings.Contains(fmt.Sprint(resp), "common_register.proto") {
				// 	sr.log.Infof("Response on %s req: %v error: %v", s, req, resp)
				// }

				if resp.GetErrorResponse() != nil {
					if resp.GetErrorResponse().ErrorCode != 5 { // 5 is rpc.NOT_FOUND
						sr.log.Errorf("Response Error on %s req: %v error: %v", service, req, resp.GetErrorResponse())
					}
					continue
				}

				if req.GetListServices() != "*" {
					r = resp
					break
				}

				// responsed to req.GetListServices() is *
				if r == nil {
					r = &rpb.ServerReflectionResponse{
						MessageResponse: &rpb.ServerReflectionResponse_ListServicesResponse{
							ListServicesResponse: &rpb.ListServiceResponse{
								Service: make([]*rpb.ServiceResponse, 0),
							},
						},
					}
				}

				list := r.MessageResponse.(*rpb.ServerReflectionResponse_ListServicesResponse)
				dst := list.ListServicesResponse
				srcServices := resp.GetListServicesResponse().Service
				for i := range srcServices {
					serviceName := srcServices[i].GetName()
					if _, ok := sr.serviceIgnore[serviceName]; !ok {
						if _, ok := servicesAdded[serviceName]; !ok {
							servicesAdded[serviceName] = struct{}{}
							dst.Service = append(dst.Service, srcServices[i])
						}
					}
				}
			}

			if req.GetListServices() == "*" {
				services := r.MessageResponse.(*rpb.ServerReflectionResponse_ListServicesResponse).ListServicesResponse.Service
				sort.Slice(services, func(i, j int) bool {
					return services[i].GetName() < services[j].GetName()
				})
				sr.log.Tracef("Send MessageResponse: %v\n", r.GetListServicesResponse().Service)
			}

			if enableCache {
				sr.cache.Store(fmt.Sprint(req), r)
			}
		}

		if r != nil {
			if err := stream.Send(r); err != nil {
				sr.log.Errorf("Error On send: %v", err)
			}
			r = nil
		} else {
			sr.log.Warnf("Response is nil: %v", req)
		}
	}

	// log.Println("Finish")

	return nil
}

func (sr *ServerReflectionServer) getReflectionInfo(ctx context.Context, service string, client rpb.ServerReflectionClient, req *rpb.ServerReflectionRequest) (*rpb.ServerReflectionResponse, error) {
	// ctx, cancel := context.WithTimeout(ctx, time.Minute)
	// defer cancel()

	stream, err := client.ServerReflectionInfo(ctx)
	if err != nil {
		return nil, err
	}

	if err := stream.Send(req); err != nil {
		return nil, err
	}

	resp, err := stream.Recv()
	if err == io.EOF {
		return nil, err
	}
	if err != nil {
		sr.log.Errorf("%s can not receive %v", service, err)
		return nil, err
	}

	return resp, nil
}
