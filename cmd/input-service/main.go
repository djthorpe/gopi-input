package main

import (
	"os"

	// Frameworks
	"github.com/djthorpe/gopi"

	// Modules
	_ "github.com/djthorpe/gopi-input/sys/input"
	_ "github.com/djthorpe/gopi-rpc/sys/grpc"
	_ "github.com/djthorpe/gopi/sys/logger"

	// RPC Services
	_ "github.com/djthorpe/gopi-input/rpc/grpc/input"
)

///////////////////////////////////////////////////////////////////////////////

func main() {
	// Create the configuration
	config := gopi.NewAppConfig("rpc/service/input")

	// Set the RPCServiceRecord for server discovery
	config.Service = "input"

	// Run the server and register all the services
	os.Exit(gopi.RPCServerTool(config))
}
