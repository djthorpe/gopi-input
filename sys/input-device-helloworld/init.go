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

	"github.com/djthorpe/gopi"
)

////////////////////////////////////////////////////////////////////////////////
// INIT

func init() {
	// Register InputManager
	gopi.RegisterModule(gopi.Module{
		Name:     "input/device/helloworld",
		Requires: []string{"input"},
		Type:     gopi.MODULE_TYPE_OTHER,
		New: func(app *gopi.AppInstance) (gopi.Driver, error) {
			return gopi.Open(InputDevice{}, app.Logger)
		},
		Run: func(app *gopi.AppInstance, device gopi.Driver) error {
			if app.Input == nil {
				return fmt.Errorf("Missing InputManager module instance")
			} else if err := app.Input.AddDevice(device.(gopi.InputDevice)); err != nil {
				return err
			} else {
				return nil
			}
		},
	})
}
