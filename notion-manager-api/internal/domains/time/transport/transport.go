package transport

import (
	"google.golang.org/grpc"
)

type service interface {
}
type TimeTransport struct {
	service service
}

func NewTimeTransport(grpcServer *grpc.Server, service service) *TimeTransport {
	t := &TimeTransport{
		service: service,
	}

	return t
}
