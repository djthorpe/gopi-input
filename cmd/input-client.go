/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2018
	All Rights Reserved
	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	input "github.com/djthorpe/gopi-input/rpc/grpc/input"

	// Modules
	grpc "github.com/djthorpe/gopi-rpc/sys/grpc"
	_ "github.com/djthorpe/gopi/sys/logger"
)

var (
	// start signal
	start = make(chan gopi.RPCClientConn)
)

///////////////////////////////////////////////////////////////////////////////

func stringForDevice(evt gopi.InputEvent) string {
	device_type := strings.ToLower(strings.TrimPrefix(fmt.Sprint(evt.DeviceType()), "INPUT_TYPE_"))
	return fmt.Sprintf("%s", device_type)
}

func stringForEvent(evt gopi.InputEvent) string {
	return strings.TrimPrefix(fmt.Sprint(evt.EventType()), "INPUT_EVENT_")
}

func stringForKeyPosition(evt gopi.InputEvent) string {
	if evt.EventType() == gopi.INPUT_EVENT_RELPOSITION {
		return fmt.Sprintf("{%v,%v} => {%v,%v}", evt.Relative().X, evt.Relative().Y, evt.Position().X, evt.Position().Y)
	} else if evt.EventType() == gopi.INPUT_EVENT_ABSPOSITION {
		return fmt.Sprint(evt.Position())
	} else {
		return strings.TrimPrefix(fmt.Sprint(evt.KeyCode()), "KEYCODE_")
	}
}

func stringForDeviceState(evt gopi.InputEvent) string {
	if evt.DeviceType() != gopi.INPUT_TYPE_KEYBOARD {
		return "N/A"
	} else {
		key_state := fmt.Sprint(evt.KeyState())
		return strings.ToLower(strings.Replace(key_state, "KEYSTATE_", "", -1))
	}
}

func PrintInputEvent(evt gopi.InputEvent, once *sync.Once) {
	once.Do(func() {
		fmt.Printf("%-25s %-25s %-15s %-15s\n", "DEVICE", "KEY/POSITION", "EVENT", "STATE")
		fmt.Printf("%-25s %-25s %-15s %-15s\n", strings.Repeat("-", 25), strings.Repeat("-", 25), strings.Repeat("-", 15), strings.Repeat("-", 15))
	})
	fmt.Printf("%-25s %-25s %-15s %-15s\n", stringForDevice(evt), stringForKeyPosition(evt), stringForEvent(evt), stringForDeviceState(evt))
}

////////////////////////////////////////////////////////////////////////////////

func Main(app *gopi.AppInstance, done chan<- struct{}) error {
	// Obtain Client Pool and Gateway address
	pool := app.ModuleInstance("rpc/clientpool").(gopi.RPCClientPool)
	addr, _ := app.AppFlags.GetString("addr")

	// Lookup remote service and run
	ctx, _ := context.WithTimeout(context.Background(), 100*time.Millisecond)
	if records, err := pool.Lookup(ctx, "", addr, 1); err != nil {
		done <- gopi.DONE
		return err
	} else if len(records) == 0 {
		done <- gopi.DONE
		return gopi.ErrDeadlineExceeded
	} else if conn, err := pool.Connect(records[0], 0); err != nil {
		done <- gopi.DONE
		return err
	} else {
		// Send connection
		start <- conn

		// Wait until CTRL+C is pressed
		app.Logger.Info("Waiting for CTRL+C")
		app.WaitForSignal()
		done <- gopi.DONE
	}

	// Success
	return nil
}

func RunLoop(app *gopi.AppInstance, done <-chan struct{}) error {
	var once sync.Once
	pool := app.ModuleInstance("rpc/clientpool").(gopi.RPCClientPool)
	events := make(chan gopi.InputEvent)

	// Obtain the connection
	select {
	case <-done:
		return nil
	case conn := <-start:
		if client_ := pool.NewClient("gopi.Input", conn); client_ == nil {
			return gopi.ErrAppError
		} else if client, ok := client_.(*input.Client); ok == false {
			return gopi.ErrAppError
		} else if err := client.Ping(); err != nil {
			return err
		} else {
			// ListenForInputEvents blocks until done is sent, uses events
			// channel
			go func() {
				if err := client.ListenForInputEvents(done, events); err != nil && grpc.IsErrCanceled(err) == false {
					app.Logger.Error("ListenForInputEvents: %v", err)
				}
				close(events)
			}()
		}
	}

FOR_LOOP:
	for {
		select {
		case evt := <-events:
			if evt == nil {
				break FOR_LOOP
			} else {
				PrintInputEvent(evt, &once)
			}
		}
	}

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func main() {
	// Create the configuration
	config := gopi.NewAppConfig("rpc/client/input")

	// Set the RPCServiceRecord for server discovery
	config.Service = "input"

	// Address for remote service
	config.AppFlags.FlagString("addr", "", "Gateway address")

	// Run the command line tool
	os.Exit(gopi.CommandLineTool(config, Main, RunLoop))
}
