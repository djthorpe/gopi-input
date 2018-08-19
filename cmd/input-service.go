package main

import (
	"fmt"
	"os"

	// Frameworks
	"github.com/djthorpe/gopi"

	// Modules
	_ "github.com/djthorpe/gopi-input/sys/input"
	_ "github.com/djthorpe/gopi/sys/logger"
)

func Main(app *gopi.AppInstance, done chan<- struct{}) error {
	fmt.Println(app.Input)

	// All inputs
	if devices, err := app.Input.OpenDevicesByName("", gopi.INPUT_TYPE_ANY, gopi.INPUT_BUS_ANY); err != nil {
		return err
	} else {
		for i, d := range devices {
			fmt.Println(i, d)
		}
	}

	done <- gopi.DONE
	return nil
}

func main() {
	config := gopi.NewAppConfig("input")
	os.Exit(gopi.CommandLineTool(config, Main))
}
