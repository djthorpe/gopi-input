/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2018
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package input

import (
	"fmt"
	"time"

	"github.com/djthorpe/gopi"
	// Frameworks
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Input event
type input_event struct {
	source       gopi.InputDevice
	timestamp    time.Duration
	device       gopi.InputDeviceType
	event        gopi.InputEventType
	position     gopi.Point
	rel_position gopi.Point
	key_code     gopi.KeyCode
	scan_code    uint32
	device_id    uint32
	slot         uint
}

////////////////////////////////////////////////////////////////////////////////
// gopi.InputEvent INTERFACE

func (this *input_event) Name() string {
	return "InputEvent"
}

func (this *input_event) Source() gopi.Driver {
	return this.source
}

func (this *input_event) Timestamp() time.Duration {
	return this.timestamp
}

func (this *input_event) DeviceType() gopi.InputDeviceType {
	return this.device
}

func (this *input_event) Device() uint32 {
	return this.device_id
}

func (this *input_event) EventType() gopi.InputEventType {
	return this.event
}

func (this *input_event) Keycode() gopi.KeyCode {
	return this.key_code
}

func (this *input_event) Scancode() uint32 {
	return this.scan_code
}

func (this *input_event) Position() gopi.Point {
	return this.position
}

func (this *input_event) Relative() gopi.Point {
	return this.rel_position
}

func (this *input_event) Slot() uint {
	return this.slot
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *input_event) String() string {
	switch this.event {
	case gopi.INPUT_EVENT_RELPOSITION:
		return fmt.Sprintf("<sys.input.InputEvent>{ type=%v device=%v relative=%v position=%v ts=%v }", this.event, this.device, this.rel_position, this.position, this.timestamp)
	case gopi.INPUT_EVENT_ABSPOSITION:
		return fmt.Sprintf("<sys.input.InputEvent>{ type=%v device=%v position=%v ts=%v }", this.event, this.device, this.position, this.timestamp)
	case gopi.INPUT_EVENT_KEYPRESS, gopi.INPUT_EVENT_KEYRELEASE, gopi.INPUT_EVENT_KEYREPEAT:
		return fmt.Sprintf("<sys.input.InputEvent>{ type=%v device=%v key_code=%v scan_code=%v ts=%v }", this.event, this.device, this.key_code, this.scan_code, this.timestamp)
	case gopi.INPUT_EVENT_TOUCHPRESS, gopi.INPUT_EVENT_TOUCHRELEASE:
		return fmt.Sprintf("<sys.input.InputEvent>{ type=%v device=%v key_code=%v slot=%v position=%v ts=%v }", this.event, this.device, this.key_code, this.position, this.slot, this.timestamp)
	case gopi.INPUT_EVENT_TOUCHPOSITION:
		return fmt.Sprintf("<sys.input.InputEvent>{ type=%v device=%v slot=%v position=%v ts=%v }", this.event, this.device, this.position, this.slot, this.timestamp)
	default:
		return fmt.Sprintf("<sys.input.InputEvent>{ type=%v device=%v ts=%v }", this.event, this.device, this.timestamp)
	}
}
