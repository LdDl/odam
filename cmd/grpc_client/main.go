package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"time"

	"github.com/LdDl/odam"
	"google.golang.org/grpc"
)

var (
	grpcHost = flag.String("host", "0.0.0.0", "gRPC server address")
	grpcPort = flag.Uint("port", 50052, "gRPC server port")
)

func main() {
	flag.Parse()

	grpcConn, err := grpc.Dial(
		fmt.Sprintf("%s:%d", *grpcHost, *grpcPort),
		grpc.WithInsecure(),
		grpc.WithBlock(),
	)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer grpcConn.Close()

	grpcClient := NewClient(grpcConn)
	err = grpcClient.RecieveAllEvents(context.Background())
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
}

type ODaMClient struct {
	client odam.ServiceODaMClient
}

func NewClient(conn grpc.ClientConnInterface) ODaMClient {
	return ODaMClient{
		client: odam.NewServiceODaMClient(conn),
	}
}

func (client *ODaMClient) RecieveAllEvents(ctx context.Context) error {
	stream, err := client.client.Subscribe(ctx, &odam.EventTypes{
		EventTypes: []odam.EventType{
			odam.EventType_CROSS_LINE_EVENT,
			odam.EventType_ENTER_POLYGON_EVENT,
			odam.EventType_LEAVE_POLYGON_EVENT,
		},
	})
	if err != nil {
		return err
	}

	for {
		message, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				fmt.Println("Everything is OK. Just EOF")
				return nil
			}
			return err
		}
		if message == nil {
			fmt.Println("Recieved message is nil")
			continue
		}
		fmt.Printf("Recieved message is:\n\tEventID: %s\n\tEventTM: %v\n\tEventType:%v\n", message.EventId, time.Unix(message.EventTm, 0), message.EventType)
	}
}
