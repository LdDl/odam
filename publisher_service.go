package odam

import (
	"context"
)

// PublisherService Wrapper around interface
type PublisherService struct {
	ServiceODaMServer
}

func (service *PublisherService) SendGPSBulk(ctx context.Context, in *EventTypes) (*EventInformation, error) {
	return &EventInformation{}, nil
}
