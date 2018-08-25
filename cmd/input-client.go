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
	"os"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"

	// Modules
	input "github.com/djthorpe/gopi-input/rpc/grpc/input"
	_ "github.com/djthorpe/gopi/sys/logger"
	_ "github.com/djthorpe/gopi/sys/rpc/grpc"
	_ "github.com/djthorpe/gopi/sys/rpc/mdns"
)

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
		// Send connection on channel
		runconn <- conn
		// Wait until CTRL+C is pressed
		app.Logger.Info("Waiting for CTRL+C")
		app.WaitForSignal()
		done <- gopi.DONE
		if err := pool.Disconnect(conn); err != nil {
			return err
		}
	}

	// Success
	return nil
}

func RunLoop(app *gopi.AppInstance, done <-chan struct{}) error {
	pool := app.ModuleInstance("rpc/clientpool").(gopi.RPCClientPool)
	done := make(chan struct{})

	if client_ := pool.NewClient("gopi.Input", conn); client_ == nil {
		return gopi.ErrAppError
	} else if client, ok := client_.(*input.Client); ok == false {
		return gopi.ErrAppError
	} else if err := client.Ping(); err != nil {
		return err
	} else if err := client.ListenForInputEvents(done); err != nil { // method blocks until 'done' is sent
		return err
	} else {
		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////

func main() {
	// Create the configuration
	config := gopi.NewAppConfig("rpc/client/input:grpc")

	// Set the RPCServiceRecord for server discovery
	config.Service = "input"

	// Address for remote service
	config.AppFlags.FlagString("addr", "", "Gateway address")

	// Run the command line tool
	os.Exit(gopi.CommandLineTool(config, Main, RunLoop))
}
