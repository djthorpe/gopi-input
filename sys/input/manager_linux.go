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
	"fmt"

	// Frameworks

	"github.com/djthorpe/gopi"
	"github.com/djthorpe/gopi/sys/hw/linux"
	"github.com/djthorpe/gopi/util/event"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Input manager
type InputManager struct {
	// Filepoller
	FilePoll linux.FilePollInterface

	// Whether to try and get exclusivity when opening devices
	Exclusive bool
}

// Driver of multiple input devices
type manager struct {
	// Logger
	log gopi.Logger

	// Filepoller
	filepoll linux.FilePollInterface

	// Whether to try and get exclusivity when opening devices
	exclusive bool

	// List of open devices
	devices []gopi.InputDevice

	// event merger (also acts as publisher)
	event.Merger
}

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

func (config InputManager) Open(log gopi.Logger) (gopi.Driver, error) {
	log.Debug("<sys.input.InputManager.Open>{ exclusive=%v }", config.Exclusive)

	// create new input device manager
	this := new(manager)

	if config.FilePoll == nil {
		return nil, gopi.ErrBadParameter
	}

	this.exclusive = config.Exclusive
	this.log = log
	this.filepoll = config.FilePoll
	this.devices = make([]gopi.InputDevice, 0)

	// success
	return this, nil
}

// Close Input driver
func (this *manager) Close() error {
	this.log.Debug("<sys.input.InputManager.Close>{ }")

	// Close all open devices
	for _, device := range this.devices {
		if device != nil {
			if err := this.CloseDevice(device); err != nil {
				this.log.Warn("<sys.input.InputManager.Close> Error: %v", err)
			}
		}
	}

	// Close publisher
	this.log.Debug("Closing merger")
	this.Merger.Close()
	this.log.Debug("Merger closed")

	// Empty
	this.filepoll = nil
	this.devices = nil

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *manager) String() string {
	return fmt.Sprintf("<sys.input.InputManager>{ exclusive=%v }", this.exclusive)
}

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE DEVICES

// OpenDevicesByName can be called often in order to open any newly plugged in
// devices. It will only return any newly opened devices.
func (this *manager) OpenDevicesByName(alias string, flags gopi.InputDeviceType, bus gopi.InputDeviceBus) ([]gopi.InputDevice, error) {
	this.log.Debug2("<sys.input.InputManager.OpenDevicesByName>{ alias='%v' flags=%v bus=%v }", alias, flags, bus)

	opened_devices := make([]gopi.InputDevice, 0)
	new_devices := make([]gopi.InputDevice, 0)

	// Discover devices using evFind and add any new ones to the new_devices
	// array, they are left in an opened state
	evFind(func(path string) {
		this.log.Debug2("<evFind>{ path=%v }", path)
		// Don't consider devices which are already opened
		if this.deviceByPath(path) == nil {
			if input_device, err := gopi.Open(InputDevice{Path: path, Exclusive: this.exclusive, FilePoll: this.filepoll}, this.log); err != nil {
				this.log.Warn("OpenDevicesByName: %v: %v", path, err)
			} else {
				this.log.Debug2("OpenDevicesByName: Adding device %v", input_device)
				new_devices = append(new_devices, input_device.(gopi.InputDevice))
			}
		}
	})

	// Now check devices against filters and close devices which don't match
	for _, device := range new_devices {
		if device.Matches(alias, flags, bus) {
			opened_devices = append(opened_devices, device)
		} else if err := device.Close(); err != nil {
			this.log.Warn("OpenDevicesByName: %v", err)
		}
	}

	// Subscribe to events from device
	for _, device := range opened_devices {
		this.Merger.Merge(device)
		this.devices = append(this.devices, device)
	}

	return opened_devices, nil
}

func (this *manager) CloseDevice(device gopi.InputDevice) error {
	this.log.Debug2("<sys.input.InputManager.CloseDevice>{ device=%v }", device)

	// Find device in array of devices
	found := -1
	for i, d := range this.devices {
		if d == device {
			found = i
		}
	}
	if found == -1 {
		return gopi.ErrNotFound
	}

	// Unsubscribe from events
	this.Merger.Unmerge(device)

	// Close device
	if err := device.Close(); err != nil {
		return err
	}

	// Remove device from array (nil)
	this.devices[found] = nil

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// RETURN OPENED DEVICES

func (this *manager) GetOpenDevices() []gopi.InputDevice {
	devices := make([]gopi.InputDevice, 0, len(this.devices))
	for _, device := range this.devices {
		if device != nil {
			devices = append(devices, device)
		}
	}
	return devices
}

////////////////////////////////////////////////////////////////////////////////
// ADD NEW INPUT DEVICE

func (this *manager) AddDevice(device gopi.InputDevice) error {
	// TODO: This method is currently not implemented
	return gopi.ErrNotImplemented
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// deviceByPath returns an opened device based on it's path, assuming
// it is a linux device or returns nil if a device with this path is
// not found
func (this *manager) deviceByPath(path string) gopi.InputDevice {
	for _, d := range this.devices {
		if linux_device, is_linux := d.(*device); is_linux {
			if linux_device.path == path {
				return d
			}
		}
	}
	return nil
}
