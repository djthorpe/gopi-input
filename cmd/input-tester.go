package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	// Frameworks
	"github.com/djthorpe/gopi"
	"github.com/olekukonko/tablewriter"

	// Modules
	_ "github.com/djthorpe/gopi-input/sys/barcode"
	_ "github.com/djthorpe/gopi-input/sys/input"
	_ "github.com/djthorpe/gopi/sys/logger"
)

var (
	start     = make(chan struct{})
	map_type  = make(map[string]gopi.InputDeviceType)
	map_bus   = make(map[string]gopi.InputDeviceBus)
	keys_type = make([]string, 0)
	keys_bus  = make([]string, 0)
)

///////////////////////////////////////////////////////////////////////////////

func init() {
	// Device types
	for t := gopi.INPUT_TYPE_NONE; t < gopi.INPUT_TYPE_ANY; t++ {
		s := fmt.Sprint(t)
		if strings.HasPrefix(s, "INPUT_TYPE_") {
			k := strings.ToLower(strings.TrimPrefix(s, "INPUT_TYPE_"))
			map_type[k] = t
			keys_type = append(keys_type, k)
		}
	}
	// Device bus
	for b := gopi.INPUT_BUS_NONE; b < gopi.INPUT_BUS_ANY; b++ {
		s := fmt.Sprint(b)
		if strings.HasPrefix(s, "INPUT_BUS_") {
			k := strings.ToLower(strings.TrimPrefix(s, "INPUT_BUS_"))
			map_bus[k] = b
			keys_bus = append(keys_bus, k)
		}
	}
}

///////////////////////////////////////////////////////////////////////////////

func stringForDevice(evt gopi.InputEvent) string {
	device_name := evt.Source().(gopi.InputDevice).Name()
	device_type := strings.ToLower(strings.TrimPrefix(fmt.Sprint(evt.DeviceType()), "INPUT_TYPE_"))
	return fmt.Sprintf("%s [%s]", device_name, device_type)
}

func stringForEvent(evt gopi.InputEvent) string {
	return strings.TrimPrefix(fmt.Sprint(evt.EventType()), "INPUT_EVENT_")
}

func stringForKeyPosition(evt gopi.InputEvent) string {
	if evt.EventType() == gopi.INPUT_EVENT_RELPOSITION {
		return fmt.Sprint(evt.Relative())
	} else if evt.EventType() == gopi.INPUT_EVENT_ABSPOSITION {
		return fmt.Sprint(evt.Position())
	} else {
		return strings.TrimPrefix(fmt.Sprint(evt.Keycode()), "KEYCODE_")
	}
}

///////////////////////////////////////////////////////////////////////////////

func PrintDevicesTable(devices []gopi.InputDevice) {
	// Table
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Type", "Name", "Bus"})
	for _, d := range devices {
		table.Append([]string{
			fmt.Sprint(d.Type()),
			d.Name(),
			fmt.Sprint(d.Bus()),
		})
	}
	table.Render()
}

func PrintInputEvent(evt gopi.InputEvent, once *sync.Once) {
	once.Do(func() {
		fmt.Printf("%-25s %-25s %-15s\n", "DEVICE", "EVENT", "KEY/POSITION")
		fmt.Printf("%-25s %-25s %-15s\n", strings.Repeat("-", 25), strings.Repeat("-", 25), strings.Repeat("-", 15))
	})
	fmt.Printf("%-25s %-25s %-15s\n", stringForDevice(evt), stringForEvent(evt), stringForKeyPosition(evt))
}

func EventLoop(app *gopi.AppInstance, done <-chan struct{}) error {
	var once sync.Once

	// Subscribe to events
	evt_input := app.Input.Subscribe()

FOR_LOOP:
	for {
		select {
		case <-start:
			app.Logger.Info("Start")
		case <-done:
			app.Logger.Info("Done")
			break FOR_LOOP
		case event := <-evt_input:
			if event != nil {
				PrintInputEvent(event.(gopi.InputEvent), &once)
			}
		}
	}

	// Unsubscribe from events
	app.Input.Unsubscribe(evt_input)

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////

func GetFilterType(app *gopi.AppInstance) (gopi.InputDeviceType, error) {
	if device_type, exists := app.AppFlags.GetString("type"); exists == false {
		return gopi.INPUT_TYPE_ANY, nil
	} else {
		filter := gopi.INPUT_TYPE_NONE
		for _, k := range strings.Split(device_type, ",") {
			if v, exists := map_type[k]; exists == false {
				return 0, fmt.Errorf("Invalid type: %v", k)
			} else {
				filter |= v
			}
		}
		return filter, nil
	}
}

func GetFilterBus(app *gopi.AppInstance) (gopi.InputDeviceBus, error) {
	if device_bus, exists := app.AppFlags.GetString("bus"); exists == false {
		return gopi.INPUT_BUS_ANY, nil
	} else {
		filter := gopi.INPUT_BUS_NONE
		for _, k := range strings.Split(device_bus, ",") {
			if v, exists := map_bus[k]; exists == false {
				return 0, fmt.Errorf("Invalid bus: %v", k)
			} else {
				filter |= v
			}
		}
		return filter, nil
	}
}

func GetFilterParameters(app *gopi.AppInstance) (string, gopi.InputDeviceType, gopi.InputDeviceBus, error) {
	device_name, _ := app.AppFlags.GetString("name")

	if device_type, err := GetFilterType(app); err != nil {
		return "", 0, 0, err
	} else if device_bus, err := GetFilterBus(app); err != nil {
		return "", 0, 0, err
	} else {
		return device_name, device_type, device_bus, nil
	}
}

///////////////////////////////////////////////////////////////////////////////

func Main(app *gopi.AppInstance, done chan<- struct{}) error {
	// Type of device
	if device_name, device_type, device_bus, err := GetFilterParameters(app); err != nil {
		done <- gopi.DONE
		return err
	} else if devices, err := app.Input.OpenDevicesByName(device_name, device_type, device_bus); err != nil {
		done <- gopi.DONE
		return err
	} else if len(devices) == 0 {
		done <- gopi.DONE
		return errors.New("No devices opened")
	} else {
		PrintDevicesTable(devices)
	}

	if watch, _ := app.AppFlags.GetBool("watch"); watch {
		// Send start flag
		start <- gopi.DONE

		// Wait for CTRL+C
		fmt.Println("Watching for events, press CTRL+C to end")
		app.WaitForSignal()
		fmt.Println("Terminating")
	}

	done <- gopi.DONE
	return nil
}

func main() {
	config := gopi.NewAppConfig("input", "barcode")
	config.AppFlags.FlagBool("watch", false, "Watch for device events")
	config.AppFlags.FlagString("type", "", fmt.Sprintf("Filter by type of device (%v)", strings.Join(keys_type, ",")))
	config.AppFlags.FlagString("bus", "", fmt.Sprintf("Filter by one or more device busses (%v)", strings.Join(keys_bus, ",")))
	config.AppFlags.FlagString("name", "", fmt.Sprintf("Filter by device name or alias"))
	os.Exit(gopi.CommandLineTool(config, Main, EventLoop))
}
