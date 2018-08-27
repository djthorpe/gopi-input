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
	"strings"

	// Frameworks
	"github.com/djthorpe/gopi"
)

////////////////////////////////////////////////////////////////////////////////
// GLOBAL VARIABLES

var (
	map_type = make(map[string]gopi.InputDeviceType)
	map_bus  = make(map[string]gopi.InputDeviceBus)
)

////////////////////////////////////////////////////////////////////////////////
// INIT

func init() {
	// Register rpc/service/input
	gopi.RegisterModule(gopi.Module{
		Name:     "rpc/service/input",
		Type:     gopi.MODULE_TYPE_SERVICE,
		Requires: []string{"rpc/server", "input"},
		Config: func(config *gopi.AppConfig) {
			keys_type := make([]string, 0)
			keys_bus := make([]string, 0)

			// Gather device types
			for t := gopi.INPUT_TYPE_NONE; t < gopi.INPUT_TYPE_ANY; t++ {
				s := fmt.Sprint(t)
				if strings.HasPrefix(s, "INPUT_TYPE_") {
					k := strings.ToLower(strings.TrimPrefix(s, "INPUT_TYPE_"))
					map_type[k] = t
					keys_type = append(keys_type, k)
				}
			}

			// Gather device bus
			for b := gopi.INPUT_BUS_NONE; b < gopi.INPUT_BUS_ANY; b++ {
				s := fmt.Sprint(b)
				if strings.HasPrefix(s, "INPUT_BUS_") {
					k := strings.ToLower(strings.TrimPrefix(s, "INPUT_BUS_"))
					map_bus[k] = b
					keys_bus = append(keys_bus, k)
				}
			}

			config.AppFlags.FlagString("input.type", "", fmt.Sprintf("Filter by type of device (%v)", strings.Join(keys_type, ",")))
			config.AppFlags.FlagString("input.bus", "", fmt.Sprintf("Filter by one or more device busses (%v)", strings.Join(keys_bus, ",")))
			config.AppFlags.FlagString("input.name", "", fmt.Sprintf("Filter by device name or alias"))

		},
		New: func(app *gopi.AppInstance) (gopi.Driver, error) {
			config := Service{
				Server:       app.ModuleInstance("rpc/server").(gopi.RPCServer),
				InputManager: app.Input,
			}
			if device_name, exists := app.AppFlags.GetString("input.name"); exists {
				config.DeviceName = device_name
			}
			if device_type_, exists := app.AppFlags.GetString("input.type"); exists {
				if device_type, err := FilterByDeviceType(device_type_); err != nil {
					return nil, err
				} else {
					config.DeviceType = device_type
				}
			}
			if device_bus_, exists := app.AppFlags.GetString("input.bus"); exists {
				if device_bus, err := FilterByDeviceBus(device_bus_); err != nil {
					return nil, err
				} else {
					config.DeviceBus = device_bus
				}
			}
			return gopi.Open(config, app.Logger)
		},
	})

	// Register rpc/client/input
	gopi.RegisterModule(gopi.Module{
		Name:     "rpc/client/input",
		Type:     gopi.MODULE_TYPE_CLIENT,
		Requires: []string{"rpc/clientpool"},
		Run: func(app *gopi.AppInstance, _ gopi.Driver) error {
			if clientpool := app.ModuleInstance("rpc/clientpool").(gopi.RPCClientPool); clientpool == nil {
				return gopi.ErrAppError
			} else {
				clientpool.RegisterClient("gopi.Input", NewInputClient)
				return nil
			}
		},
	})
}

////////////////////////////////////////////////////////////////////////////////
// Filter by device type & bus

func FilterByDeviceType(device_type string) (gopi.InputDeviceType, error) {
	type_flags := gopi.INPUT_TYPE_NONE
	for _, k := range strings.Split(device_type, ",") {
		k := strings.ToLower(strings.TrimSpace(k))
		if v, exists := map_type[k]; exists {
			type_flags |= v
		} else {
			return 0, fmt.Errorf("Invalid type: '%v'", k)
		}
	}
	return type_flags, nil
}

func FilterByDeviceBus(device_bus string) (gopi.InputDeviceBus, error) {
	bus_flags := gopi.INPUT_BUS_NONE
	for _, k := range strings.Split(device_bus, ",") {
		k := strings.ToLower(strings.TrimSpace(k))
		if v, exists := map_bus[k]; exists {
			bus_flags |= v
		} else {
			return 0, fmt.Errorf("Invalid bus: '%v'", k)
		}
	}
	return bus_flags, nil
}
