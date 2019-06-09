/*  D3pixelbot - Custom client, recorder and bot for pixel drawing games
    Copyright (C) 2019  David Vogel

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.  */

package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"path/filepath"
	"sync"

	"github.com/sciter-sdk/go-sciter"
	"github.com/sciter-sdk/go-sciter/window"
)

// A sciter window, showing a canvas
type sciterCanvas struct {
	connection connection
	canvas     *canvas

	handlerChan chan *sciter.Value // Queue of event data, so the main logic doesn't stop while sciter is processing it
	ClosedMutex sync.RWMutex
	Closed      bool
}

// Opens a new sciter canvas and attaches itself to the given connection and canvas
//
// ONLY CALL FROM MAIN THREAD!
func sciterOpenCanvas(con connection, can *canvas) (closedChan chan struct{}) {
	sca := &sciterCanvas{
		connection: con,
		canvas:     can,
		Closed:     true,
	}

	w, err := window.New(sciter.SW_RESIZEABLE|sciter.SW_TITLEBAR|sciter.SW_CONTROLS|sciter.SW_GLASSY|sciter.SW_ENABLE_DEBUG, sciter.NewRect(50, 300, 800, 500))
	if err != nil {
		log.Fatal(err)
	}

	w.DefineFunction("subscribeCanvasEvents", func(args ...*sciter.Value) *sciter.Value {
		if len(args) != 2 {
			return sciter.NewValue("Wrong number of parameters")
		}
		//obj, cbHandler := args[0].Clone(), args[1].Clone() // Always clone, otherwise those are just references to sciter values and will be invalid if used after return
		//if !obj.IsObject() || !cbHandler.IsObjectFunction() {
		//	return sciter.NewValue("Wrong type of parameters")
		//}

		sca.ClosedMutex.Lock()
		defer sca.ClosedMutex.Unlock()

		if sca.handlerChan != nil {
			return sciter.NewValue("Already subscribed")
		}

		err := can.subscribeListener(sca, true) // Let the canvas manage virtual chunks for us
		if err != nil {
			return sciter.NewValue("Can't subscribe to canvas: " + err.Error())
		}

		sca.handlerChan = make(chan *sciter.Value, 300) // Can be after can.subscribeListener, as the ClosedMutex is still locked here
		sca.Closed = false

		go func(channel <-chan *sciter.Value) {
			for {
				val, ok := <-channel
				if !ok {
					// Channel closed, so just close goroutine
					return
				}
				w.Call("processEvents", val)
				val.Release()
			}
		}(sca.handlerChan)

		return nil
	})

	w.DefineFunction("unsubscribeCanvasEvents", func(args ...*sciter.Value) *sciter.Value {
		if len(args) != 0 {
			return sciter.NewValue("Wrong number of parameters")
		}

		sca.ClosedMutex.Lock()
		defer sca.ClosedMutex.Unlock()

		if sca.handlerChan == nil {
			return sciter.NewValue("Not subscribed")
		}

		err := can.unsubscribeListener(sca)
		if err != nil {
			return sciter.NewValue("Can't subscribe to canvas: " + err.Error())
		}

		close(sca.handlerChan)
		sca.handlerChan = nil // Goroutine has its own reference to this channel
		sca.Closed = true

		return nil
	})

	rectsChan := make(chan []image.Rectangle, 1)
	go func() {
		for rects := range rectsChan {
			can.registerRects(sca, rects)
		}
	}()

	w.DefineFunction("registerRects", func(args ...*sciter.Value) *sciter.Value {
		if len(args) != 1 {
			return sciter.NewValue("Wrong number of parameters")
		}
		jsonRect := args[0] // Clone if value is needed after this function returned
		if !jsonRect.IsObject() {
			return sciter.NewValue("Wrong type of parameters")
		}

		//log.Tracef("registerRects #2")

		jsonRect.ConvertToString(sciter.CVT_JSON_LITERAL)

		rects := []image.Rectangle{}
		if err := json.Unmarshal([]byte(jsonRect.String()), &rects); err != nil {
			return sciter.NewValue(fmt.Sprintf("Error reading json: %v", err))
		}

		//log.Tracef("registerRects #3")

		// Write rect into channel, or replace the current one if the goroutine is busy
		select {
		case rectsChan <- rects:
		default:
			select {
			case <-rectsChan:
			default:
			}
			rectsChan <- rects
		}

		//log.Tracef("registerRects #4")

		return nil
	})

	w.DefineFunction("trace", func(args ...*sciter.Value) *sciter.Value {
		if len(args) != 1 {
			return sciter.NewValue("Wrong number of parameters")
		}
		message := args[0].String()

		log.Trace(message)

		return nil
	})

	closedChan = make(chan struct{}) // Signals that the window got closed
	w.DefineFunction("signalClosed", func(args ...*sciter.Value) *sciter.Value {
		if len(args) != 0 {
			return sciter.NewValue("Wrong number of parameters")
		}

		close(rectsChan)
		close(closedChan)

		return nil
	})

	path, err := filepath.Abs("ui/canvas.htm")
	if err != nil {
		log.Fatal(err)
	}

	if err := w.LoadFile("file://" + path); err != nil {
		log.Fatal(err)
	}

	// Testing pixel events
	/*go func() {
		for {
			can.setPixel(image.Point{rand.Intn(128), rand.Intn(128)}, color.RGBA{uint8(rand.Intn(256)), uint8(rand.Intn(256)), uint8(rand.Intn(256)), uint8(rand.Intn(256))})
			time.Sleep(10 * time.Millisecond)
		}
	}()*/

	w.Show()

	return closedChan
}

