/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2018
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package keymap

import (
	"fmt"
	"os"
	"strings"

	// Frameworks
	"github.com/djthorpe/gopi"
	"github.com/djthorpe/gopi/util/event"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Keymap manager
type KeymapManager struct {
	InputManager gopi.InputManager
	Root, Ext    string
}

// Driver of multiple input devices
type manager struct {
	// Logger
	log gopi.Logger

	// The input manager
	input gopi.InputManager

	// Root path and file extension
	root, ext string

	// Receive events done signal
	done chan struct{}

	// Publisher
	event.Publisher
}

/////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	DEFAULT_EXT = ".keymap"
)

/////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

func (config KeymapManager) Open(log gopi.Logger) (gopi.Driver, error) {
	log.Debug("<keymap.manager>Open{ InputManager=%v root=\"%v\" ext=\"%v\" }", config.InputManager, config.Root, config.ext())

	// Check for required input manager
	if config.InputManager == nil {
		return nil, gopi.ErrBadParameter
	}

	this := new(manager)
	this.log = log
	this.root = config.root()
	this.ext = config.ext()
	this.done = make(chan struct{})

	// Subscribe to input manager events
	//go this.receiveEvents()

	// Return success
	return this, nil
}

func (this *manager) Close() error {
	this.log.Debug("<keymap.manager>Close{ root=\"%v\" }", this.root)

	// Wait for receiveEvents completion
	<-this.done

	// Return success
	return nil
}

func (this *manager) String() string {
	return fmt.Sprintf("<keymap.manager>{ root=\"%v\" }", this.root)
}

/////////////////////////////////////////////////////////////////////
// UTILITY METHODS

func (config KeymapManager) ext() string {
	if config.Ext == "" {
		return DEFAULT_EXT
	} else if strings.HasPrefix(".", config.Ext) {
		return config.Ext
	} else {
		return "." + config.Ext
	}
}

func (config KeymapManager) root() string {
	if config.Root == "" {
		if root, err := os.Getwd(); err != nil {
			return ""
		} else {
			return root
		}
	} else if stat, err := os.Stat(config.Root); os.IsNotExist(err) || stat.IsDir() == false {
		return ""
	} else {
		return config.Root
	}
}

/*
/////////////////////////////////////////////////////////////////////
// RECEIVE EVENTS FROM INPUT MANAGER

func (this *manager) receiveEvents() {
	evt := this.input.Subscribe()

	for {
		select {
		case event := <-evt:
			fmt.Println("GOT %v", event)
		}
	}

	// Signal end of goroutine
	this.input.Unsubscribe(evt)
	done <- gopi.DONE
}
*/
