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
	"testing"
)

func Test_canvas_newCanvasDiskWriter(t *testing.T) {
	can := &canvas{
		Palette:   pixelcanvasioPalette,
		ChunkSize: pixelSize{64, 64},
	}

	cdw, err := can.newCanvasDiskWriter("Test")
	if err != nil {
		t.Errorf("Can't create canvas disk writer: %v", err)
	}

	for i := 0; i < 128; i++ {
		rect := image.Rectangle{image.Point{i * 64, i * 64}, image.Point{i*64 + 64, i*64 + 64}}
		if err := cdw.handleSignalDownload(rect); err != nil {
			t.Errorf("Can't signal download at rectangle %v: %v", rect, err)
		}
	}

	for i := 0; i < 128; i++ {
		pos := image.Point{i, i}
		if err := cdw.handleSetPixel(pos, uint8(i%len(can.Palette))); err != nil {
			t.Errorf("Can't set pixel at %v: %v", pos, err)
		}
	}

	for i := 0; i < 128; i++ {
		rect := image.Rectangle{image.Point{i * 64, i * 64}, image.Point{i*64 + 64, i*64 + 64}}
		img := image.NewPaletted(rect, can.Palette)
		if err := cdw.handleSetImage(img); err != nil {
			t.Errorf("Can't set image at %v: %v", rect, err)
		}
	}

	cdw.Close()
}
