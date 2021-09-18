package odam

import (
	"fmt"
	"time"

	uuid "github.com/satori/go.uuid"
)

// PublisherService Wrapper around interface
type PublisherService struct {
	ServiceODaMServer
}

func (service *PublisherService) Subscribe(in *EventTypes, stream ServiceODaM_SubscribeServer) error {
	// @todo: implement
	// @todo: until then just test
	for {
		err := stream.Send(
			&EventInformation{
				EventId: uuid.NewV4().String(),
			},
		)
		if err != nil {
			fmt.Println("Error while sending data via stream:", err)
			break
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}
