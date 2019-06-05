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
	"path/filepath"
	"sync"

	"github.com/sciter-sdk/go-sciter"
	"github.com/sciter-sdk/go-sciter/window"
	"github.com/spf13/viper"
)

// A sciter window, showing a canvas
type sciterRecorder struct {
	connection connection
	canvas     *canvas

	DiskWriter *canvasDiskWriter

	ClosedMutex sync.RWMutex
	Closed      bool
}

// Opens a new sciter recorder and attaches itself to the given connection and canvas
//
// ONLY CALL FROM MAIN THREAD!
func sciterOpenRecorder(con connection, can *canvas) (closedChan chan struct{}) {
	sre := &sciterRecorder{
		connection: con,
		canvas:     can,
		Closed:     true,
	}

	w, err := window.New(sciter.SW_RESIZEABLE|sciter.SW_TITLEBAR|sciter.SW_CONTROLS|sciter.SW_GLASSY|sciter.SW_ENABLE_DEBUG, sciter.NewRect(50, 300, 400, 500))
	if err != nil {
		log.Fatal(err)
	}

	w.DefineFunction("getRects", func(args ...*sciter.Value) *sciter.Value {
		if len(args) != 0 {
			return sciter.NewValue("Wrong number of parameters")
		}

		rects := []image.Rectangle{}

		viper.UnmarshalKey("recorder."+con.getShortName()+".rects", &rects)

		b, err := json.Marshal(rects)
		if err != nil {
			return sciter.NewValue(fmt.Sprintf("Error marshalling json: %v", err))
		}

		val := sciter.NewValue()
		val.ConvertFromString(string(b), sciter.CVT_JSON_LITERAL)
		return val
	})

	w.DefineFunction("registerRects", func(args ...*sciter.Value) *sciter.Value {
		if len(args) != 1 {
			return sciter.NewValue("Wrong number of parameters")
		}
		jsonRects := args[0] // Clone if value is needed after this function returned
		if !jsonRects.IsObject() {
			return sciter.NewValue("Wrong type of parameters")
		}

		jsonRects.ConvertToString(sciter.CVT_JSON_LITERAL)

		rects := []image.Rectangle{}
		if err := json.Unmarshal([]byte(jsonRects.String()), &rects); err != nil {
			return sciter.NewValue(fmt.Sprintf("Error reading json: %v", err))
		}

		sre.DiskWriter.setListeningRects(rects)
		viper.Set("recorder."+con.getShortName()+".rects", rects)
		log.Debug(viper.WriteConfig())

		return nil
	})

	closedChan = make(chan struct{}) // Signals that the window got closed
	w.DefineFunction("signalClosed", func(args ...*sciter.Value) *sciter.Value {
		if len(args) != 0 {
			return sciter.NewValue("Wrong number of parameters")
		}

		sre.DiskWriter.Close()

		close(closedChan)

		return nil
	})

	path, err := filepath.Abs("ui/recorder.htm")
	if err != nil {
		log.Fatal(err)
	}

	if err := w.LoadFile("file://" + path); err != nil {
		log.Fatal(err)
	}

	cdw, err := can.newCanvasDiskWriter(con.getShortName())
	if err != nil {
		log.Fatal(err)
	}
	sre.DiskWriter = cdw

	rects := []image.Rectangle{}
	viper.UnmarshalKey("recorder."+con.getShortName()+".rects", &rects)
	cdw.setListeningRects(rects)

	w.Show()

	return closedChan
}
