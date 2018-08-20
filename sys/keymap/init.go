/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2017
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package keymap

import (
	// Frameworks
	"github.com/djthorpe/gopi"
)

////////////////////////////////////////////////////////////////////////////////
// INIT

func init() {
	// Register InputManager
	gopi.RegisterModule(gopi.Module{
		Name:     "keymap/manager",
		Requires: []string{"input"},
		Type:     gopi.MODULE_TYPE_KEYMAP,
		Config: func(config *gopi.AppConfig) {
			config.AppFlags.FlagString("keymap.path", "", "Path to keymap files")
			config.AppFlags.FlagString("keymap.ext", DEFAULT_EXT, "Keymap file extension")
		},
		New: func(app *gopi.AppInstance) (gopi.Driver, error) {
			path, _ := app.AppFlags.GetString("keymap.path")
			ext, _ := app.AppFlags.GetString("keymap.ext")
			return gopi.Open(KeymapManager{
				InputManager: app.Input,
				Root:         path,
				Ext:          ext,
			}, app.Logger)
		},
	})
}
