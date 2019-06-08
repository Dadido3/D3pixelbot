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
	"image"
	"math/rand"
	"testing"
)

func Test_canvas_newCanvasDiskWriter(t *testing.T) {
	can, _ := newCanvas(pixelSize{64, 64}, image.Point{}, pixelcanvasioCanvasRect)

	cdw, err := can.newCanvasDiskWriter("Test")
	if err != nil {
		t.Errorf("Can't create canvas disk writer: %v", err)
	}

	can.subscribeListener(cdw, false) // Don't let the canvas manage virtual chunks for us

	for i := 0; i < 128; i++ {
		rect := image.Rectangle{image.Point{i * 64, i * 64}, image.Point{i*64 + 64, i*64 + 64}}
		if _, err := can.signalDownload(rect); err != nil {
			t.Errorf("Can't signal download at rectangle %v: %v", rect, err)
		}
	}

	for i := 0; i < 128; i++ {
		pos := image.Point{i, i}
		if err := can.setPixel(pos, pixelcanvasioPalette[rand.Intn(len(pixelcanvasioPalette))]); err != nil {
			t.Errorf("Can't set pixel at %v: %v", pos, err)
		}
	}

	for i := 0; i < 128; i++ {
		rect := image.Rectangle{image.Point{i * 64, i * 64}, image.Point{i*64 + 64, i*64 + 64}}
		if err := can.setImage(rect, false, false); err != nil {
			t.Errorf("Can't set image at %v: %v", rect, err)
		}
	}

	cdw.Close()

	can.Close()
}
