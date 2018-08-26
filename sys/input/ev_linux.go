// +build linux

/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2016-2018
	All Rights Reserved

	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package input

import (
	"encoding/binary"
	"io"
	"os"
	"path"
	"path/filepath"
	"time"
	"unsafe"

	// Frameworks

	"github.com/djthorpe/gopi"
	"github.com/djthorpe/gopi/sys/hw/linux"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type evType uint16
type evKeyCode uint16
type evKeyAction uint32
type evLEDState uint8

type evEvent struct {
	Second      uint32
	Microsecond uint32
	Type        evType
	Code        evKeyCode
	Value       uint32
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

// Event types
// See https://www.kernel.org/doc/Documentation/input/event-codes.txt
const (
	EV_SYN       evType = 0x0000 // Used as markers to separate events
	EV_KEY       evType = 0x0001 // Used to describe state changes of keyboards, buttons
	EV_REL       evType = 0x0002 // Used to describe relative axis value changes
	EV_ABS       evType = 0x0003 // Used to describe absolute axis value changes
	EV_MSC       evType = 0x0004 // Miscellaneous uses that didn't fit anywhere else
	EV_SW        evType = 0x0005 // Used to describe binary state input switches
	EV_LED       evType = 0x0011 // Used to turn LEDs on devices on and off
	EV_SND       evType = 0x0012 // Sound output, such as buzzers
	EV_REP       evType = 0x0014 // Enables autorepeat of keys in the input core
	EV_FF        evType = 0x0015 // Sends force-feedback effects to a device
	EV_PWR       evType = 0x0016 // Power management events
	EV_FF_STATUS evType = 0x0017 // Device reporting of force-feedback effects back to the host
	EV_MAX       evType = 0x001F
)

const (
	EV_CODE_X        evKeyCode = 0x0000
	EV_CODE_Y        evKeyCode = 0x0001
	EV_CODE_SCANCODE evKeyCode = 0x0004 // Keyboard scan code
	EV_CODE_SLOT     evKeyCode = 0x002F // Slot for multi touch positon
	EV_CODE_SLOT_X   evKeyCode = 0x0035 // X for multi touch position
	EV_CODE_SLOT_Y   evKeyCode = 0x0036 // Y for multi touch position
	EV_CODE_SLOT_ID  evKeyCode = 0x0039 // Unique ID for multi touch position
)

const (
	EV_VALUE_KEY_NONE   evKeyAction = 0x00000000
	EV_VALUE_KEY_UP     evKeyAction = 0x00000000
	EV_VALUE_KEY_DOWN   evKeyAction = 0x00000001
	EV_VALUE_KEY_REPEAT evKeyAction = 0x00000002
)

// LED Constants
const (
	EV_LED_NUML     evLEDState = 0x00
	EV_LED_CAPSL    evLEDState = 0x01
	EV_LED_SCROLLL  evLEDState = 0x02
	EV_LED_COMPOSE  evLEDState = 0x03
	EV_LED_KANA     evLEDState = 0x04
	EV_LED_SLEEP    evLEDState = 0x05
	EV_LED_SUSPEND  evLEDState = 0x06
	EV_LED_MUTE     evLEDState = 0x07
	EV_LED_MISC     evLEDState = 0x08
	EV_LED_MAIL     evLEDState = 0x09
	EV_LED_CHARGING evLEDState = 0x0A
	EV_LED_MAX      evLEDState = 0x0F
)

////////////////////////////////////////////////////////////////////////////////
// CALLBACK

func (this *device) evReceive(dev *os.File, mode linux.FilePollMode) {
	// Read raw event data
	var raw_event evEvent
	if err := binary.Read(dev, binary.LittleEndian, &raw_event); err == io.EOF {
		return
	} else if err != nil {
		this.log.Error("sys.input.linux.InputDevice.Receive: %v", err)
		return
	}

	// Decode the event
	switch raw_event.Type {
	case EV_SYN:
		if evt := this.evDecodeSyn(&raw_event); evt != nil {
			this.Emit(evt)
		}
	case EV_KEY:
		this.evDecodeKey(&raw_event)
	case EV_ABS:
		if evt := this.evDecodeAbs(&raw_event); evt != nil {
			this.Emit(evt)
		}
	case EV_REL:
		this.evDecodeRel(&raw_event)
	case EV_MSC:
		this.evDecodeMsc(&raw_event)
	case EV_LED:
		// Ignore EV_LED events
	default:
		this.log.Warn("sys.input.linux.InputDevice.Receive: Ignoring event with type %v", raw_event.Type)
	}
}

////////////////////////////////////////////////////////////////////////////////
// DECODE

// Decode the EV_SYN syncronization raw event.
func (this *device) evDecodeSyn(raw_event *evEvent) gopi.InputEvent {
	evt := &input_event{
		source:    this,
		timestamp: time.Duration(time.Duration(raw_event.Second)*time.Second + time.Duration(raw_event.Microsecond)*time.Microsecond),
		device:    this.device_type,
		device_id: this.device_id,
	}

	// Mouse and keyboard movements
	if this.rel_position.Equals(gopi.ZeroPoint) == false {
		evt.event = gopi.INPUT_EVENT_RELPOSITION
		evt.rel_position = this.rel_position
		this.rel_position = gopi.ZeroPoint
		this.last_position = this.position
	} else if this.position.Equals(this.last_position) == false {
		evt.event = gopi.INPUT_EVENT_ABSPOSITION
		this.last_position = this.position
	} else if this.key_action == EV_VALUE_KEY_UP {
		evt.event = gopi.INPUT_EVENT_KEYRELEASE
		evt.key_code = gopi.KeyCode(this.key_code)
		evt.scan_code = this.scan_code
		this.key_action = EV_VALUE_KEY_NONE
	} else if this.key_action == EV_VALUE_KEY_DOWN {
		evt.event = gopi.INPUT_EVENT_KEYPRESS
		evt.key_code = gopi.KeyCode(this.key_code)
		evt.scan_code = this.scan_code
		this.key_action = EV_VALUE_KEY_NONE
	} else if this.key_action == EV_VALUE_KEY_REPEAT {
		evt.event = gopi.INPUT_EVENT_KEYREPEAT
		evt.key_code = gopi.KeyCode(this.key_code)
		evt.scan_code = this.scan_code
		this.key_action = EV_VALUE_KEY_NONE
	} else {
		return nil
	}

	// Check for multi-touch positional changes
	/*slot := device.slots[device.slot]
	if slot.active {
		this.log.Debug("SLOT=%v POSITION=%v", device.slot, slot.position)
	}*/

	return evt
}

func (this *device) evDecodeKey(raw_event *evEvent) {
	this.key_code = evKeyCode(raw_event.Code)
	this.key_action = evKeyAction(raw_event.Value)

	// Set the device state from the key action. For the locks (Caps, Scroll
	// and Num) we also reflect the change with the LED and "flip" the state
	// from the current state.
	key_state := gopi.KEYSTATE_NONE
	switch gopi.KeyCode(this.key_code) {
	case gopi.KEYCODE_CAPSLOCK:
		// Flip CAPS LOCK state and set LED
		if this.key_action == EV_VALUE_KEY_DOWN {
			this.key_state ^= gopi.KEYSTATE_CAPSLOCK
			evSetLEDState(this.handle, EV_LED_CAPSL, this.key_state&gopi.KEYSTATE_CAPSLOCK != gopi.KEYSTATE_NONE)
		}
	case gopi.KEYCODE_NUMLOCK:
		// Flip NUM LOCK state and set LED
		if this.key_action == EV_VALUE_KEY_DOWN {
			this.key_state ^= gopi.KEYSTATE_NUMLOCK
			evSetLEDState(this.handle, EV_LED_NUML, this.key_state&gopi.KEYSTATE_NUMLOCK != gopi.KEYSTATE_NONE)
		}
	case gopi.KEYCODE_SCROLLLOCK:
		// Flip SCROLL LOCK state and set LED
		if this.key_action == EV_VALUE_KEY_DOWN {
			this.key_state ^= gopi.KEYSTATE_SCROLLLOCK
			evSetLEDState(this.handle, EV_LED_SCROLLL, this.key_state&gopi.KEYSTATE_SCROLLLOCK != gopi.KEYSTATE_NONE)
		}
	case gopi.KEYCODE_LEFTSHIFT:
		key_state = gopi.KEYSTATE_LEFTSHIFT
	case gopi.KEYCODE_RIGHTSHIFT:
		key_state = gopi.KEYSTATE_RIGHTSHIFT
	case gopi.KEYCODE_LEFTCTRL:
		key_state = gopi.KEYSTATE_LEFTCTRL
	case gopi.KEYCODE_RIGHTCTRL:
		key_state = gopi.KEYSTATE_RIGHTCTRL
	case gopi.KEYCODE_LEFTALT:
		key_state = gopi.KEYSTATE_LEFTALT
	case gopi.KEYCODE_RIGHTALT:
		key_state = gopi.KEYSTATE_RIGHTALT
	case gopi.KEYCODE_LEFTMETA:
		key_state = gopi.KEYSTATE_LEFTMETA
	case gopi.KEYCODE_RIGHTMETA:
		key_state = gopi.KEYSTATE_RIGHTMETA
	}

	// Set device state from key action
	if key_state != gopi.KEYSTATE_NONE {
		if this.key_action == EV_VALUE_KEY_DOWN || this.key_action == EV_VALUE_KEY_REPEAT {
			this.key_state |= key_state
		} else if this.key_action == EV_VALUE_KEY_UP {
			this.key_state &= (gopi.KEYSTATE_MAX ^ key_state)
		}
	}
}

func (this *device) evDecodeAbs(raw_event *evEvent) gopi.InputEvent {
	if raw_event.Code == EV_CODE_X {
		this.position.X = float32(int32(raw_event.Value))
	} else if raw_event.Code == EV_CODE_Y {
		this.position.Y = float32(int32(raw_event.Value))
	} else if raw_event.Code == EV_CODE_SLOT {
		this.slot = raw_event.Value
	} else if raw_event.Code == EV_CODE_SLOT_ID || raw_event.Code == EV_CODE_SLOT_X || raw_event.Code == EV_CODE_SLOT_Y {
		switch {
		case this.slot < uint32(0) || this.slot >= INPUT_MAX_MULTITOUCH_SLOTS:
			this.log.Warn("evDecodeAbs: Ignoring out-of-range slot %v", this.slot)
		case raw_event.Code == EV_CODE_SLOT_ID:
			return this.evDecodeAbsTouch(raw_event)
		case raw_event.Code == EV_CODE_SLOT_X:
			this.slots[this.slot].position.X = float32(int32(raw_event.Value))
			this.slots[this.slot].active = true
		case raw_event.Code == EV_CODE_SLOT_Y:
			this.slots[this.slot].position.Y = float32(int32(raw_event.Value))
			this.slots[this.slot].active = true
		}
	} else {
		this.log.Warn("evDecodeAbs: %v Ignoring code %v", raw_event.Type, raw_event.Code)
	}
	return nil
}

func (this *device) evDecodeAbsTouch(raw_event *evEvent) gopi.InputEvent {
	evt := &input_event{
		source:    this,
		timestamp: time.Duration(time.Duration(raw_event.Second)*time.Second + time.Duration(raw_event.Microsecond)*time.Microsecond),
		device:    this.device_type,
		device_id: this.device_id,
	}

	// Decode the slot_id, if -1 then this is the release for a slot
	if slot_id := int16(raw_event.Value); slot_id == -1 {
		this.slots[this.slot].active = false
		evt.event = gopi.INPUT_EVENT_TOUCHRELEASE
	} else if slot_id < INPUT_MAX_MULTITOUCH_SLOTS {
		this.slots[this.slot].active = true
		this.slots[this.slot].id = slot_id
		evt.event = gopi.INPUT_EVENT_TOUCHPRESS
	} else {
		this.log.Warn("evDecodeAbsTouch: %v Ignoring slot %v", raw_event.Type, slot_id)
	}

	// Populate the slot and keycode
	evt.slot = uint(this.slot)
	evt.key_code = gopi.KEYCODE_BTNTOUCH

	// Return the event to emit
	return evt
}

func (this *device) evDecodeRel(raw_event *evEvent) {
	switch raw_event.Code {
	case EV_CODE_X:
		this.position.X = this.position.X + float32(int32(raw_event.Value))
		this.rel_position.X = float32(int32(raw_event.Value))
	case EV_CODE_Y:
		this.position.Y = this.position.Y + float32(int32(raw_event.Value))
		this.rel_position.Y = float32(int32(raw_event.Value))
	default:
		this.log.Warn("evDecodeRel: %v Ignoring code %v", raw_event.Type, raw_event.Code)
	}
}

func (this *device) evDecodeMsc(raw_event *evEvent) {
	switch raw_event.Code {
	case EV_CODE_SCANCODE:
		this.scan_code = raw_event.Value
	default:
		this.log.Warn("evDecodeMsc: %v Ignoring code=%v, value=%v", raw_event.Type, raw_event.Code, raw_event.Value)
	}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// evFind finds all input devices on the path and calls a callback function
// for each one with the device file path
func evFind(callback func(string)) error {
	files, err := filepath.Glob(INPUT_PATH_DEVICES)
	if err != nil {
		return err
	}
	for _, file := range files {
		callback(path.Clean(path.Join("/", "dev", "input", path.Base(file))))
	}
	return nil
}

// evSupportsEventType returns true if all event types are supported
// else returns false
func evSupportsEventType(capabilities []evType, types ...evType) bool {
	count := 0
	for _, capability := range capabilities {
		for _, typ := range types {
			if typ == capability {
				count = count + 1
			}
		}
	}
	return (count == len(types))
}

// evSetLEDState sets a single LED state
func evSetLEDState(handle *os.File, led evLEDState, state bool) error {
	var event evEvent

	event.Type = EV_LED
	event.Code = evKeyCode(led)

	if state {
		event.Value = 1
	}
	if err := binary.Write(handle, binary.LittleEndian, &event); err != nil {
		return err
	}
	return nil
}

// evGetLEDState gets LED states as an array of LED's which are on
func evGetLEDState(handle *os.File) ([]evLEDState, error) {
	evbits := new([MAX_IOCTL_SIZE_BYTES]byte)
	err := evIoctl(handle.Fd(), uintptr(EVIOCGLED), unsafe.Pointer(evbits))
	if err != 0 {
		return nil, err
	}
	states := make([]evLEDState, 0, EV_LED_MAX)
	// Shift bits to get the state of each LED value
OuterLoop:
	for i := 0; i < len(evbits); i++ {
		evbyte := evbits[i]
		for j := 0; j < 8; j++ {
			state := evLEDState(i<<3 + j)
			switch {
			case state >= EV_LED_MAX:
				break OuterLoop
			case evbyte&0x01 != 0x00:
				states = append(states, state)
			}
			evbyte >>= 1
		}
	}
	return states, nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (t evType) String() string {
	switch t {
	case EV_SYN:
		return "EV_SYN"
	case EV_KEY:
		return "EV_KEY"
	case EV_REL:
		return "EV_REL"
	case EV_ABS:
		return "EV_ABS"
	case EV_MSC:
		return "EV_MSC"
	case EV_SW:
		return "EV_SW"
	case EV_LED:
		return "EV_LED"
	case EV_SND:
		return "EV_SND"
	case EV_REP:
		return "EV_REP"
	case EV_FF:
		return "EV_FF"
	case EV_PWR:
		return "EV_PWR"
	case EV_FF_STATUS:
		return "EV_FF_STATUS"
	default:
		return "[?? Unknown evType value]"
	}
}

func (s evLEDState) String() string {
	switch s {
	case EV_LED_NUML:
		return "EV_LED_NUML"
	case EV_LED_CAPSL:
		return "EV_LED_CAPSL"
	case EV_LED_SCROLLL:
		return "EV_LED_SCROLLL"
	case EV_LED_COMPOSE:
		return "EV_LED_COMPOSE"
	case EV_LED_KANA:
		return "EV_LED_KANA"
	case EV_LED_SLEEP:
		return "EV_LED_SLEEP"
	case EV_LED_SUSPEND:
		return "EV_LED_SUSPEND"
	case EV_LED_MUTE:
		return "EV_LED_MUTE"
	case EV_LED_MISC:
		return "EV_LED_MISC"
	case EV_LED_MAIL:
		return "EV_LED_MAIL"
	case EV_LED_CHARGING:
		return "EV_LED_CHARGING"
	default:
		return "[?? Invalid evLEDState value]"
	}
}
