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
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/nfnt/resize"
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

	w, err := window.New(sciter.SW_RESIZEABLE|sciter.SW_TITLEBAR|sciter.SW_CONTROLS|sciter.SW_GLASSY|sciter.SW_ENABLE_DEBUG, sciter.NewRect(50, 300, 800, 800))
	if err != nil {
		log.Fatal(err)
	}

	w.DefineFunction("subscribeCanvasEvents", func(args ...*sciter.Value) *sciter.Value {
		if len(args) != 2 {
			log.Errorf("Wrong number of parameters")
			return sciter.NewValue("Wrong number of parameters")
		}
		obj, cbHandler := args[0].Clone(), args[1].Clone() // Always clone, otherwise those are just references to sciter values and will be invalid if used after return
		if !obj.IsObject() || !cbHandler.IsObjectFunction() {
			log.Errorf("Wrong type of parameters")
			return sciter.NewValue("Wrong type of parameters")
		}

		sca.ClosedMutex.Lock()
		defer sca.ClosedMutex.Unlock()

		if sca.handlerChan != nil {
			log.Errorf("Already subscribed")
			return sciter.NewValue("Already subscribed")
		}

		err := can.subscribeListener(sca, true) // Let the canvas manage virtual chunks for us
		if err != nil {
			log.Errorf("Can't subscribe to canvas: %v", err)
			return sciter.NewValue(fmt.Sprintf("Can't subscribe to canvas: %v", err))
		}

		sca.handlerChan = make(chan *sciter.Value, 300) // Can be after can.subscribeListener, as the ClosedMutex is still locked here
		sca.Closed = false

		go func(channel <-chan *sciter.Value) {
			for {
				// Batch read from channel, or return if the channel got closed
				events := []*sciter.Value{}
				event, ok := <-channel
				if !ok {
					// Channel closed, so just close goroutine
					return
				}
				events = append(events, event)
			batchLoop:
				for i := 1; i < 50; i++ { // Limit batch size to 50
					select {
					case event, ok := <-channel:
						if ok {
							events = append(events, event)
						}
					default:
						break batchLoop
					}
				}

				val := sciter.NewValue()
				for _, event := range events {
					val.Append(event)
					event.Release()
				}
				//log.Tracef("Invoke cbHandler with %v", val)
				cbHandler.Invoke(obj, "[Native Script]", val)
				//log.Tracef("Invoke cbHandler with %v done", val)
				val.Release()
				//log.Tracef("val released")
			}
		}(sca.handlerChan)

		return nil
	})

	w.DefineFunction("unsubscribeCanvasEvents", func(args ...*sciter.Value) *sciter.Value {
		if len(args) != 0 {
			log.Errorf("Wrong number of parameters")
			return sciter.NewValue("Wrong number of parameters")
		}

		sca.ClosedMutex.Lock()
		defer sca.ClosedMutex.Unlock()

		if sca.handlerChan == nil {
			log.Errorf("Not subscribed")
			return sciter.NewValue("Not subscribed")
		}

		err := can.unsubscribeListener(sca)
		if err != nil {
			log.Errorf("Can't unsubscribe to canvas: %v", err)
			return sciter.NewValue(fmt.Sprintf("Can't unsubscribe to canvas: %v", err))
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
			log.Errorf("Wrong number of parameters")
			return sciter.NewValue("Wrong number of parameters")
		}
		jsonRect := args[0] // Clone if value is needed after this function returned
		if !jsonRect.IsObject() {
			log.Errorf("Wrong type of parameters")
			return sciter.NewValue("Wrong type of parameters")
		}

		jsonRect.ConvertToString(sciter.CVT_JSON_LITERAL)

		rects := []image.Rectangle{}
		if err := json.Unmarshal([]byte(jsonRect.String()), &rects); err != nil {
			log.Errorf("Error reading json: %v", err)
			return sciter.NewValue(fmt.Sprintf("Error reading json: %v", err))
		}

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

		return nil
	})

	w.DefineFunction("setReplayTime", func(args ...*sciter.Value) *sciter.Value {
		if len(args) != 1 {
			log.Errorf("Wrong number of parameters")
			return sciter.NewValue("Wrong number of parameters")
		}
		jsonRect := args[0] // Clone if value is needed after this function returned
		if !jsonRect.IsDate() {
			log.Errorf("Wrong type of parameters")
			return sciter.NewValue("Wrong type of parameters")
		}

		conR, ok := con.(connectionReplay) // Check if connection has replay time methods
		if !ok {
			log.Errorf("Can't set replay time on %T", con)
			return sciter.NewValue(fmt.Sprintf("Can't set replay time on %T", con))
		}

		ft := &sciterTime{
			LowDateTime:  uint32(jsonRect.Int64()),
			HighDateTime: uint32(jsonRect.Int64() >> 32),
		}

		t := time.Unix(0, ft.Nanoseconds())

		err := conR.setReplayTime(t)
		if err != nil {
			log.Errorf("Can't set replay time %T", err)
			return sciter.NewValue(err.Error())
		}

		return nil
	})

	// TODO: Hide UI if connection doesn't have the replayTime method
	w.DefineFunction("hasReplayTime", func(args ...*sciter.Value) (val *sciter.Value) {
		val = sciter.NewValue()

		if len(args) != 0 {
			log.Errorf("Wrong number of parameters")
			val.Set("Error", "Wrong number of parameters")
			return
		}

		conRep, ok := con.(connectionReplay) // Check if connection has replay time methods
		if !ok {
			log.Errorf("%T doesn't support setReplayTime", con)
			val.Set("Error", fmt.Sprintf("%T doesn't support setReplayTime", con))
			return
		}

		recs := conRep.getRecordings()

		val.Set("HasReplayTime", sciter.NewValue(true))
		if len(recs) >= 1 {
			val.Set("StartTime", recs[0].StartTime.Format(time.RFC3339Nano))
		}

		return val
	})

	w.DefineFunction("saveImage", func(args ...*sciter.Value) *sciter.Value {
		if len(args) != 4 {
			log.Errorf("Wrong number of parameters")
			return sciter.NewValue("Wrong number of parameters")
		}
		sciterRect, sciterSize, sciterPath, cbHandler := args[0], args[1], args[2], args[3].Clone() // Clone if value is needed after this function has returned
		if !sciterRect.IsObject() || !sciterSize.IsObject() || !sciterPath.IsString() || !cbHandler.IsObjectFunction() {
			log.Errorf("Wrong type of parameters")
			return sciter.NewValue("Wrong type of parameters")
		}

		min, max := sciterRect.Get("Min"), sciterRect.Get("Max")

		rect := image.Rectangle{
			image.Point{int(int32(min.Get("X").Int())), int(int32(min.Get("Y").Int()))},
			image.Point{int(int32(max.Get("X").Int())), int(int32(max.Get("Y").Int()))},
		}

		size := pixelSize{sciterSize.Get("X").Int(), sciterSize.Get("Y").Int()}

		filename := sciterPath.String()

		log.Tracef("Starting to save image %v at %v with size of %v", filename, rect, size)

		file, err := os.Create(filename)
		if err != nil {
			log.Errorf("Can't create file %v: %v", filename, err)
			return sciter.NewValue(fmt.Sprintf("Can't create file %v: %v", filename, err))
		}

		go func() {
			defer file.Close()

			img, err := can.getImageCopy(rect, false, true)
			if err != nil {
				log.Errorf("Can't get image at %v: %v", rect, err)
				return
			}
			resized := resize.Resize(uint(size.X), uint(size.Y), img, resize.Lanczos3)
			png.Encode(file, resized)

			log.Tracef("Finished to save image %v", filename)

			cbHandler.Invoke(sciter.NewValue(), "[Native Script]")
		}()

		return nil
	})

	closedChan = make(chan struct{}) // Signals that the window got closed
	w.DefineFunction("signalClosed", func(args ...*sciter.Value) *sciter.Value {
		if len(args) != 0 {
			log.Errorf("Wrong number of parameters")
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

	val := sciter.NewValue()
	val.Set("Type", "InvalidateAll")

	s.handlerChan <- val

	return nil
}

func (s *sciterCanvas) handleInvalidateRect(rect image.Rectangle, vcIDs []int) error {
	s.ClosedMutex.RLock()
	defer s.ClosedMutex.RUnlock()
	if s.Closed {
		return fmt.Errorf("Listener is closed")
	}

	val := sciter.NewValue()
	val.Set("Type", "InvalidateRect")
	val.Set("X", rect.Min.X)
	val.Set("Y", rect.Min.Y)
	val.Set("Width", rect.Dx())
	val.Set("Height", rect.Dy())
	valArray := sciter.NewValue()
	defer valArray.Release()
	for k, v := range vcIDs {
		valArray.SetIndex(k, v)
	}
	val.Set("VcIDs", valArray)

	s.handlerChan <- val

	return nil
}

func (s *sciterCanvas) handleRevalidateRect(rect image.Rectangle, vcIDs []int) error {
	s.ClosedMutex.RLock()
	defer s.ClosedMutex.RUnlock()
	if s.Closed {
		return fmt.Errorf("Listener is closed")
	}

	val := sciter.NewValue()
	val.Set("Type", "RevalidateRect")
	val.Set("X", rect.Min.X)
	val.Set("Y", rect.Min.Y)
	val.Set("Width", rect.Dx())
	val.Set("Height", rect.Dy())
	valArray := sciter.NewValue()
	defer valArray.Release()
	for k, v := range vcIDs {
		valArray.SetIndex(k, v)
	}
	val.Set("VcIDs", valArray)

	s.handlerChan <- val

	return nil
}

func (s *sciterCanvas) handleSetImage(img image.Image, valid bool, vcIDs []int) error {
	s.ClosedMutex.RLock()
	defer s.ClosedMutex.RUnlock()
	if s.Closed {
		return fmt.Errorf("Listener is closed")
	}

	imageArray := imageToBGRAArray(img)
	headerArray := [12]byte{'B', 'G', 'R', 'A'}
	binary.BigEndian.PutUint32(headerArray[4:8], uint32(img.Bounds().Dx()))
	binary.BigEndian.PutUint32(headerArray[8:12], uint32(img.Bounds().Dy()))
	array := append(headerArray[:], imageArray...)

	val := sciter.NewValue()
	val.Set("Type", "SetImage")
	val.Set("X", img.Bounds().Min.X)
	val.Set("Y", img.Bounds().Min.Y)
	val.Set("Width", img.Bounds().Dx())
	val.Set("Height", img.Bounds().Dy())
	valArray := sciter.NewValue()
	defer valArray.Release()
	valArray.SetBytes(array)
	val.Set("Array", valArray)
	val.Set("Valid", valid)
	valArray = sciter.NewValue()
	defer valArray.Release()
	for k, v := range vcIDs {
		valArray.SetIndex(k, v)
	}
	val.Set("VcIDs", valArray)

	s.handlerChan <- val

	return nil
}

func (s *sciterCanvas) handleSetPixel(pos image.Point, color color.Color, vcID int) error {
	s.ClosedMutex.RLock()
	defer s.ClosedMutex.RUnlock()
	if s.Closed {
		return fmt.Errorf("Listener is closed")
	}

	r, g, b, a := color.RGBA()

	val := sciter.NewValue()
	val.Set("Type", "SetPixel")
	val.Set("X", pos.X)
	val.Set("Y", pos.Y)
	val.Set("R", r>>8)
	val.Set("G", g>>8)
	val.Set("B", b>>8)
	val.Set("A", a>>8)
	val.Set("VcID", vcID)

	s.handlerChan <- val

	return nil
}

func (s *sciterCanvas) handleSignalDownload(rect image.Rectangle, vcIDs []int) error {
	s.ClosedMutex.RLock()
	defer s.ClosedMutex.RUnlock()
	if s.Closed {
		return fmt.Errorf("Listener is closed")
	}

	val := sciter.NewValue()
	val.Set("Type", "SignalDownload")
	val.Set("X", rect.Min.X)
	val.Set("Y", rect.Min.Y)
	val.Set("Width", rect.Dx())
	val.Set("Height", rect.Dy())
	valArray := sciter.NewValue()
	defer valArray.Release()
	for k, v := range vcIDs {
		valArray.SetIndex(k, v)
	}
	val.Set("VcIDs", valArray)

	s.handlerChan <- val

	return nil
}

func (s *sciterCanvas) handleChunksChange(create, remove map[image.Rectangle]int) error {
	s.ClosedMutex.RLock()
	defer s.ClosedMutex.RUnlock()
	if s.Closed {
		return fmt.Errorf("Listener is closed")
	}

	removeIDs := []int{}
	for _, id := range remove {
		removeIDs = append(removeIDs, id)
	}

	createIDs := []struct {
		Rect image.Rectangle
		VcID int
	}{}
	for rect, id := range create {
		createIDs = append(createIDs, struct {
			Rect image.Rectangle
			VcID int
		}{
			rect, id,
		})
	}

	jsonData := struct {
		Type   string
		Create []struct {
			Rect image.Rectangle
			VcID int
		}
		Remove []int
	}{
		"ChunksChange",
		createIDs, removeIDs,
	}

	// TODO: Don't use json as intermediary

	b, err := json.Marshal(jsonData)
	if err != nil {
		return fmt.Errorf("Can't convert to JSON object: %v", err)
	}

	val := sciter.NewValue()
	val.ConvertFromString(string(b), sciter.CVT_JSON_LITERAL)
	s.handlerChan <- val

	return nil
}

func (s *sciterCanvas) handleSetTime(t time.Time) error {
	s.ClosedMutex.RLock()
	defer s.ClosedMutex.RUnlock()
	if s.Closed {
		return fmt.Errorf("Listener is closed")
	}

	val := sciter.NewValue()
	val.Set("Type", "SetTime")
	val.Set("RFC3339Nano", t.Format(time.RFC3339Nano))

	s.handlerChan <- val

	return nil
}
