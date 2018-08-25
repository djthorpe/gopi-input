/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2017
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package barcode

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	// Frameworks
	"github.com/djthorpe/gopi"
	"github.com/djthorpe/gopi/util/event"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Barcode manager
type Barcode struct {
	InputManager gopi.InputManager
}

// Driver for barcode interface
type barcode struct {
	// Logger
	log gopi.Logger

	// The input manager
	input gopi.InputManager

	// Receive events done signal
	done chan struct{}

	// Array of keypresses
	keypress []rune

	// Publisher
	event.Publisher
}

type Products struct {
	Code   string
	Total  uint
	Offset uint
	Items  []*Product
}

type Product struct {
	Ean                  string
	Title                string
	Description          string
	Asin                 string
	Brand                string
	Model                string
	Color                string
	Size                 string
	Dimension            string
	Weight               string
	LowestRecordedPrice  float32
	HighestRecordedPrice float32
	Elid                 string
}

/////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

func (config Barcode) Open(log gopi.Logger) (gopi.Driver, error) {
	log.Debug("<barcode>Open{ InputManager=%v }", config.InputManager)

	// Check for required input manager
	if config.InputManager == nil {
		return nil, gopi.ErrBadParameter
	}

	this := new(barcode)
	this.log = log
	this.input = config.InputManager
	this.done = make(chan struct{})
	this.keypress = make([]rune, 0)

	// Subscribe to input manager events
	go this.receiveEvents()

	// Return success
	return this, nil
}

func (this *barcode) Close() error {
	this.log.Debug("<barcoder>Close{ }")

	// Signal done to receiveEvents and wait for
	// channel closure
	this.done <- gopi.DONE
	<-this.done

	// Return success
	return nil
}

func (this *barcode) String() string {
	return fmt.Sprintf("<barcode>{ InputManager=%v  }", this.input)
}

/////////////////////////////////////////////////////////////////////
// RECEIVE EVENTS FROM INPUT MANAGER

func (this *barcode) receiveEvents() {
	evt := this.input.Subscribe()

FOR_LOOP:
	for {
		select {
		case event_ := <-evt:
			// push onto the stack
			if event, ok := event_.(gopi.InputEvent); ok {
				this.push(event)
			}
		case <-this.done:
			break FOR_LOOP
		}
	}

	// Signal end of goroutine
	this.input.Unsubscribe(evt)
	close(this.done)
}

/////////////////////////////////////////////////////////////////////
// PUSH EVENTS ONTO THE STACK

func (this *barcode) push(evt gopi.InputEvent) {
	// Ignore events which aren't from a keyboard and arent KEYPRESS
	if evt.DeviceType() != gopi.INPUT_TYPE_KEYBOARD {
		return
	}
	if evt.EventType() != gopi.INPUT_EVENT_KEYPRESS {
		return
	}

	// UPC is a 12-digit barcode
	switch evt.Keycode() {
	case gopi.KEYCODE_0:
		this.keypress = append(this.keypress, '0')
	case gopi.KEYCODE_1:
		this.keypress = append(this.keypress, '1')
	case gopi.KEYCODE_2:
		this.keypress = append(this.keypress, '2')
	case gopi.KEYCODE_3:
		this.keypress = append(this.keypress, '3')
	case gopi.KEYCODE_4:
		this.keypress = append(this.keypress, '4')
	case gopi.KEYCODE_5:
		this.keypress = append(this.keypress, '5')
	case gopi.KEYCODE_6:
		this.keypress = append(this.keypress, '6')
	case gopi.KEYCODE_7:
		this.keypress = append(this.keypress, '7')
	case gopi.KEYCODE_8:
		this.keypress = append(this.keypress, '8')
	case gopi.KEYCODE_9:
		this.keypress = append(this.keypress, '9')
	case gopi.KEYCODE_ENTER:
		go this.lookup(string(this.keypress))
		// Empty slice but keep capacity
		this.keypress = this.keypress[:0]
	}
}

/////////////////////////////////////////////////////////////////////
// LOOKUP BARCODES

func (this *barcode) lookup(upc string) {
	u, _ := url.Parse("https://api.upcitemdb.com/prod/trial/lookup")
	c := &http.Client{}
	products := &Products{}

	u.RawQuery = url.Values{"upc": {upc}}.Encode()
	if req, err := http.NewRequest("GET", u.String(), nil); err != nil {
		this.log.Error("Lookup: %v", err)
		return
	} else if resp, err := c.Do(req); err != nil {
		this.log.Error("Lookup: %v", err)
		return
	} else {
		defer resp.Body.Close()
		decoder := json.NewDecoder(resp.Body)
		defer resp.Body.Close()
		if err := decoder.Decode(products); err != nil {
			this.log.Error("Lookup: %v", err)
			return
		} else {
			fmt.Println(products)
		}
	}
}

func (p *Products) String() string {
	return fmt.Sprintf("<barcode.products>{ code=%v total=%v offset=%v items=%v }", p.Code, p.Total, p.Offset, p.Items)
}

func (p *Product) String() string {
	return fmt.Sprintf("<barcode.product>{ ean=%v title=%v description=%v asin=%v brand=%v model=%v color=%v size=%v dimension=%v weight=%v lowest_recordedPrice=%v highest_recorded_price=%v elid=%v }", p.Ean, p.Title, p.Description, p.Asin, p.Brand, p.Model, p.Color, p.Size, p.Dimension, p.Weight, p.LowestRecordedPrice, p.HighestRecordedPrice, p.Elid)
}
