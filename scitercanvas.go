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
	"log"
	"path/filepath"

	"github.com/sciter-sdk/go-sciter"
	"github.com/sciter-sdk/go-sciter/window"
)

// A sciter window, showing a canvas
type sciterCanvas struct {
	connection connection
	canvas     *canvas

	object    *sciter.Value
	cbHandler *sciter.Value // Event handler callback of the window
}

func sciterOpenCanvas(con connection, can *canvas) {
	sca := &sciterCanvas{
		connection: con,
		canvas:     can,
	}

	sciter.SetOption(sciter.SCITER_SET_DEBUG_MODE, 1)
	sciter.SetOption(sciter.SCITER_SET_SCRIPT_RUNTIME_FEATURES, sciter.ALLOW_FILE_IO|sciter.ALLOW_SOCKET_IO|sciter.ALLOW_EVAL|sciter.ALLOW_SYSINFO) // Needed for the inspector to work!

	w, err := window.New(sciter.SW_MAIN|sciter.SW_RESIZEABLE|sciter.SW_TITLEBAR|sciter.SW_CONTROLS|sciter.SW_ENABLE_DEBUG /*|sciter.SW_GLASSY*/, sciter.DefaultRect)
	if err != nil {
		log.Fatal(err)
	}

	path, err := filepath.Abs("ui/canvas.htm")
	if err != nil {
		log.Fatal(err)
	}

	if err := w.LoadFile("file://" + path); err != nil {
		log.Fatal(err)
	}

	ok := w.SetOption(sciter.SCITER_SET_DEBUG_MODE, 1)
	if !ok {
		log.Println("set debug mode failed")
	}

	w.DefineFunction("setEventHandler", func(args ...*sciter.Value) *sciter.Value {
		if len(args) != 2 {
			return sciter.NewValue("Wrong number of parameters")
		}
		obj, cbHandler := args[0].Clone(), args[1].Clone() // Always clone, otherwise those are just references to sciter values and will be invalid if used after return
		if !obj.IsObject() || !cbHandler.IsObjectFunction() {
			return sciter.NewValue("Wrong type of parameters")
		}

		if sca.cbHandler != nil {
			return sciter.NewValue("Callback already set")
		}

		sca.object = obj
		sca.cbHandler = cbHandler
		can.subscribeListener(sca)

		return nil
	})

	w.DefineFunction("queryRect", func(args ...*sciter.Value) *sciter.Value {
		if len(args) != 1 {
			return sciter.NewValue("Wrong number of parameters")
		}
		jsonRect := args[0].Clone() // Always clone, otherwise those are just references to sciter values and will be invalid if used after return
		if !jsonRect.IsObject() {
			return sciter.NewValue("Wrong type of parameters")
		}

		jsonRect.ConvertToString(sciter.CVT_JSON_LITERAL)

		rect := image.Rectangle{}
		if err := json.Unmarshal([]byte(jsonRect.String()), &rect); err != nil {
			return sciter.NewValue(fmt.Sprintf("Error reading json: %v", err))
		}

		can.queryRect(rect)

		return nil
	})

	w.Show()
	w.Run()
}

func (s *sciterCanvas) handleInvalidateAll() error {
	jsonData := struct {
		Type string
	}{
		"InvalidateAll",
	}

	b, err := json.Marshal(jsonData)
	if err == nil {
		val := sciter.NullValue()
		val.ConvertFromString(string(b), sciter.CVT_JSON_LITERAL)
		s.cbHandler.Invoke(s.object, "[Native Script]", val)
	}

	return nil
}

func (s *sciterCanvas) handleInvalidateRect(rect image.Rectangle) error {
	jsonData := struct {
		Type string
		Rect image.Rectangle
	}{
		"InvalidateRect",
		rect,
	}

	b, err := json.Marshal(jsonData)
	if err == nil {
		val := sciter.NullValue()
		val.ConvertFromString(string(b), sciter.CVT_JSON_LITERAL)
		s.cbHandler.Invoke(s.object, "[Native Script]", val)
	}

	return nil
}

func (s *sciterCanvas) handleSetImage(img *image.Paletted) error {
	jsonData := struct {
		Type string
		Img  *image.Paletted // TODO: Forward RGBA image, instead of paletted image
	}{
		"SetImage",
		img,
	}

	b, err := json.Marshal(jsonData)
	if err == nil {
		val := sciter.NullValue()
		val.ConvertFromString(string(b), sciter.CVT_JSON_LITERAL)
		s.cbHandler.Invoke(s.object, "[Native Script]", val)
	}

	return nil
}

func (s *sciterCanvas) handleSetPixel(pos image.Point, colorIndex uint8) error {
	jsonData := struct {
		Type       string
		Pos        image.Point
		ColorIndex uint8 // TODO: Forward RGBA color, instead of index
	}{
		"SetPixel",
		pos,
		colorIndex,
	}

	b, err := json.Marshal(jsonData)
	if err == nil {
		val := sciter.NullValue()
		val.ConvertFromString(string(b), sciter.CVT_JSON_LITERAL)
		s.cbHandler.Invoke(s.object, "[Native Script]", val)
	}

	return nil
}

func (s *sciterCanvas) handleSignalDownload(rect image.Rectangle) error {
	jsonData := struct {
		Type string
		Rect image.Rectangle
	}{
		"SignalDownload",
		rect,
	}

	b, err := json.Marshal(jsonData)
	if err == nil {
		val := sciter.NullValue()
		val.ConvertFromString(string(b), sciter.CVT_JSON_LITERAL)
		s.cbHandler.Invoke(s.object, "[Native Script]", val)
	}

	return nil
}