func (s *sciterCanvas) handleInvalidateAll() error {
	s.ClosedMutex.RLock()
	defer s.ClosedMutex.RUnlock()
	if s.Closed {
		return fmt.Errorf("Listener is closed")
	}

	//val := sciter.NewValue(1)
	//s.handlerChan <- val

	return nil
}

func (s *sciterCanvas) handleInvalidateRect(rect image.Rectangle, vcIDs []int) error {
	s.ClosedMutex.RLock()
	defer s.ClosedMutex.RUnlock()
	if s.Closed {
		return fmt.Errorf("Listener is closed")
	}

	//val := sciter.NewValue(2)
	//s.handlerChan <- val

	return nil
}

func (s *sciterCanvas) handleRevalidateRect(rect image.Rectangle, vcIDs []int) error {
	s.ClosedMutex.RLock()
	defer s.ClosedMutex.RUnlock()
	if s.Closed {
		return fmt.Errorf("Listener is closed")
	}

	//val := sciter.NewValue(3)
	//s.handlerChan <- val

	return nil
}

func (s *sciterCanvas) handleSetImage(img image.Image, valid bool, vcIDs []int) error {
	s.ClosedMutex.RLock()
	defer s.ClosedMutex.RUnlock()
	if s.Closed {
		return fmt.Errorf("Listener is closed")
	}

	//val := sciter.NewValue(4)
	//s.handlerChan <- val

	return nil
}

func (s *sciterCanvas) handleSetPixel(pos image.Point, color color.Color, vcID int) error {
	s.ClosedMutex.RLock()
	defer s.ClosedMutex.RUnlock()
	if s.Closed {
		return fmt.Errorf("Listener is closed")
	}

	//val := sciter.NewValue(5)
	//s.handlerChan <- val

	return nil
}

func (s *sciterCanvas) handleSignalDownload(rect image.Rectangle, vcIDs []int) error {
	s.ClosedMutex.RLock()
	defer s.ClosedMutex.RUnlock()
	if s.Closed {
		return fmt.Errorf("Listener is closed")
	}

	//val := sciter.NewValue(6)
	//s.handlerChan <- val

	return nil
}

func (s *sciterCanvas) handleChunksChange(create, remove map[image.Rectangle]int) error {
	s.ClosedMutex.RLock()
	defer s.ClosedMutex.RUnlock()
	if s.Closed {
		return fmt.Errorf("Listener is closed")
	}

	//val := sciter.NewValue(7)
	//s.handlerChan <- val

	return nil
}
