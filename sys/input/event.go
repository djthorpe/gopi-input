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

	// Frameworks
	"github.com/djthorpe/gopi"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Input event
type input_event struct {
	source       gopi.Driver
	timestamp    time.Duration
	device       gopi.InputDeviceType
	event        gopi.InputEventType
	position     gopi.Point
	rel_position gopi.Point
	key_code     gopi.KeyCode
	key_state    gopi.KeyState
	scan_code    uint32
	device_id    uint32
	slot         uint
}

////////////////////////////////////////////////////////////////////////////////
// gopi.InputEvent INTERFACE

func NewInputEvent(source gopi.InputDevice, timestamp time.Duration, event_type gopi.InputEventType, key_code gopi.KeyCode, scan_code uint32, slot uint, position gopi.Point, rel_position gopi.Point) gopi.InputEvent {
	return &input_event{
		source:       source,
		timestamp:    timestamp,
		device:       source.Type(),
		event:        event_type,
		position:     position,
		rel_position: rel_position,
		key_code:     key_code,
		key_state:    source.KeyState(),
		scan_code:    scan_code,
		slot:         slot,
	}
}

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

func (this *input_event) EventType() gopi.InputEventType {
	return this.event
}

func (this *input_event) KeyCode() gopi.KeyCode {
	return this.key_code
}

func (this *input_event) KeyState() gopi.KeyState {
	return this.key_state
}

func (this *input_event) ScanCode() uint32 {
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
		return fmt.Sprintf("<sys.input.InputEvent>{ type=%v device=%v key_code=%v key_state=%v scan_code=0x%08X ts=%v }", this.event, this.device, this.key_code, this.key_state, this.scan_code, this.timestamp)
	case gopi.INPUT_EVENT_TOUCHPRESS, gopi.INPUT_EVENT_TOUCHRELEASE:
		return fmt.Sprintf("<sys.input.InputEvent>{ type=%v device=%v key_code=%v key_state=%v slot=%v position=%v ts=%v }", this.event, this.device, this.key_code, this.key_state, this.position, this.slot, this.timestamp)
	case gopi.INPUT_EVENT_TOUCHPOSITION:
		return fmt.Sprintf("<sys.input.InputEvent>{ type=%v device=%v slot=%v position=%v ts=%v }", this.event, this.device, this.position, this.slot, this.timestamp)
	default:
		return fmt.Sprintf("<sys.input.InputEvent>{ type=%v device=%v ts=%v }", this.event, this.device, this.timestamp)
	}
}
