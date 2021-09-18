package odam

import (
	"fmt"
)

// PublisherService Wrapper around interface
type PublisherService struct {
	ServiceODaMServer
	eventsChannel chan *Event
	clientNum     int
}

func (service *PublisherService) Subscribe(in *EventTypes, stream ServiceODaM_SubscribeServer) error {
	for {
		event := <-service.eventsChannel
		err := stream.Send(
			&EventInformation{
				EventId:   event.ID.String(),
				EventType: EventType(event.EventType),
				EventTm:   event.Timestamp,
			},
		)
		if err != nil {
			fmt.Println("Error while sending data via stream:", err)
			break
		}
	}
	return nil
}
