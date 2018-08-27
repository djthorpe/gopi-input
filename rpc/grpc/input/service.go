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
	"github.com/djthorpe/gopi-rpc/sys/grpc"
	"github.com/djthorpe/gopi/util/event"

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

	// Implements publisher for when the service is stopped
	event.Publisher
}

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

// Open the server
func (config Service) Open(log gopi.Logger) (gopi.Driver, error) {
	log.Debug("<grpc.service.input.Open>{ server=%v input=%v device_name='%v' device_type=%v device_bus=%v  }", config.Server, config.InputManager, config.DeviceName, config.DeviceType, config.DeviceBus)

	// Check for bad input parameters
	if config.Server == nil || config.InputManager == nil {
		return nil, gopi.ErrBadParameter
	}

	this := new(service)
	this.log = log
	this.input = config.InputManager

	// Register service with GRPC server
	pb.RegisterInputServer(config.Server.(grpc.GRPCServer).GRPCServer(), this)

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
	this.log.Debug("<grpc.service.input.Close>{}")

	// Close publisher
	this.Publisher.Close()

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
// Cancel all streaming requests

func (this *service) CancelRequests() error {
	this.log.Debug2("<grpc.service.input.CancelRequests>{}")

	// Cancel any streaming requests
	this.Publisher.Emit(event.NullEvent)

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// Ping method

func (this *service) Ping(ctx context.Context, _ *pb.EmptyRequest) (*pb.EmptyReply, error) {
	this.log.Debug2("<grpc.service.input.Ping>{ }")
	return &pb.EmptyReply{}, nil
}

////////////////////////////////////////////////////////////////////////////////
// Listen for InputManager events

func (this *service) ListenForInputEvents(_ *pb.EmptyRequest, stream pb.Input_ListenForInputEventsServer) error {
	this.log.Debug2("<grpc.service.input.ListenForInputEvents>{ }")

	// Subscribe to events
	events := this.input.Subscribe()
	cancel_requests := this.Publisher.Subscribe()

FOR_LOOP:
	// Send until loop is broken - either due to stream error
	// or request cancellation
	for {
		select {
		case evt := <-events:
			if evt == nil {
				this.log.Warn("<grpc.service.input.ListenForInputEvents> Error: channel closed: closing request")
				break FOR_LOOP
			} else if input_evt, ok := evt.(gopi.InputEvent); ok == false {
				this.log.Warn("<grpc.service.input.ListenForInputEvents> Warning: ignoring event: %v", evt)
			} else if err := stream.Send(toProtobufInputEvent(input_evt)); err != nil {
				this.log.Warn("<grpc.service.input.ListenForInputEvents> Warning: %v: closing request", err)
				break FOR_LOOP
			}
		case <-cancel_requests:
			break FOR_LOOP
		}
	}

	// Unsubscribe from events
	this.input.Unsubscribe(events)
	this.Publisher.Unsubscribe(cancel_requests)

	this.log.Debug2("<grpc.service.input.ListenForInputEvents> Ended")

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// Return opened devices

func (this *service) Devices(ctx context.Context, _ *pb.EmptyRequest) (*pb.InputDevices, error) {
	this.log.Debug2("<grpc.service.input.Devices>{ }")
	devices := this.input.GetOpenDevices()
	return &pb.InputDevices{
		Device: make([]*pb.InputDevice, len(devices)),
	}, nil
}
