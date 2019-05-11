package main

import (
	"image"
	"image/color/palette"
	"testing"
)

func Test_canvas_newCanvasDiskWriter(t *testing.T) {
	can := &canvas{
		Palette:   palette.Plan9,
		ChunkSize: pixelSize{64, 64},
	}

	cdw, err := can.newCanvasDiskWriter("Test")
	if err != nil {
		t.Errorf("Can't create canvas disk writer: %v", err)
	}

	for i := 0; i < 128; i++ {
		rect := image.Rectangle{image.Point{i * 64, i * 64}, image.Point{i*64 + 64, i*64 + 64}}
		if err := cdw.handleInvalidateRect(rect); err != nil {
			t.Errorf("Can't invalidate rectangle %v: %v", rect, err)
		}
	}

	for i := 0; i < 128; i++ {
		pos := image.Point{i, i}
		if err := cdw.handleSetPixel(pos, 0); err != nil {
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
