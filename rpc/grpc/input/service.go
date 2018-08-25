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

	// Frameworks
	"github.com/djthorpe/gopi"

	// Modules
	"github.com/djthorpe/gopi/sys/rpc/grpc"

	// Protocol buffers
	pb "github.com/djthorpe/gopi-input/rpc/protobuf/input"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Service struct {
	Server       gopi.RPCServer
	InputManager gopi.InputManager
	DeviceName   string
	DeviceType   gopi.InputDeviceType
	DeviceBus    gopi.InputDeviceBus
}

type service struct {
	log   gopi.Logger
	input gopi.InputManager

	done chan struct{}
}

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

// Open the server
func (config Service) Open(log gopi.Logger) (gopi.Driver, error) {
	log.Debug("<grpc.service.input>Open{ server=%v input=%v device_name='%v' device_type=%v device_bus=%v  }", config.Server, config.InputManager, config.DeviceName, config.DeviceType, config.DeviceBus)

	// Check for bad input parameters
	if config.Server == nil || config.InputManager == nil {
		return nil, gopi.ErrBadParameter
	}

	this := new(service)
	this.log = log
	this.input = config.InputManager
	this.done = make(chan struct{})

	// Register service with GRPC server
	pb.RegisterInputServer(config.Server.(grpc.GRPCServer).GRPCServer(), this)

	// Get events in the background
	go this.receiveEvents()

	// Subscribe to input devices
	if devices, err := this.input.OpenDevicesByName(config.DeviceName, config.DeviceType, config.DeviceBus); err != nil {
		return nil, err
	} else {
		this.log.Debug("Number of devices opened: %v", len(devices))
	}

	// Success
	return this, nil
}

func (this *service) Close() error {
	this.log.Debug("<grpc.service.input>Close{}")

	// Signal done and wait for close
	this.done <- gopi.DONE
	<-this.done

	// Release resources
	this.input = nil

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// Stringify

func (this *service) String() string {
	return fmt.Sprintf("grpc.service.input{}")
}

////////////////////////////////////////////////////////////////////////////////
// Receive events in the background

func (this *service) receiveEvents() {
	evt_device := this.input.Subscribe()

FOR_LOOP:
	for {
		select {
		case evt := <-evt_device:
			this.log.Info("evt=%v", evt)
		case <-this.done:
			this.log.Info("done")
			break FOR_LOOP
		}
	}
	this.input.Unsubscribe(evt_device)
	close(this.done)
}

////////////////////////////////////////////////////////////////////////////////
// Ping method

func (this *service) Ping(ctx context.Context, _ *pb.EmptyRequest) (*pb.EmptyReply, error) {
	return &pb.EmptyReply{}, nil
}

////////////////////////////////////////////////////////////////////////////////
// Listen for inputmanager events

func (this *service) ListenForInputEvents(_ *pb.EmptyRequest, stream pb.Input_ListenForInputEventsServer) error {
	this.log.Debug("ListenForInputEvents: Subscribe")
	events := this.input.Subscribe()

FOR_LOOP:
	// Send until loop is broken
	for {
		select {
		case evt := <-events:
			if evt == nil {
				this.log.Warn("ListenForInputEvents: channel closed: closing request")
				break FOR_LOOP
			} else if input_evt, ok := evt.(gopi.InputEvent); ok == false {
				this.log.Warn("ListenForInputEvents: ignoring event: %v", evt)
			} else if err := stream.Send(toProtobufInputEvent(input_evt)); err != nil {
				this.log.Warn("ListenForInputEvents: error sending: %v: closing request", err)
				break FOR_LOOP
			}
		}
	}

	// Unsubscribe from events
	this.log.Debug("ListenForInputEvents: Unsubscribe")
	this.input.Unsubscribe(events)

	// Return success
	return nil
}
