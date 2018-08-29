/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2018
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package helloworld

import (
	"fmt"
	"time"

	// Frameworks
	"github.com/djthorpe/gopi"
	"github.com/djthorpe/gopi-input/sys/input"
	"github.com/djthorpe/gopi/util/event"
)

var (
	ts = time.Now()
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Input device
type InputDevice struct{}

// Represents an input device
type device struct {
	log  gopi.Logger
	done chan struct{}

	// Publisher
	event.Publisher
}

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

// Create new InputDevice object or return error
func (config InputDevice) Open(log gopi.Logger) (gopi.Driver, error) {
	log.Debug("<sys.inputdevice.Helloworld.Open>{ }")

	this := new(device)
	this.log = log
	this.done = make(chan struct{})

	// Start emitting events
	go this.emitEvents()

	// Success
	return this, nil
}

// Close InputDevice
func (this *device) Close() error {
	this.log.Debug("<sys.inputdevice.Helloworld.Close>{ }")

	if this.done != nil {
		this.done <- gopi.DONE
		<-this.done
	}

	// Close publisher
	this.Publisher.Close()

	// Release resources
	this.done = nil

	// return success
	return nil
}

// Stringify
func (this *device) String() string {
	return fmt.Sprintf("<sys.inputdevice.Helloworld>{}")
}

////////////////////////////////////////////////////////////////////////////////
// INTERFACE

func (this *device) Name() string {
	return "sys.inputdevice.Helloworld"
}

func (this *device) Type() gopi.InputDeviceType {
	return gopi.INPUT_TYPE_NONE
}

func (this *device) Bus() gopi.InputDeviceBus {
	return gopi.INPUT_BUS_NONE
}

func (this *device) Position() gopi.Point {
	return gopi.ZeroPoint
}

func (this *device) SetPosition(gopi.Point) {}

func (this *device) KeyState() gopi.KeyState {
	return gopi.KEYSTATE_NONE
}

func (this *device) SetKeyState(flags gopi.KeyState, state bool) error {
	return gopi.ErrNotImplemented
}

func (this *device) Matches(name string, device_type gopi.InputDeviceType, device_bus gopi.InputDeviceBus) bool {
	if name != "" && name != this.Name() {
		return false
	}
	if device_type != gopi.INPUT_TYPE_NONE && device_type != gopi.INPUT_TYPE_ANY {
		return false
	}
	if device_bus != gopi.INPUT_BUS_NONE && device_bus != gopi.INPUT_BUS_ANY {
		return false
	}
	return true
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (this *device) emitEvents() {
	timer := time.NewTicker(1 * time.Second)
FOR_LOOP:
	for {
		select {
		case <-timer.C:
			this.emitEvent()
		case <-this.done:
			break FOR_LOOP
		}
	}
	close(this.done)
}

func (this *device) emitEvent() {
	this.Emit(input.NewInputEvent(this, time.Now().Sub(ts), this.Type(),
		gopi.INPUT_EVENT_KEYPRESS, this.Position(), gopi.ZeroPoint, gopi.KEYCODE_ENTER, this.KeyState(), 0, 0, 0))
}
