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
	"fmt"
	"image"
	"image/png"
	"os"
	"testing"
	"time"
)

func saveCanvasImage(can *canvas, rect image.Rectangle, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("Can't create file %v: %v", filename, err)
	}
	defer file.Close()

	img, err := can.getImageCopy(rect, true, false)
	if err != nil {
		return fmt.Errorf("Can't get image at %v: %v", rect, err)
	}
	//resized := resize.Resize(1000, 1000, img, resize.Lanczos3)
	png.Encode(file, img)

	return nil
}

func Test_newPixelcanvasio(t *testing.T) {
	con, can := newPixelcanvasio()
	defer con.Close()

	cdw, err := can.newCanvasDiskWriter("pixelcanvas.io")
	if err != nil {
		t.Errorf("Can't create canvas disk writer: %v", err)
	}
	defer cdw.Close()

	rect := image.Rect(-960, -450, 960, 450)
	err = cdw.setListeningRects([]image.Rectangle{rect})
	if err != nil {
		t.Errorf("Can't set listening rectangle: %v", err)
	}

	// Stupid way of polling the canvas to check if everything is downloaded
	for valid := can.isValid(rect); valid == false; valid = can.isValid(rect) {
		time.Sleep(1 * time.Second)
	}

	if err := saveCanvasImage(can, rect, "./pixelcanvas-test.png"); err != nil {
		t.Errorf("Can't save image to disk: %v", err)
	}
}
