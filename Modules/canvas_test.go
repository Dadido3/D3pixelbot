package main

import (
	"image"
	"testing"
)

func Test_newCanvas(t *testing.T) {
	can := newCanvas(pixelSize{64, 64}, pixelcanvasioPalette)

	cdw, err := can.newCanvasDiskWriter("Test")
	if err != nil {
		t.Errorf("Can't create canvas disk writer: %v", err)
	}

	can.subscribeListener(cdw)

	for i := 0; i < 128; i++ {
		rect := image.Rectangle{image.Point{i * 64, i * 64}, image.Point{i*64 + 64, i*64 + 64}}
		if err := can.invalidateRect(rect); err != nil {
			t.Errorf("Can't invalidate rectangle %v: %v", rect, err)
		}
	}

	for i := 0; i < 128; i++ {
		pos := image.Point{i, i}
		if err := can.setPixelIndex(pos, uint8(i%len(can.Palette))); err != nil {
			t.Errorf("Can't set pixel at %v: %v", pos, err)
		}
	}

	for i := 0; i < 128; i++ {
		rect := image.Rectangle{image.Point{i * 64, i * 64}, image.Point{i*64 + 64, i*64 + 64}}
		if err := can.setImage(rect); err != nil {
			t.Errorf("Can't set image at %v: %v", rect, err)
		}
	}

	can.unsubscribeListener(cdw)
	cdw.Close()

	can.Close()
}
