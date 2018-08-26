// +build linux

/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2017
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package input

import (
	"github.com/djthorpe/gopi"
	"github.com/djthorpe/gopi/sys/hw/linux"
)

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	// Pattern for finding event-driven input devices
	INPUT_PATH_DEVICES = "/sys/class/input/event*"

	// Maximum multi-touch slots
	INPUT_MAX_MULTITOUCH_SLOTS = 32
)

////////////////////////////////////////////////////////////////////////////////
// INIT

func init() {
	// Register InputManager
	gopi.RegisterModule(gopi.Module{
		Name:     "linux/input",
		Requires: []string{"linux/filepoll"},
		Type:     gopi.MODULE_TYPE_INPUT,
		Config: func(config *gopi.AppConfig) {
			config.AppFlags.FlagBool("input.exclusive", true, "Input device exclusivity")
		},
		New: func(app *gopi.AppInstance) (gopi.Driver, error) {
			exclusive, _ := app.AppFlags.GetBool("input.exclusive")
			return gopi.Open(InputManager{
				FilePoll:  app.ModuleInstance("linux/filepoll").(linux.FilePollInterface),
				Exclusive: exclusive,
			}, app.Logger)
		},
	})
}
