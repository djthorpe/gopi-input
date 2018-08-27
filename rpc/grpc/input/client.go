/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2016-2018
	All Rights Reserved
	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package input

import (
	"context"
	"fmt"
	"io"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	grpc "github.com/djthorpe/gopi-rpc/sys/grpc"

	// Protocol buffers
	pb "github.com/djthorpe/gopi-input/rpc/protobuf/input"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Client struct {
	pb.InputClient
	conn gopi.RPCClientConn
}

////////////////////////////////////////////////////////////////////////////////
// NEW

func NewInputClient(conn gopi.RPCClientConn) gopi.RPCClient {
	return &Client{pb.NewInputClient(conn.(grpc.GRPCClientConn).GRPCConn()), conn}
}

func (this *Client) NewContext() context.Context {
	if this.conn.Timeout() == 0 {
		return context.Background()
	} else {
		ctx, _ := context.WithTimeout(context.Background(), this.conn.Timeout())
		return ctx
	}
}

////////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (this *Client) Conn() gopi.RPCClientConn {
	return this.conn
}

////////////////////////////////////////////////////////////////////////////////
// CALLS

func (this *Client) Ping() error {
	this.conn.Lock()
	defer this.conn.Unlock()

	// Perform Ping
	if _, err := this.InputClient.Ping(this.NewContext(), &pb.EmptyRequest{}); err != nil {
		return err
	} else {
		return nil
	}
}

func (this *Client) ListenForInputEvents(done <-chan struct{}, events chan<- gopi.InputEvent) error {
	this.conn.Lock()
	defer this.conn.Unlock()

	// Create a context with a cancel function, and wait for the 'done'
	// in background
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-done
		cancel()
	}()

	// Receive a stream of messages, when done is received then
	// context.Cancel() is called to end the loop, which returns nil
	if stream, err := this.InputClient.ListenForInputEvents(ctx, &pb.EmptyRequest{}); err != nil {
		return err
	} else {
		for {
			if input_event_, err := stream.Recv(); err == io.EOF {
				break
			} else if err != nil {
				return err
			} else if input_event := fromProtobufInputEvent(nil, input_event_); input_event.EventType() != gopi.INPUT_EVENT_NONE {
				// We don't emit INPUT_EVENT_NONE events (null events) which are continuously sent to ensure the
				// connection stays alive
				events <- input_event
			}
		}
	}

	// Success
	return nil
}

func (this *Client) Devices() error {
	this.conn.Lock()
	defer this.conn.Unlock()

	if devices, err := this.InputClient.Devices(this.NewContext(), &pb.EmptyRequest{}); err != nil {
		return err
	} else {
		fmt.Println(devices)
		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *Client) String() string {
	return fmt.Sprintf("<gopi.InputClient>{ conn=%v }", this.conn)
}
