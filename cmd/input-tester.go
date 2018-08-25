package main

import (
	"fmt"
	"os"
	"strings"
	"sync"

	// Frameworks
	"github.com/djthorpe/gopi"
	"github.com/olekukonko/tablewriter"

	// Modules
	_ "github.com/djthorpe/gopi-input/sys/input"
	_ "github.com/djthorpe/gopi/sys/logger"
)

var (
	start = make(chan struct{})
)

func stringForDevice(evt gopi.InputEvent) string {
	return fmt.Sprintf("%s - %s", evt.Source().(gopi.InputDevice).Name(), strings.TrimPrefix(fmt.Sprint(evt.DeviceType()), "INPUT_TYPE_"))
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

func Main(app *gopi.AppInstance, done chan<- struct{}) error {
	// Open devices
	if devices, err := app.Input.OpenDevicesByName("", gopi.INPUT_TYPE_ANY, gopi.INPUT_BUS_ANY); err != nil {
		return err
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
	config := gopi.NewAppConfig("input")
	config.AppFlags.FlagBool("watch", false, "Watch for device events")
	os.Exit(gopi.CommandLineTool(config, Main, EventLoop))
}
