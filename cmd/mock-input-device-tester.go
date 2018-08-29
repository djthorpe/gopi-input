package main

import (
	"fmt"
	"os"

	// Frameworks
	gopi "github.com/djthorpe/gopi"

	// Modules
	_ "github.com/djthorpe/gopi-input/sys/input"
	_ "github.com/djthorpe/gopi-input/sys/input-device-helloworld"
	_ "github.com/djthorpe/gopi/sys/logger"
)

///////////////////////////////////////////////////////////////////////////////

func EventLoop(app *gopi.AppInstance, done <-chan struct{}) error {
	// Subscribe to events
	evt_input := app.Input.Subscribe()
FOR_LOOP:
	for {
		select {
		case <-done:
			break FOR_LOOP
		case event := <-evt_input:
			fmt.Println(event)
		}
	}

	// Unsubscribe from events
	app.Input.Unsubscribe(evt_input)

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////

func Main(app *gopi.AppInstance, done chan<- struct{}) error {
	// Wait for CTRL+C
	fmt.Println("Watching for events, press CTRL+C to end")
	app.WaitForSignal()
	fmt.Println("Terminating")
	done <- gopi.DONE
	return nil
}

func main() {
	config := gopi.NewAppConfig("input/device/helloworld")
	os.Exit(gopi.CommandLineTool(config, Main, EventLoop))
}
