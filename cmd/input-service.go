package main

import (
	"fmt"
	"os"

	// Frameworks
	"github.com/djthorpe/gopi"
	"github.com/olekukonko/tablewriter"

	// Modules
	_ "github.com/djthorpe/gopi-input/sys/input"
	_ "github.com/djthorpe/gopi-input/sys/keymap"
	_ "github.com/djthorpe/gopi/sys/logger"
)

var (
	start = make(chan struct{})
)

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

func EventLoop(app *gopi.AppInstance, done <-chan struct{}) error {
	// Subscribe to events
	evt_input := app.Input.Subscribe()
	evt_keymap := app.ModuleInstance("keymap").(gopi.Publisher).Subscribe()

FOR_LOOP:
	for {
		select {
		case <-start:
			app.Logger.Info("Start")
		case <-done:
			app.Logger.Info("Done")
			break FOR_LOOP
		case event := <-evt_input:
			app.Logger.Info("Input: %v", event)
		case event := <-evt_keymap:
			app.Logger.Info("Keymap: %v", event)
		}
	}

	// Unsubscribe from events
	app.Input.Unsubscribe(evt_input)
	app.ModuleInstance("keymap").(gopi.Publisher).Unsubscribe(evt_keymap)

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
	config := gopi.NewAppConfig("input", "keymap")
	config.AppFlags.FlagBool("watch", false, "Watch for device events")
	os.Exit(gopi.CommandLineTool(config, Main, EventLoop))
}
