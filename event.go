package odam

import (
	"fmt"
	"time"

	"github.com/LdDl/gocv-blob/v2/blob"
	uuid "github.com/satori/go.uuid"
)

// EVENT_TYPE Alias to int
type EVENT_TYPE uint32

const (
	// When object has crossed given virtual line
	EVENT_CROSS_LINE = EVENT_TYPE(iota + 1)
	// When object has entered given virtual polygon
	EVENT_ENTER_POLYGON
	// When object has left given virtual polygon
	EVENT_LEAVE_POLYGON
)

// String Implements "fmt" interface
func (e EVENT_TYPE) String() string {
	switch e {
	case EVENT_CROSS_LINE:
		return fmt.Sprintf("EVENT_CROSS_LINE")
	case EVENT_ENTER_POLYGON:
		return fmt.Sprintf("EVENT_ENTER_POLYGON")
	case EVENT_LEAVE_POLYGON:
		return fmt.Sprintf("EVENT_LEAVE_POLYGON")
	default:
		return fmt.Sprintf("Event type '%d' is not supported", e)
	}
}

// Event Information about event
//
// ID - an event's identifier
// EventType - corresponding event type
// Timestamp - corresponding time in UnixTimestamp format (UTC)
// PreviousEvent - optional information about event chain. Could be used for event type EVENT_LEAVE_POLYGON
// SpatialObject - Spatial object (see 'SpatialObjectInfo' description)
//
type Event struct {
	ID            uuid.UUID
	EventType     EVENT_TYPE
	Timestamp     int64
	PreviousEvent *Event
	SpatialObject SpatialObjectInfo
}

// NewEvent Constructor for events
//
// etype - Type of an event
// object - detected object
// prevEvent - optional reference to previous event. Could be used for event type EVENT_LEAVE_POLYGON only
//
func NewEvent(etype EVENT_TYPE, object blob.Blobie, prevEvent ...*Event) (*Event, error) {
	if len(prevEvent) > 1 {
		return nil, fmt.Errorf("Previous event could be only one")
	}
	switch etype {
	case EVENT_CROSS_LINE, EVENT_ENTER_POLYGON:
		return &Event{
			ID:            uuid.NewV4(),
			EventType:     etype,
			Timestamp:     time.Now().UTC().Unix(),
			PreviousEvent: nil,
		}, nil
	case EVENT_LEAVE_POLYGON:
		if len(prevEvent) != 1 {
			return nil, fmt.Errorf("Event is 'EVENT_LEAVE_POLYGON' but previous event hasn't been provided")
		}
		return &Event{
			ID:            uuid.NewV4(),
			EventType:     etype,
			Timestamp:     time.Now().UTC().Unix(),
			PreviousEvent: prevEvent[0],
		}, nil
	default:
		return nil, fmt.Errorf("Event type '%d' is not supported", etype)
	}
}

// String Implements "fmt" interface
func (e Event) String() string {
	switch e.EventType {
	case EVENT_CROSS_LINE, EVENT_ENTER_POLYGON:
		return fmt.Sprintf("Event info:\n\tID: %s\n\tEvent type: %s\n\tEvent time: %v", e.ID.String(), e.EventType.String(), time.Unix(e.Timestamp, 0))
	case EVENT_LEAVE_POLYGON:
		if e.PreviousEvent != nil {
			return fmt.Sprintf("Event info:\n\tID: %s\n\tEvent type: %s\n\tEvent time: %v\n\tPrevious event id: %s", e.ID.String(), e.EventType.String(), time.Unix(e.Timestamp, 0), e.PreviousEvent.ID.String())
		} else {
			return fmt.Sprintf("Event info:\n\tID: %s\n\tEvent type: %s\n\tEvent time: %v\n\tPrevious event id: should be provided", e.ID.String(), e.EventType.String(), time.Unix(e.Timestamp, 0))
		}
	default:
		return fmt.Sprintf("Event type '%d' is not supported", e.EventType)
	}
}
