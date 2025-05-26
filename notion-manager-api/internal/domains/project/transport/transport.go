package transport

type service interface {
}
type ProjectTransport struct {
	service service
}

func NewProjectTransport(service service) *ProjectTransport {
	t := &ProjectTransport{
		service: service,
	}

	return t
}
