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
	input "github.com/djthorpe/gopi-input/sys/input"

	// Protocol buffers
	pb "github.com/djthorpe/gopi-input/rpc/protobuf/input"
	ptype "github.com/golang/protobuf/ptypes"
)

////////////////////////////////////////////////////////////////////////////////
// TO PROTOBUF

func toProtobufNullEvent() *pb.InputEvent {
	return &pb.InputEvent{}
}

func toProtobufInputEvent(evt gopi.InputEvent) *pb.InputEvent {
	input_event := &pb.InputEvent{
		Ts:         ptype.DurationProto(evt.Timestamp()),
		DeviceType: pb.InputDeviceType(evt.DeviceType()),
		EventType:  pb.InputEventType(evt.EventType()),
		ScanCode:   evt.ScanCode(),
		KeyCode:    uint32(evt.KeyCode()),
		KeyState:   uint32(evt.KeyState()),
		Position:   toProtobufPoint(evt.Position()),
		Relative:   toProtobufPoint(evt.Relative()),
		Slot:       uint32(evt.Slot()),
	}
	return input_event
}

func fromProtobufInputEvent(source gopi.InputDevice, evt *pb.InputEvent) gopi.InputEvent {
	ts, _ := ptype.Duration(evt.Ts)
	// TODO: Set device_type and key_state from protobuf
	return input.NewInputEvent(
		source, ts, gopi.InputEventType(evt.EventType),
		gopi.KeyCode(evt.KeyCode), uint32(evt.ScanCode),
		uint(evt.Slot), fromProtobufPoint(evt.Position), fromProtobufPoint(evt.Relative),
	)
}

func toProtobufPoint(pt gopi.Point) *pb.Point {
	return &pb.Point{
		X: pt.X,
		Y: pt.Y,
	}
}

func fromProtobufPoint(pt *pb.Point) gopi.Point {
	if pt == nil {
		return gopi.ZeroPoint
	} else {
		return gopi.Point{
			X: pt.X,
			Y: pt.Y,
		}
	}
}

func toProtobufInputDevice(device gopi.InputDevice) *pb.InputDevice {
	// TODO: Set device here
	input_device := &pb.InputDevice{}
	return input_device
}
