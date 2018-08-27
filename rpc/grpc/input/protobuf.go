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
		Device:     evt.Device(),
		ScanCode:   evt.ScanCode(),
		KeyCode:    uint32(evt.KeyCode()),
		KeyState:   uint32(evt.KeyState()),
		Position:   toProtobufPoint(evt.Position()),
		Relative:   toProtobufPoint(evt.Relative()),
		Slot:       uint32(evt.Slot()),
	}
	return input_event
}

func fromProtobufInputEvent(source gopi.Driver, evt *pb.InputEvent) gopi.InputEvent {
	ts, _ := ptype.Duration(evt.Ts)
	return input.NewInputEvent(
		source, ts, gopi.InputDeviceType(evt.DeviceType), gopi.InputEventType(evt.EventType),
		fromProtobufPoint(evt.Position), fromProtobufPoint(evt.Relative),
		gopi.KeyCode(evt.KeyCode), gopi.KeyState(evt.KeyState), uint32(evt.ScanCode),
		evt.Device, uint(evt.Slot),
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
	input_device := &pb.InputDevice{}
	return input_device
}
