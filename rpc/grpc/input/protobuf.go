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
	event "github.com/djthorpe/gopi/util/event"

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

func toProtobufPoint(pt gopi.Point) *pb.Point {
	return &pb.Point{
		X: pt.X,
		Y: pt.Y,
	}
}

func toProtobufInputDevice(device gopi.InputDevice) *pb.InputDevice {
	if device == nil {
		return nil
	}
	return &pb.InputDevice{
		DeviceName:     device.Name(),
		DeviceType:     pb.InputDeviceType(device.Type()),
		DeviceBus:      pb.InputDeviceBus(device.Bus()),
		DevicePosition: toProtobufPoint(device.Position()),
	}
}

////////////////////////////////////////////////////////////////////////////////
// FROM PROTOBUF

func fromProtobufInputEvent(source gopi.InputDevice, evt *pb.InputEvent) gopi.InputEvent {
	ts, _ := ptype.Duration(evt.Ts)
	// TODO: Set device_type and key_state from protobuf
	return input.NewInputEvent(
		source, ts, gopi.InputEventType(evt.EventType),
		gopi.KeyCode(evt.KeyCode), uint32(evt.ScanCode),
		uint(evt.Slot), fromProtobufPoint(evt.Position), fromProtobufPoint(evt.Relative),
	)
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

func fromProtobufInputDevice(pb_device *pb.InputDevice) gopi.InputDevice {
	return &device{pb_device, event.Publisher{}}
}

////////////////////////////////////////////////////////////////////////////////
// INPUTDEVICE INTERFACE IMPLEMENTATION

type device struct {
	*pb.InputDevice
	event.Publisher
}

func (this *device) Name() string {
	return this.DeviceName
}

func (this *device) Type() gopi.InputDeviceType {
	return gopi.InputDeviceType(this.DeviceType)
}

func (this *device) Bus() gopi.InputDeviceBus {
	return gopi.InputDeviceBus(this.DeviceBus)
}

func (this *device) Position() gopi.Point {
	return fromProtobufPoint(this.DevicePosition)
}

// Non-functional methods
func (this *device) Close() error {
	return gopi.ErrNotImplemented
}

func (this *device) SetPosition(pt gopi.Point) {
	this.DevicePosition = toProtobufPoint(pt)
}

func (this *device) KeyState() gopi.KeyState {
	return 0
}

func (this *device) SetKeyState(gopi.KeyState, bool) error {
	return gopi.ErrNotImplemented
}

func (this *device) Matches(string, gopi.InputDeviceType, gopi.InputDeviceBus) bool {
	return false
}
