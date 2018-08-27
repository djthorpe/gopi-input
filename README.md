# gopi-input

This respository contains input and keymap modules for
[gopi](http://github.com/djthorpe/gopi)
and some example programs in the `cmd` folder. The input module supports 
keyboards, mice and touchscreens at present, and also allows ypu to add
your own devices (for example, the [remotes](http://github.com/djthorpe/remotes)
repository contains devices which create key presses from IR remote controls).

There are also some examples for remotely accessing input events from devices,
so you could for example run a service taking input from devices in one place
and consume those input events on a different host.

The gopi modules provided by this repository are:

| Platform | Import | Type | Name |
| -------- | ------ | ---- | ---- |
| Linux    | `github.com/djthorpe/gopi-input/sys/input`  | `gopi.MODULE_TYPE_INPUT` | `linux/input` |
| Any      | `github.com/djthorpe/gopi-input/sys/keymap` | `gopi.MODULE_TYPE_KEYMAP` | `sys/keymap` |

The `input` module provides an Input Manager which can be used for discovering input
devices (keyboards, mouse or touchscreen). It publishes events when the input devices
receive events (key presses, releases and cursor moves, for example).

The `keymap` module can receive input events from keyboards and output runes based
on a set of rules. For example, the 'A' key pressed whilst the shift key is pressed will
result in the upper-case 'A' rune event being published, and so forth. You can
create, modify and delete keymap files through this module.

## Input Tester

The first example is the `input-tester` which allows you to view input devices
and events. In order to build on Linux:

```
bash% cd gopi-input && go install cmd/input-tester.go
bash% input-tester -help
Usage of input-tester:
  -bus string
    	Filter by one or more device busses (none,pci,isapnp,usb,hil,bluetooth,virtual,isa,i8042,xtkbd,rs232,gameport,parport,amiga,adb,i2c,host,gsc,atari,spi)
  -debug
    	Set debugging mode
  -input.exclusive
    	Input device exclusivity (default true)
  -log.append
    	When writing log to file, append output to end of file
  -log.file string
    	File for logging (default: log to stderr)
  -name string
    	Filter by device name or alias
  -type string
    	Filter by type of device (none,keyboard,mouse,touchscreen,joystick,remote)
  -verbose
    	Verbose logging
  -watch
    	Watch for device events
```

The `-name`, `-type` and `-bus` flags allow you to chose the devices you want to open. Without
specifying these flags, any device would be chosen. By default, the program displays opened
input devices and quits. By specifying the `-watch` flag the program prints out events generated
until interrupted (CTRL+C):

```
bash$ input-tester -watch
+------------------------+----------------------------+----------------+
|          TYPE          |            NAME            |      BUS       |
+------------------------+----------------------------+----------------+
| INPUT_TYPE_KEYBOARD    | Kano Keyboard              | INPUT_BUS_USB  |
| INPUT_TYPE_MOUSE       | Kano Keyboard              | INPUT_BUS_USB  |
| INPUT_TYPE_KEYBOARD    | USB Adapter USB Device     | INPUT_BUS_USB  |
| INPUT_TYPE_TOUCHSCREEN | FT5406 memory based driver | INPUT_BUS_NONE |
+------------------------+----------------------------+----------------+

Watching for events, press CTRL+C to end

DEVICE                    KEY/POSITION              EVENT           STATE          
------------------------- ------------------------- --------------- ---------------
Kano Keyboard [keyboard]  H                         KEYPRESS        none           
Kano Keyboard [keyboard]  H                         KEYRELEASE      none           
Kano Keyboard [keyboard]  E                         KEYPRESS        none           
Kano Keyboard [keyboard]  E                         KEYRELEASE      none           
Kano Keyboard [keyboard]  L                         KEYPRESS        none           
Kano Keyboard [keyboard]  L                         KEYRELEASE      none           
Kano Keyboard [keyboard]  L                         KEYPRESS        none           
Kano Keyboard [keyboard]  L                         KEYRELEASE      none           
Kano Keyboard [keyboard]  O                         KEYPRESS        none           
Kano Keyboard [keyboard]  O                         KEYRELEASE      none           
Kano Keyboard [keyboard]  W                         KEYPRESS        none           
Kano Keyboard [keyboard]  O                         KEYPRESS        none           
Kano Keyboard [keyboard]  W                         KEYRELEASE      none           
Kano Keyboard [keyboard]  O                         KEYRELEASE      none           
Kano Keyboard [keyboard]  R                         KEYPRESS        none           
Kano Keyboard [keyboard]  R                         KEYRELEASE      none           
Kano Keyboard [keyboard]  L                         KEYPRESS        none           
Kano Keyboard [keyboard]  L                         KEYRELEASE      none           
Kano Keyboard [keyboard]  D                         KEYPRESS        none           
Kano Keyboard [keyboard]  D                         KEYRELEASE      none           
Kano Keyboard [mouse]     BTNLEFT                   KEYPRESS        N/A            
Kano Keyboard [mouse]     BTNLEFT                   KEYRELEASE      N/A            
Kano Keyboard [mouse]     BTNLEFT                   KEYPRESS        N/A            
Kano Keyboard [mouse]     BTNLEFT                   KEYRELEASE      N/A            
FT5406 [touchscreen]      BTNTOUCH                  NONE            N/A            
FT5406 [touchscreen]      gopi.Point{ 362.0,145.0 } ABSPOSITION     N/A            
FT5406 [touchscreen]      BTNTOUCH                  TOUCHRELEASE    N/A            
FT5406 [touchscreen]      gopi.Point{ 362.0,145.0 } ABSPOSITION     N/A       
```

## Input Microservice

The input microservice emits input events to any connected microservice
clients using gRPC. The [protobuf file](https://github.com/djthorpe/gopi-input/blob/master/rpc/protobuf/input/input.proto) defines how a client should interact with the service.

Firstly, install the protoc compiler and the GRPC plugin for golang. On Debian 
Linux (including Raspian Linux) use the following commands:

```
bash% sudo apt install protobuf-compiler
bash% sudo apt install libprotobuf-dev
bash% go get -u github.com/golang/protobuf/protoc-gen-go
```

Then, in order to build the microservice:

```
bash% cd gopi-input && \
  go generate github.com/djthorpe/gopi-input/rpc/protobuf && \
  go install cmd/input-service.go
bash% input-service -help
Usage of input-service:
  -debug
    	Set debugging mode
  -input.bus string
    	Filter by one or more device busses (none,pci,isapnp,usb,hil,bluetooth,virtual,isa,i8042,xtkbd,rs232,gameport,parport,amiga,adb,i2c,host,gsc,atari,spi)
  -input.exclusive
    	Input device exclusivity (default true)
  -input.name string
    	Filter by device name or alias
  -input.type string
    	Filter by type of device (none,keyboard,mouse,touchscreen,joystick,remote)
  -log.append
    	When writing log to file, append output to end of file
  -log.file string
    	File for logging (default: log to stderr)
  -rpc.port uint
    	Server Port
  -rpc.sslcert string
    	SSL Certificate Path
  -rpc.sslkey string
    	SSL Key Path
  -verbose
    	Verbose logging
```

In addition to the `-input.name`,`-input.type` and `-input.bus` arguments as before, the
`-rpc.port` specifies a port to listen for client requests on. The `-rpc.sslcert` and
`-rpc.sslkey` arguments can be used to specify a path for your SSL certificate and key.
If you need to generate these, use the following commands, replacing `$ORG` with your
own organization name:

```
bash% export DAYS=99999 OUT=/var/local/ssl ORG="mutablelogic" HOST=`hostname` && \
  install -d "${OUT}" && \
  openssl req \
    -x509 -nodes \
    -newkey rsa:2048 \
    -keyout "${OUT}/selfsigned.key" \
    -out "${OUT}/selfsigned.crt" \
    -days "${DAYS}" \
    -subj "/O=${ORG}/CN=${HOST}"
```

Then you can run your microservice as follows:

```
bash% input-service \
  -rpc.port 8000 \
  -rpc.sslkey="${OUT}/selfsigned.key" -rpc.sslcert="${OUT}/selfsigned.crt"
```

You will receive a warning that "_Microservice discovery is not enabled, continuing_" but
this simply means that microservice discovery is not enabled. Since you specified a port
you can use this port number in your client when connecting.

## Input Client

There is a client which can connect to the input microservice. You can install it as
follows:

```
bash% cd gopi-input && \
  go generate github.com/djthorpe/gopi-input/rpc/protobuf && \
  go install cmd/input-client.go
bash% input-client -help
Usage of input-client:
  -addr string
    	Gateway address
  -debug
    	Set debugging mode
  -log.append
    	When writing log to file, append output to end of file
  -log.file string
    	File for logging (default: log to stderr)
  -rpc.insecure
    	Disable SSL Connection
  -rpc.service string
    	Comma-separated list of service names
  -rpc.skipverify
    	Skip SSL Verification (default true)
  -rpc.timeout duration
    	Connection timeout
  -verbose
    	Verbose logging
```

The output from running the client is the same as running the `input-tester`
except ypu use the `-addr` flag to determine which microservice to connect to:

```
bash% input-client -addr rpi3plus.local:8000
DEVICE                    KEY/POSITION              EVENT           STATE          
------------------------- ------------------------- --------------- ---------------
keyboard                  H                         KEYPRESS        none           
keyboard                  H                         KEYRELEASE      none           
keyboard                  E                         KEYPRESS        none           
keyboard                  E                         KEYRELEASE      none           
keyboard                  L                         KEYPRESS        none           
keyboard                  L                         KEYRELEASE      none           
keyboard                  L                         KEYPRESS        none           
keyboard                  L                         KEYRELEASE      none           
keyboard                  O                         KEYPRESS        none           
keyboard                  O                         KEYRELEASE      none           
```

Use the `-rpc.insecure` flag on the command line if you don't use SSL
for communication.

## Features and Bugs

I'd appreciate it if you filed feature requests and bugs on [github](https://github.com/djthorpe/gopi-input/issues).
At the moment the following features are in progress:

* Implement the `AddDevice` method for the InputManager
* Deal with the case where devices are added and removed whilst the software
  is running
* Provide events for when devices are added and removed so that they can
  be consumed and more devices opened.
* Implement the keymap module which translates key presses into runes.
* Implement a barcode reading module which validates barcodes and perhaps
  looks up products using an API









