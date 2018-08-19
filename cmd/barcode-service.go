package main

import (
	"fmt"
	"os"

	// Frameworks
	"github.com/djthorpe/gopi"

	// Modules
	_ "github.com/djthorpe/barcode/sys/input"
)

func Main(app *gopi.AppInstance, done chan<- struct{}) error {
	fmt.Println(app.Input)
}

func main() {
	config := gopi.NewAppConfig("input")
	os.Exit(gopi.CommandLineTool(config, Main))
}
