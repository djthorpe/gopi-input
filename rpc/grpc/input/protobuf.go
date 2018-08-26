/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2016-2018
	All Rights Reserved
	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package input

import (
	// Frameworks
	gopi "github.com/djthorpe/gopi"

	// Protocol buffers
	pb "github.com/djthorpe/gopi-input/rpc/protobuf/input"
	ptype "github.com/golang/protobuf/ptypes"
)

////////////////////////////////////////////////////////////////////////////////
// TO PROTOBUF

func toProtobufInputEvent(evt gopi.InputEvent) *pb.InputEvent {
	input_event := &pb.InputEvent{
		Ts:         ptype.DurationProto(evt.Timestamp()),
		DeviceType: pb.InputDeviceType(evt.DeviceType()),
		EventType:  pb.InputEventType(evt.EventType()),
		Device:     evt.Device(),
		Scancode:   evt.Scancode(),
		Position:   toProtobufPoint(evt.Position()),
		Relative:   toProtobufPoint(evt.Relative()),
		Slot:       uint32(evt.Slot()),
	}
	return input_event
}

func fromProtobufInputEvent(evt *pb.InputEvent) gopi.InputEvent {
	// TODO
	return nil
}

func toProtobufPoint(pt gopi.Point) *pb.Point {
	return &pb.Point{
		X: pt.X,
		Y: pt.Y,
	}
}

func toProtobufInputDevice(device gopi.InputDevice) *pb.InputDevice {
	input_device := &pb.InputDevice{}
	return input_device
}
